package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/logicmonitor/lm-logs-sdk-go/ingest"
	jsonq "github.com/thedevsaddam/gojsonq/v2"
)

var s3Regex, _ = regexp.Compile(`("bucketName":")(?P<bucketName>[^/][^,][^"]*)|("ARN":")(?P<arn>[^/][^,][^"]*)`)
var lambdaRegex, _ = regexp.Compile(`("functionName":")(?P<functionName>[^/][^,][^"]*)|("resource":")(?P<resource>[^/][^,][^"]*)|("functionVersion":")(?P<functionVersion>[^/][^,][^"]*)`)
var awsRegionRegex, _ = regexp.Compile(`("awsRegion":")(?P<awsRegion>[^/][^,][^"]*)`)
var awsEventSourceRegex, _ = regexp.Compile(`("eventSource":")(?P<eventSource>[^/][^,][^"]*)`)
var awsARNRegex, _ = regexp.Compile(`("arn":")(?P<arn>[^/][^,][^"]*)`)
var sqsRegex, _ = regexp.Compile(`("queueName":")(?P<queueName>[^/][^,][^"]*)|("queueUrl":")(?P<queueUrl>[^/][^,][^"]*)`)
var ec2Regex, _ = regexp.Compile(`("instanceId":")(?P<instanceId>[^/][^"]*)`)
var cloudwatchResourceRegex, _ = regexp.Compile(`("resources":\[")(?P<resources>[^/][^,][^"]*)`)

var defaultJsonMetadataKeys []string
var addCloudWatchMetadata = false
var metadataArray []string
var goJsonQ = jsonq.New()

func parseELBlogs(request events.S3Event, getContentsFromS3Bucket GetContentFromS3Bucket) ([]ingest.Log, error) {
	lmBatch := make([]ingest.Log, 0)

	bucketName := request.Records[0].S3.Bucket.Name
	key := request.Records[0].S3.Object.Key
	content := getContentsFromS3Bucket(bucketName, key)

	filetype := http.DetectContentType([]byte(content))
	if filetype == "application/x-gzip" {
		content = decompressGzip(content)
	}

	keySplit := strings.Split(key, "_")

	re := regexp.MustCompile(`AWSLogs\/(.*)\/elasticloadbalancing`)
	accountIDMatches := re.FindStringSubmatch(keySplit[0])
	if len(accountIDMatches) < 2 {
		return lmBatch, fmt.Errorf("failed to parse accountId for: %s", key)
	}
	accountId := accountIDMatches[1]

	re = regexp.MustCompile(`\/elasticloadbalancing\/(.*?)\/`)
	regionMatches := re.FindStringSubmatch(keySplit[0])
	if len(regionMatches) < 2 {
		return lmBatch, fmt.Errorf("failed to parse region for: %s", key)
	}
	region := regionMatches[1]

	name := keySplit[3]
	elbName := strings.ReplaceAll(name, ".", "/")
	allMessages := strings.Split(content, "\n")

	arn := fmt.Sprintf("arn:aws:elasticloadbalancing:%s:%s:loadbalancer/%s", region, accountId, elbName)

	metadataMap := extractMetadata(region, arn, "elb.amazonaws.com")
	for _, message := range allMessages {

		log := ingest.Log{
			Message:    message,
			ResourceID: map[string]string{"system.aws.arn": arn},
			Timestamp:  request.Records[0].EventTime,
			Metadata:   metadataMap,
		}

		lmBatch = append(lmBatch, log)
	}
	return lmBatch, nil
}

func parseS3logs(request events.S3Event, getContentsFromS3Bucket GetContentFromS3Bucket) []ingest.Log {
	var arn string
	bucketName := request.Records[0].S3.Bucket.Name
	fileName := request.Records[0].S3.Object.Key

	content := getContentsFromS3Bucket(bucketName, fileName)

	filetype := http.DetectContentType([]byte(content))

	if filetype != "application/x-gzip" {
		originBucketName := strings.Split(content, " ")[1]
		arn = fmt.Sprintf("arn:aws:s3:::%s", originBucketName)
	} else {
		content = decompressGzip(content)
		arn = fmt.Sprintf("arn:aws:s3:::%s", bucketName)
	}

	lmBatch := make([]ingest.Log, 0)

	metadataMap := extractMetadata(request.Records[0].AWSRegion, arn, request.Records[0].EventSource)
	lmEv := ingest.Log{
		Message:    content,
		ResourceID: map[string]string{"system.aws.arn": arn},
		Timestamp:  request.Records[0].EventTime,
		Metadata:   metadataMap,
	}

	lmBatch = append(lmBatch, lmEv)

	return lmBatch
}

func parseCloudWatchLogs(request events.CloudwatchLogsEvent) []ingest.Log {
	var metadataMap = map[string]interface{}{}
	lmBatch := make([]ingest.Log, 0)
	d, err := request.AWSLogs.Parse()
	var resourceValue string
	var resoureProp = make(map[string]string)
	var isEC2NetworkInterface bool = false
	var resourceProperty string = "system.aws.arn"

	if d.LogGroup == "RDSOSMetrics" {
		rdsEnhancedEvent := make(map[string]interface{})
		err := json.Unmarshal([]byte(d.LogEvents[0].Message), &rdsEnhancedEvent)
		handleFatalError("RDSOSMetrics event parsing failed", err)
		rdsInstance := rdsEnhancedEvent["instanceID"]
		resourceValue = fmt.Sprintf("arn:aws:rds:%s:%s:db:%s", awsRegion, d.Owner, rdsInstance)
		resoureProp[resourceProperty] = resourceValue
		metadataMap = extractMetadata(awsRegion, fmt.Sprintf("arn:aws:rds:%s:%s:db:%s", awsRegion, d.Owner, rdsInstance), "rds.amazonaws.com")
	} else if strings.Contains(d.LogGroup, "/aws/rds") {
		splitLogGroup := strings.Split(d.LogGroup, "/")
		if splitLogGroup[len(splitLogGroup)-1] == "networkInterface" {
			resourceProperty = "system.aws.networkInterfaceId"
			splitLogStream := strings.Split(d.LogStream, "-")
			resourceValue = splitLogStream[0] + "-" + splitLogStream[1]
			resoureProp[resourceProperty] = resourceValue
			metadataMap = extractMetadata(awsRegion, "", "rds.amazonaws.com")

		} else {
			re1, _ := regexp.Compile(`/aws/rds/(instance|cluster)/([^/]*)`)
			result := re1.FindStringSubmatch(d.LogGroup)
			rdsInstance := result[2]
			resourceValue = fmt.Sprintf("arn:aws:rds:%s:%s:db:%s", awsRegion, d.Owner, rdsInstance)
			resoureProp[resourceProperty] = resourceValue
			metadataMap = extractMetadata(awsRegion, fmt.Sprintf("arn:aws:rds:%s:%s:db:%s", awsRegion, d.Owner, rdsInstance), "rds.amazonaws.com")

		}
	} else if d.LogGroup != "/aws/lambda/lm" && strings.Contains(d.LogGroup, "/aws/lambda") {
		re1, _ := regexp.Compile(`aws/lambda/(.*)`)
		result := re1.FindStringSubmatch(d.LogGroup)
		lambdaName := result[1]
		resourceValue = fmt.Sprintf("arn:aws:lambda:%s:%s:function:%s", awsRegion, d.Owner, lambdaName)
		resoureProp[resourceProperty] = resourceValue
		metadataMap = extractMetadata(awsRegion, fmt.Sprintf("arn:aws:lambda:%s:%s:function:%s", awsRegion, d.Owner, lambdaName), "lambda.amazonaws.com")
	} else if strings.Contains(d.LogGroup, "/aws/ec2/networkInterface") {
		isEC2NetworkInterface = true
	} else if strings.Contains(d.LogGroup, "/aws/natGateway/networkInterface") {
		resourceProperty = "system.aws.networkInterfaceId"
		splitLogStream := strings.Split(d.LogStream, "-")
		resourceValue = splitLogStream[0] + "-" + splitLogStream[1]
		resoureProp[resourceProperty] = resourceValue
		metadataMap = extractMetadata(awsRegion, "", "natGateway.amazonaws.com")

	} else if strings.Contains(d.LogGroup, "/aws/kinesisfirehose") {
		splitLogGroup := strings.Split(d.LogGroup, "/")
		resourceValue = splitLogGroup[3]
		resoureProp[resourceProperty] = fmt.Sprintf("arn:aws:firehose:%s:%s:deliverystream/%s", awsRegion, d.Owner, resourceValue)
		metadataMap = extractMetadata(awsRegion, fmt.Sprintf("arn:aws:firehose:%s:%s:deliverystream/%s", awsRegion, d.Owner, resourceValue), "firehose.amazonaws.com")
	} else if strings.Contains(d.LogGroup, "/aws/elb/networkInterface") {
		resourceProperty = "system.aws.networkInterfaceId"
		splitLogStream := strings.Split(d.LogStream, "-")
		resourceValue = splitLogStream[0] + "-" + splitLogStream[1]
		resoureProp[resourceProperty] = resourceValue
		metadataMap = extractMetadata(awsRegion, "", "networkInterfaceId.amazonaws.com")

	} else if strings.Contains(d.LogGroup, "/aws/fargate") {
		resoureProp["system.aws.accountid"] = d.Owner
		resoureProp["system.cloud.category"] = "AWS/LMAccount"
		metadataMap = extractMetadata(awsRegion, "", "fargate.amazonaws.com")
	} else if strings.Contains(d.LogGroup, "/aws/cloudtrail") {
		return parseCloudTrailLogs(d)
	} else {
		resourceValue = fmt.Sprintf("arn:aws:ec2:%s:%s:instance/%s", awsRegion, d.Owner, d.LogStream)
		resoureProp[resourceProperty] = resourceValue
		metadataMap = extractMetadata(awsRegion, fmt.Sprintf("arn:aws:ec2:%s:%s:instance/%s", awsRegion, d.Owner, d.LogStream), "ec2.amazonaws.com")

	}

	handleFatalError("failed to parse cloudwatch event", err)

	cloudWatchEventMetadata := make(map[string]interface{})
	addCloudWatchEventMetadata(cloudWatchEventMetadata, &d)

	for _, event := range d.LogEvents {
		if strings.TrimSpace(event.Message) != "" {
			if isEC2NetworkInterface && resourceValue == "" {
				splitEventMessage := strings.Split(event.Message, " ")
				ec2InstanceID := splitEventMessage[0]
				resourceValue = fmt.Sprintf("arn:aws:ec2:%s:%s:instance/%s", awsRegion, d.Owner, ec2InstanceID)
				resoureProp[resourceProperty] = resourceValue
				metadataMap = extractMetadata(awsRegion, fmt.Sprintf("arn:aws:ec2:%s:%s:instance/%s", awsRegion, d.Owner, ec2InstanceID), "ec2.amazonaws.com")

			}
			addCustomMetadataFromRawJson(metadataMap, event.Message, defaultJsonMetadataKeys)
			mergeMaps(metadataMap, cloudWatchEventMetadata)

			lmEv := ingest.Log{
				Message:    event.Message,
				ResourceID: resoureProp,
				Timestamp:  time.Unix(0, event.Timestamp*1000000),
				Metadata:   metadataMap,
			}
			lmBatch = append(lmBatch, lmEv)
		}
	}

	return lmBatch
}

func decompressGzip(content string) string {
	rdata := strings.NewReader(content)
	ioReaderContent, err := gzip.NewReader(rdata)

	if err != nil {
		handleFatalError("error while parsing gzip file", err)
	}

	defer ioReaderContent.Close()

	strContent, _ := ioutil.ReadAll(ioReaderContent)
	return string(strContent)
}

func parseCloudTrailLogs(data events.CloudwatchLogsData) []ingest.Log {
	lmBatch := make([]ingest.Log, 0)

	cloudWatchMetadata := make(map[string]interface{})
	addCloudWatchEventMetadata(cloudWatchMetadata, &data)
	for _, event := range data.LogEvents {
		metadataMap := extractMetadataForCloudTrail(event.Message)
		mergeMaps(metadataMap, cloudWatchMetadata)
		resoureIDMap := processResourceMapping(event.Message, data.Owner)

		lmEv := ingest.Log{
			Message:    event.Message,
			ResourceID: resoureIDMap,
			Timestamp:  time.Unix(0, event.Timestamp*1000000),
			Metadata:   metadataMap,
		}

		if debug {
			fmt.Printf("request generated to lm-logs api: %s\n", lmEv)
		}
		lmBatch = append(lmBatch, lmEv)
	}

	return lmBatch

}

func extractMetadataForCloudTrail(message string) map[string]interface{} {
	var metadataMap = make(map[string]interface{})
	metadataMap["_integration"] = "aws"
	for _, str := range metadataArray {
		if strings.TrimSpace(str) == "awsRegion" {
			regionRegexArray := awsRegionRegex.FindStringSubmatch(message)
			awsRegion := awsRegionRegex.SubexpIndex("awsRegion")
			if len(regionRegexArray) > 0 && awsRegion != 0 {
				metadataMap["region"] = fmt.Sprintf(regionRegexArray[awsRegion])
			}
		} else if strings.TrimSpace(str) == "arn" {
			arnRegexArray := awsARNRegex.FindStringSubmatch(message)
			awsARN := awsARNRegex.SubexpIndex("arn")
			if len(arnRegexArray) > 0 && awsARN != 0 {
				metadataMap["arn"] = fmt.Sprintf(arnRegexArray[awsARN])
			}
		}
	}
	eventSourceRegexArray := awsEventSourceRegex.FindStringSubmatch(message)
	eventSourceRegex := awsEventSourceRegex.SubexpIndex("eventSource")
	if len(eventSourceRegexArray) > 0 && eventSourceRegex != 0 {
		metadataMap["_type"] = fmt.Sprintf(eventSourceRegexArray[eventSourceRegex])
	}
	addCustomMetadataFromRawJson(metadataMap, message, defaultJsonMetadataKeys)
	return metadataMap
}

func extractMetadata(region string, arn string, eventsource string) map[string]interface{} {
	var metadataMap = make(map[string]interface{})
	metadataMap["_integration"] = "aws"
	for _, str := range metadataArray {
		if strings.TrimSpace(str) == "awsRegion" {
			metadataMap["region"] = region

		} else if strings.TrimSpace(str) == "arn" && arn != "" {
			metadataMap["arn"] = arn
		}
	}
	metadataMap["_type"] = eventsource
	return metadataMap
}

func addCustomMetadataFromRawJson(initialMap map[string]interface{}, rawMessage string, jsonKeys []string) {

	if len(jsonKeys) < 1 {
		return
	}
	if !json.Valid([]byte(rawMessage)) {
		return
	}
	jsonQRead := goJsonQ.FromString(rawMessage)
	for _, str := range jsonKeys {
		val := jsonQRead.Find(str)
		jsonQRead.Reset()
		if val != nil {
			initialMap[str] = val
		}
	}
}

func addCloudWatchEventMetadata(initialMap map[string]interface{}, logData *events.CloudwatchLogsData) {
	if !addCloudWatchMetadata {
		return
	}
	if len(logData.LogGroup) > 0 {
		initialMap["logGroup"] = logData.LogGroup
	}
	if len(logData.LogStream) > 0 {
		initialMap["logStream"] = logData.LogStream
	}
}

func processResourceMapping(message string, accountId string) map[string]string {
	eventSourceRegex, _ := regexp.Compile(`("eventSource":")([^",]*)`)
	eventSourceArray := eventSourceRegex.FindStringSubmatch(message)
	eventSource := eventSourceArray[2]
	var lambdaMapping string
	accountLevelLog := true
	var resoureIDMap = make(map[string]string)

	if eventSource == "firehose.amazonaws.com" {
		kinesisFirehoseRegex, _ := regexp.Compile(`("deliveryStreamName":"|"deliveryStreamName": "|:deliverystream/)([^/][^,][^"]*)`)
		deliveryStreamArray := kinesisFirehoseRegex.FindStringSubmatch(message)
		if len(deliveryStreamArray) > 2 {
			resoureIDMap["system.aws.arn"] = fmt.Sprintf("arn:aws:firehose:%s:%s:deliverystream/%s", awsRegion, accountId, deliveryStreamArray[2])
			accountLevelLog = false
		}
	} else if eventSource == "kinesis.amazonaws.com" {
		kinesisDataStreamRegex, _ := regexp.Compile(`("streamName":"|"streamName": "|:stream/)([^/][^,][^"]*)`)
		dataStreamArray := kinesisDataStreamRegex.FindStringSubmatch(message)
		if len(dataStreamArray) > 2 {
			resoureIDMap["system.aws.arn"] = fmt.Sprintf("arn:aws:kinesis:%s:%s:stream/%s", awsRegion, accountId, dataStreamArray[2])
			accountLevelLog = false
		}
	} else if eventSource == "ecs.amazonaws.com" {
		ecsStreamRegex, _ := regexp.Compile(`("cluster":"|"cluster": "|:cluster/)([^/][^,][^"]*)`)
		ecsStreamArray := ecsStreamRegex.FindStringSubmatch(message)
		if len(ecsStreamArray) > 2 {
			resoureIDMap["system.aws.arn"] = fmt.Sprintf("arn:aws:ecs:%s:%s:cluster/%s", awsRegion, accountId, ecsStreamArray[2])
			accountLevelLog = false
		}
	} else if eventSource == "s3.amazonaws.com" {
		s3RegexArray := s3Regex.FindStringSubmatch(message)

		s3Arn := s3Regex.SubexpIndex("arn")
		s3Bucket := s3Regex.SubexpIndex("bucketName")

		if len(s3RegexArray) > 0 {
			if s3RegexArray[s3Bucket] != "" {
				resoureIDMap["system.aws.arn"] = fmt.Sprintf("arn:aws:s3:::%s", s3RegexArray[s3Bucket])
				accountLevelLog = false
			} else if s3RegexArray[s3Arn] != "" {
				resoureIDMap["system.aws.arn"] = fmt.Sprintf(s3RegexArray[s3Arn])
				accountLevelLog = false
			}
		}

	} else if eventSource == "lambda.amazonaws.com" {

		lambdaRegexArray := lambdaRegex.FindStringSubmatch(message)
		if len(lambdaRegexArray) > 0 {
			lambdaFunctionName := lambdaRegex.SubexpIndex("functionName")
			lambdaResourceName := lambdaRegex.SubexpIndex("resource")
			lambdaFunctionWithVersion := lambdaRegex.SubexpIndex("functionVersion")

			functionNameStr := fmt.Sprintf("%v", lambdaRegexArray[lambdaFunctionName])
			resourceNameStr := fmt.Sprintf("%v", lambdaRegexArray[lambdaResourceName])
			lambdaFunctionWithVersionStr := fmt.Sprintf("%v", lambdaRegexArray[lambdaFunctionWithVersion])

			if functionNameStr != "" {
				lambdaMapping = functionNameStr
			} else if resourceNameStr != "" {
				lambdaMapping = resourceNameStr
			} else {
				lambdaMapping = lambdaFunctionWithVersionStr
			}
			if lambdaMapping != "" {
				accountLevelLog = false
				if strings.Contains(lambdaMapping, "arn:aws:lambda") && !strings.Contains(lambdaMapping, ":$") {
					resoureIDMap["system.aws.arn"] = lambdaMapping

				} else if !strings.Contains(lambdaMapping, "arn:aws:lambda") {
					resoureIDMap["system.aws.arn"] = fmt.Sprintf("arn:aws:lambda:%s:%s:function:%s", awsRegion, accountId, lambdaMapping)

				} else if strings.Contains(lambdaMapping, ":$") {
					resoureIDMap["system.aws.arn"] = strings.Split(lambdaMapping, ":$")[0]
				} else {
					accountLevelLog = true
				}
			}
		}

	} else if eventSource == "ec2.amazonaws.com" {
		ec2RegexArray := ec2Regex.FindAllStringSubmatch(message, -1)
		if len(ec2RegexArray) == 1 {
			resoureIDMap["system.aws.arn"] = fmt.Sprintf("arn:aws:ec2:%s:%s:instance/%s", awsRegion, accountId, ec2RegexArray[0][2])
			accountLevelLog = false
		}

	} else if eventSource == "sqs.amazonaws.com" {
		sqsRegexArray := sqsRegex.FindStringSubmatch(message)

		sqsName := sqsRegex.SubexpIndex("queueName")
		sqsUrl := sqsRegex.SubexpIndex("queueUrl")

		if len(sqsRegexArray) > 0 {
			if sqsRegexArray[sqsName] != "" {
				resoureIDMap["system.aws.arn"] = fmt.Sprintf("arn:aws:sqs:%s:%s:%s", awsRegion, accountId, sqsRegexArray[sqsName])
				accountLevelLog = false
			} else if sqsRegexArray[sqsUrl] != "" {
				subStr := strings.Split(sqsRegexArray[sqsUrl], "/")
				resoureIDMap["system.aws.arn"] = fmt.Sprintf("arn:aws:sqs:%s:%s:%s", awsRegion, accountId, subStr[len(subStr)-1])
				accountLevelLog = false
			}
		}

	}

	if accountLevelLog {
		resoureIDMap["system.aws.accountid"] = accountId
		resoureIDMap["system.cloud.category"] = "AWS/LMAccount"
	}
	return resoureIDMap
}

func parseCloudWatchEvents(request events.CloudWatchEvent) []ingest.Log {
	lmBatch := make([]ingest.Log, 0)
	var resoureIDMap = make(map[string]string)
	var metadataMap map[string]interface{}
	var event string
	if request.DetailType == "AWS API Call via CloudTrail" {
		detailStr, err := json.Marshal(&request.Detail)
		if err != nil {
			panic(err)
		}

		event = string(detailStr)
		metadataMap = extractMetadataForCloudTrail(event)
		resoureIDMap = processResourceMapping(event, request.AccountID)

	} else {
		requestStr, err := json.Marshal(&request)
		if err != nil {
			panic(err)
		}
		event = string(requestStr)
		// the other detail-types for cloudwatch events have different json format and hence processing it separately

		cloudwatchResourceRegexArray := cloudwatchResourceRegex.FindStringSubmatch(event)
		cloudwatchResource := cloudwatchResourceRegex.SubexpIndex("resources")
		if len(cloudwatchResourceRegexArray) > 0 && cloudwatchResource != 0 {
			resoureIDMap["system.aws.arn"] = fmt.Sprintf(cloudwatchResourceRegexArray[cloudwatchResource])
		}
		metadataMap = extractMetadata(request.Region, fmt.Sprintf(cloudwatchResourceRegexArray[cloudwatchResource]), request.Source)

	}

	lmEv := ingest.Log{
		Message:    event,
		ResourceID: resoureIDMap,
		Timestamp:  request.Time.Local(),
		Metadata:   metadataMap,
	}
	lmBatch = append(lmBatch, lmEv)

	if debug {
		fmt.Printf("request generated to lm-logs api: %s\n", lmEv)
	}

	return lmBatch

}
