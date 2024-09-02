package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/logicmonitor/lm-data-sdk-go/model"
	jsonq "github.com/thedevsaddam/gojsonq/v2"
)

var defaultJsonMetadataKeys []string
var addCloudWatchMetadata = false
var metadataArray []string
var lmTenantID string
var resourceType string
var goJsonQ = jsonq.New()

func parseELBlogs(request events.S3Event, getContentsFromS3Bucket GetContentFromS3Bucket) ([]model.LogInput, error) {
	lmBatch := make([]model.LogInput, 0)

	bucketName := request.Records[0].S3.Bucket.Name
	key := request.Records[0].S3.Object.Key
	content := getContentsFromS3Bucket(bucketName, key)

	filetype := http.DetectContentType([]byte(content))
	if filetype == "application/x-gzip" {
		content = decompressGzip(content)
	}

	keySplit := strings.Split(key, "_")

	accountIDMatches := regexCompile(elbAccountId).FindStringSubmatch(keySplit[0])
	if len(accountIDMatches) < 2 {
		return lmBatch, fmt.Errorf("failed to parse accountId for: %s", key)
	}
	accountId := accountIDMatches[1]

	regionMatches := regexCompile(elbRegion).FindStringSubmatch(keySplit[0])
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

		log := model.LogInput{
			Message:    message,
			ResourceID: map[string]interface{}{"system.aws.arn": arn},
			Timestamp:  request.Records[0].EventTime.String(),
			Metadata:   metadataMap,
		}

		lmBatch = append(lmBatch, log)
	}
	return lmBatch, nil
}

func parseS3logs(request events.S3Event, getContentsFromS3Bucket GetContentFromS3Bucket) []model.LogInput {
	var arn string
	bucketName := request.Records[0].S3.Bucket.Name
	fileName := request.Records[0].S3.Object.Key
	lmBatch := make([]model.LogInput, 0)

	content := getContentsFromS3Bucket(bucketName, fileName)

	filetype := http.DetectContentType([]byte(content))

	if filetype != "application/x-gzip" {
		originBucketName := strings.Split(content, " ")[1]
		arn = fmt.Sprintf("arn:aws:s3:::%s", originBucketName)
	} else {
		content = decompressGzip(content)
		arn = fmt.Sprintf("arn:aws:s3:::%s", bucketName)
	}

	metadataMap := extractMetadata(request.Records[0].AWSRegion, arn, request.Records[0].EventSource)
	lmEv := model.LogInput{
		Message:    content,
		ResourceID: map[string]interface{}{"system.aws.arn": arn},
		Timestamp:  request.Records[0].EventTime.String(),
		Metadata:   metadataMap,
	}

	lmBatch = append(lmBatch, lmEv)

	return lmBatch
}

func parseCloudWatchLogs(request events.CloudwatchLogsEvent) []model.LogInput {

	var metadataMap = make(map[string]interface{})
	lmBatch := make([]model.LogInput, 0)
	d, err := request.AWSLogs.Parse()
	var resourceValue string
	var resoureProp = make(map[string]interface{})
	var isEC2NetworkInterface bool = false
	var resourceProperty string = "system.aws.arn"
	if d.LogGroup == "RDSOSMetrics" {
		rdsEnhancedEvent := make(map[string]interface{})
		err := json.Unmarshal([]byte(d.LogEvents[0].Message), &rdsEnhancedEvent)
		handleFatalError("RDSOSMetrics event parsing failed", err)
		rdsInstance := rdsEnhancedEvent["instanceID"]
		resourceValue = fmt.Sprintf("arn:aws:rds:%s:%s:db:%s", awsRegion, d.Owner, rdsInstance)
		resoureProp[resourceProperty] = resourceValue
		metadataMap = extractMetadata(awsRegion, resourceValue, "rds.amazonaws.com")
	} else if strings.Contains(d.LogGroup, "/aws/rds") {
		splitLogGroup := strings.Split(d.LogGroup, "/")
		if splitLogGroup[len(splitLogGroup)-1] == "networkInterface" {
			resourceProperty = "system.aws.networkInterfaceId"
			splitLogStream := strings.Split(d.LogStream, "-")
			resourceValue = splitLogStream[0] + "-" + splitLogStream[1]
			resoureProp[resourceProperty] = resourceValue
			metadataMap = extractMetadata(awsRegion, "", "rds.amazonaws.com")

		} else {
			result := regexCompile(rdsInstanceRegex).FindStringSubmatch(d.LogGroup)
			rdsInstance := result[2]
			resourceValue = fmt.Sprintf("arn:aws:rds:%s:%s:db:%s", awsRegion, d.Owner, rdsInstance)
			resoureProp[resourceProperty] = resourceValue
			metadataMap = extractMetadata(awsRegion, resourceValue, "rds.amazonaws.com")

		}
	} else if d.LogGroup != "/aws/lambda/lm" && strings.Contains(d.LogGroup, "/aws/lambda") {
		re1, _ := regexp.Compile(`aws/lambda/(.*)`)
		result := re1.FindStringSubmatch(d.LogGroup)
		lambdaName := result[1]
		resourceValue = fmt.Sprintf("arn:aws:lambda:%s:%s:function:%s", awsRegion, d.Owner, lambdaName)
		resoureProp[resourceProperty] = resourceValue
		metadataMap = extractMetadata(awsRegion, resourceValue, "lambda.amazonaws.com")
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
		resourceValue = fmt.Sprintf("arn:aws:firehose:%s:%s:deliverystream/%s", awsRegion, d.Owner, splitLogGroup[3])
		resoureProp[resourceProperty] = resourceValue
		metadataMap = extractMetadata(awsRegion, resourceValue, "firehose.amazonaws.com")
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
	} else if strings.Contains(d.LogGroup, "eks") {
		resoureProp["system.aws.accountid"] = d.Owner
		resoureProp["system.cloud.category"] = "AWS/LMAccount"
		metadataMap = extractMetadata(awsRegion, "", "eks.amazonaws.com")
	} else {
		resourceValue = fmt.Sprintf("arn:aws:ec2:%s:%s:instance/%s", awsRegion, d.Owner, d.LogStream)
		resoureProp[resourceProperty] = resourceValue
		metadataMap = extractMetadata(awsRegion, resourceValue, "ec2.amazonaws.com")

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
				metadataMap = extractMetadata(awsRegion, resourceValue, "ec2.amazonaws.com")

			}
			addCustomMetadataFromRawJson(metadataMap, event.Message, defaultJsonMetadataKeys)
			mergeMaps(metadataMap, cloudWatchEventMetadata)

			lmEv := model.LogInput{
				Message:    event.Message,
				ResourceID: resoureProp,
				Timestamp:  time.Unix(0, event.Timestamp*1000000).String(),
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

func parseCloudTrailLogs(data events.CloudwatchLogsData) []model.LogInput {
	lmBatch := make([]model.LogInput, 0)

	cloudWatchMetadata := make(map[string]interface{})
	addCloudWatchEventMetadata(cloudWatchMetadata, &data)
	for _, event := range data.LogEvents {
		metadataMap := extractMetadataForCloudTrail(event.Message)
		mergeMaps(metadataMap, cloudWatchMetadata)
		resoureIDMap := processResourceMapping(event.Message, data.Owner)

		lmEv := model.LogInput{
			Message:    event.Message,
			ResourceID: resoureIDMap,
			Timestamp:  time.Unix(0, event.Timestamp*1000000).String(),
			Metadata:   metadataMap,
		}

		if debug {
			log.Printf("request generated to lm-logs api: %s\n", lmEv)
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
			regionRegexArray := regexCompile(awsRegionRegex).FindStringSubmatch(message)
			awsRegion := regexCompile(awsRegionRegex).SubexpIndex("awsRegion")
			if len(regionRegexArray) > 0 && awsRegion != 0 {
				metadataMap["region"] = fmt.Sprintf(regionRegexArray[awsRegion])
			}
		} else if strings.TrimSpace(str) == "arn" {
			arnRegexArray := regexCompile(awsARNRegex).FindStringSubmatch(message)
			awsARN := regexCompile(awsARNRegex).SubexpIndex("arn")
			if len(arnRegexArray) > 0 && awsARN != 0 {
				metadataMap["arn"] = fmt.Sprintf(arnRegexArray[awsARN])
			}
		}
	}
	eventSourceRegexArray := regexCompile(awsEventSourceRegex).FindStringSubmatch(message)
	eventSourceRegex := regexCompile(awsEventSourceRegex).SubexpIndex("eventSource")
	if len(eventSourceRegexArray) > 0 && eventSourceRegex != 0 {
		metadataMap["_type"] = fmt.Sprintf(eventSourceRegexArray[eventSourceRegex])
	}
	addCustomMetadataFromRawJson(metadataMap, message, defaultJsonMetadataKeys)
	if strings.TrimSpace(lmTenantID) != "" {
		metadataMap["_lm.tenantId"] = lmTenantID
	}
	if strings.TrimSpace(resourceType) != "" {
		metadataMap["resourceType"] = resourceType
	}
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
	if strings.TrimSpace(lmTenantID) != "" {
		metadataMap["_lm.tenantId"] = lmTenantID
	}
	if strings.TrimSpace(resourceType) != "" {
		metadataMap["_resourceType"] = resourceType
	}

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

func processResourceMapping(message string, accountId string) map[string]interface{} {
	eventSourceArray := regexCompile(eventSourceRegex).FindStringSubmatch(message)
	eventSource := eventSourceArray[2]
	var lambdaMapping string
	accountLevelLog := true
	var resoureIDMap = make(map[string]interface{})
	var resourceProperty string = "system.aws.arn"

	if strings.Contains(eventSource, "firehose") {
		deliveryStreamArray := regexCompile(kinesisFirehoseRegex).FindStringSubmatch(message)
		if len(deliveryStreamArray) > 2 {
			resoureIDMap[resourceProperty] = fmt.Sprintf("arn:aws:firehose:%s:%s:deliverystream/%s", awsRegion, accountId, deliveryStreamArray[2])
			accountLevelLog = false
		}
	} else if strings.Contains(eventSource, "kinesis") {
		dataStreamArray := regexCompile(kinesisDataStreamRegex).FindStringSubmatch(message)
		if len(dataStreamArray) > 2 {
			resoureIDMap[resourceProperty] = fmt.Sprintf("arn:aws:kinesis:%s:%s:stream/%s", awsRegion, accountId, dataStreamArray[2])
			accountLevelLog = false
		}
	} else if strings.Contains(eventSource, "ecs") {
		ecsStreamArray := regexCompile(ecsStreamRegex).FindStringSubmatch(message)
		if len(ecsStreamArray) > 2 {
			resoureIDMap[resourceProperty] = fmt.Sprintf("arn:aws:ecs:%s:%s:cluster/%s", awsRegion, accountId, ecsStreamArray[2])
			accountLevelLog = false
		}
	} else if strings.Contains(eventSource, "s3") {
		s3RegexArray := regexCompile(s3Regex).FindStringSubmatch(message)

		s3Arn := regexCompile(s3Regex).SubexpIndex("arn")
		s3Bucket := regexCompile(s3Regex).SubexpIndex("bucketName")

		if len(s3RegexArray) > 0 {
			if s3RegexArray[s3Bucket] != "" {
				resoureIDMap[resourceProperty] = fmt.Sprintf("arn:aws:s3:::%s", s3RegexArray[s3Bucket])
				accountLevelLog = false
			} else if s3RegexArray[s3Arn] != "" {
				resoureIDMap[resourceProperty] = fmt.Sprintf(s3RegexArray[s3Arn])
				accountLevelLog = false
			}
		}

	} else if strings.Contains(eventSource, "lambda") {

		lambdaRegexArray := regexCompile(lambdaRegex).FindStringSubmatch(message)
		if len(lambdaRegexArray) > 0 {
			lambdaFunctionName := regexCompile(lambdaRegex).SubexpIndex("functionName")
			lambdaResourceName := regexCompile(lambdaRegex).SubexpIndex("resource")
			lambdaFunctionWithVersion := regexCompile(lambdaRegex).SubexpIndex("functionVersion")

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
					resoureIDMap[resourceProperty] = lambdaMapping

				} else if !strings.Contains(lambdaMapping, "arn:aws:lambda") {
					resoureIDMap[resourceProperty] = fmt.Sprintf("arn:aws:lambda:%s:%s:function:%s", awsRegion, accountId, lambdaMapping)

				} else if strings.Contains(lambdaMapping, ":$") {
					resoureIDMap[resourceProperty] = strings.Split(lambdaMapping, ":$")[0]
				} else {
					accountLevelLog = true
				}
			}
		}

	} else if strings.Contains(eventSource, "ec2") {
		ec2RegexArray := regexCompile(ec2Regex).FindAllStringSubmatch(message, -1)
		if len(ec2RegexArray) == 1 {
			resoureIDMap[resourceProperty] = fmt.Sprintf("arn:aws:ec2:%s:%s:instance/%s", awsRegion, accountId, ec2RegexArray[0][2])
			accountLevelLog = false
		}

	} else if strings.Contains(eventSource, "sqs") {
		sqsRegexArray := regexCompile(sqsRegex).FindStringSubmatch(message)

		sqsName := regexCompile(sqsRegex).SubexpIndex("queueName")
		sqsUrl := regexCompile(sqsRegex).SubexpIndex("queueUrl")

		if len(sqsRegexArray) > 0 {
			if sqsRegexArray[sqsName] != "" {
				resoureIDMap[resourceProperty] = fmt.Sprintf("arn:aws:sqs:%s:%s:%s", awsRegion, accountId, sqsRegexArray[sqsName])
				accountLevelLog = false
			} else if sqsRegexArray[sqsUrl] != "" {
				subStr := strings.Split(sqsRegexArray[sqsUrl], "/")
				resoureIDMap[resourceProperty] = fmt.Sprintf("arn:aws:sqs:%s:%s:%s", awsRegion, accountId, subStr[len(subStr)-1])
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

func parseCloudWatchEvents(request events.CloudWatchEvent) []model.LogInput {
	lmBatch := make([]model.LogInput, 0)
	var resoureIDMap = make(map[string]interface{})
	var metadataMap map[string]interface{}
	var event string
	if strings.EqualFold(request.DetailType, "AWS API Call via CloudTrail") {
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

		cloudwatchResourceRegexArray := regexCompile(cloudwatchResourceRegex).FindStringSubmatch(event)
		cloudwatchResource := regexCompile(cloudwatchResourceRegex).SubexpIndex("resources")
		if len(cloudwatchResourceRegexArray) > 0 && cloudwatchResource != 0 {
			resoureIDMap["system.aws.arn"] = fmt.Sprintf(cloudwatchResourceRegexArray[cloudwatchResource])
		}
		metadataMap = extractMetadata(request.Region, fmt.Sprintf(cloudwatchResourceRegexArray[cloudwatchResource]), request.Source)

	}

	lmEv := model.LogInput{
		Message:    event,
		ResourceID: resoureIDMap,
		Timestamp:  request.Time.String(),
		Metadata:   metadataMap,
	}
	lmBatch = append(lmBatch, lmEv)

	if debug {
		log.Printf("request generated to lm-logs api: %s\n", lmEv)
	}

	return lmBatch

}
