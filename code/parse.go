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
)

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

	for _, message := range allMessages {

		log := ingest.Log{
			Message:    message,
			ResourceID: map[string]string{"system.aws.arn": arn},
			Timestamp:  request.Records[0].EventTime,
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

	lmEv := ingest.Log{
		Message:    content,
		ResourceID: map[string]string{"system.aws.arn": arn},
		Timestamp:  request.Records[0].EventTime,
	}

	lmBatch = append(lmBatch, lmEv)

	return lmBatch
}

func parseCloudWatchLogs(request events.CloudwatchLogsEvent) []ingest.Log {

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
	} else if strings.Contains(d.LogGroup, "/aws/rds") {
		re1, _ := regexp.Compile(`/aws/rds/(instance|cluster)/([^/]*)`)
		result := re1.FindStringSubmatch(d.LogGroup)
		rdsInstance := result[2]
		resourceValue = fmt.Sprintf("arn:aws:rds:%s:%s:db:%s", awsRegion, d.Owner, rdsInstance)
		resoureProp[resourceProperty] = resourceValue
	} else if d.LogGroup != "/aws/lambda/lm" && strings.Contains(d.LogGroup, "/aws/lambda") {
		re1, _ := regexp.Compile(`aws/lambda/(.*)`)
		result := re1.FindStringSubmatch(d.LogGroup)
		lambdaName := result[1]
		resourceValue = fmt.Sprintf("arn:aws:lambda:%s:%s:function:%s", awsRegion, d.Owner, lambdaName)
		resoureProp[resourceProperty] = resourceValue
	} else if strings.Contains(d.LogGroup, "/aws/ec2/networkInterface") {
		isEC2NetworkInterface = true
	} else if strings.Contains(d.LogGroup, "/aws/natGateway/networkInterface") {
		resourceProperty = "system.aws.networkInterfaceId"
		splitLogStream := strings.Split(d.LogStream, "-")
		resourceValue = splitLogStream[0] + "-" + splitLogStream[1]
		resoureProp[resourceProperty] = resourceValue
	} else if strings.Contains(d.LogGroup, "/aws/kinesisfirehose") {
		splitLogGroup := strings.Split(d.LogGroup, "/")
		resourceValue = splitLogGroup[3]
		resoureProp[resourceProperty] = fmt.Sprintf("arn:aws:firehose:%s:%s:deliverystream/%s", awsRegion, d.Owner, resourceValue)
	} else if strings.Contains(d.LogGroup, "/aws/cloudtrail") {
		return parseCloudTrailLogs(d)
	} else {
		resourceValue = fmt.Sprintf("arn:aws:ec2:%s:%s:instance/%s", awsRegion, d.Owner, d.LogStream)
		resoureProp[resourceProperty] = resourceValue
	}

	handleFatalError("failed to parse cloudwatch event", err)

	for _, event := range d.LogEvents {
		if strings.TrimSpace(event.Message) != "" {
			if isEC2NetworkInterface && resourceValue == "" {
				splitEventMessage := strings.Split(event.Message, " ")
				ec2InstanceID := splitEventMessage[0]
				resourceValue = fmt.Sprintf("arn:aws:ec2:%s:%s:instance/%s", awsRegion, d.Owner, ec2InstanceID)
				resoureProp[resourceProperty] = resourceValue
			}

			lmEv := ingest.Log{
				Message:    event.Message,
				ResourceID: resoureProp,
				Timestamp:  time.Unix(0, event.Timestamp*1000000),
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

	for _, event := range data.LogEvents {
		var resoureIDMap = make(map[string]string)
		eventSourceRegex, _ := regexp.Compile(`("eventSource":")([^",]*)`)
		eventSourceArray := eventSourceRegex.FindStringSubmatch(event.Message)
		eventSource := eventSourceArray[2]

		accountLevelLog := true

		if eventSource == "firehose.amazonaws.com" {
			kinesisFirehoseRegex, _ := regexp.Compile(`("deliveryStreamName":"|"deliveryStreamName": "|:deliverystream/)([^/][^,][^"]*)`)
			deliveryStreamArray := kinesisFirehoseRegex.FindStringSubmatch(event.Message)
			if len(deliveryStreamArray) > 2 {
				resoureIDMap["system.aws.arn"] = fmt.Sprintf("arn:aws:firehose:%s:%s:deliverystream/%s", awsRegion, data.Owner, deliveryStreamArray[2])
				accountLevelLog = false
			}
		} else if eventSource == "kinesis.amazonaws.com" {
			kinesisDataStreamRegex, _ := regexp.Compile(`("streamName":"|"streamName": "|:stream/)([^/][^,][^"]*)`)
			dataStreamArray := kinesisDataStreamRegex.FindStringSubmatch(event.Message)
			if len(dataStreamArray) > 2 {
				resoureIDMap["system.aws.arn"] = fmt.Sprintf("arn:aws:kinesis:%s:%s:stream/%s", awsRegion, data.Owner, dataStreamArray[2])
				accountLevelLog = false
			}
		} else if eventSource == "ecs.amazonaws.com" {
			ecsStreamRegex, _ := regexp.Compile(`("cluster":"|"cluster": "|:cluster/)([^/][^,][^"]*)`)
			ecsStreamArray := ecsStreamRegex.FindStringSubmatch(event.Message)
			if len(ecsStreamArray) > 2 {
				resoureIDMap["system.aws.arn"] = fmt.Sprintf("arn:aws:ecs:%s:%s:cluster/%s", awsRegion, data.Owner, ecsStreamArray[2])
				accountLevelLog = false
			}
		}

		if accountLevelLog {
			resoureIDMap["system.aws.accountid"] = data.Owner
			resoureIDMap["system.cloud.category"] = "AWS/LMAccount"
		}

		lmEv := ingest.Log{
			Message:    event.Message,
			ResourceID: resoureIDMap,
			Timestamp:  time.Unix(0, event.Timestamp*1000000),
		}
		lmBatch = append(lmBatch, lmEv)
	}

	return lmBatch

}
