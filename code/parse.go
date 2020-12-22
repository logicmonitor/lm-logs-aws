package main

import (
	"encoding/json"
	"fmt"
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
	bucketName := request.Records[0].S3.Bucket.Name
	fileName := request.Records[0].S3.Object.Key

	content := getContentsFromS3Bucket(bucketName, fileName)

	originBucketName := strings.Split(content, " ")[1]

	arn := fmt.Sprintf("arn:aws:s3:::%s", originBucketName)

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
	var arn string

	if d.LogGroup == "RDSOSMetrics" {
		rdsEnhancedEvent := make(map[string]interface{})
		err := json.Unmarshal([]byte(d.LogEvents[0].Message), &rdsEnhancedEvent)
		handleFatalError("RDSOSMetrics event parsing failed", err)
		rdsInstance := rdsEnhancedEvent["instanceID"]
		arn = fmt.Sprintf("arn:aws:rds:%s:%s:db:%s", awsRegion, d.Owner, rdsInstance)

	} else if strings.Contains(d.LogGroup, "/aws/rds") {
		re1, _ := regexp.Compile(`/aws/rds/(instance|cluster)/([^/]*)`)
		result := re1.FindStringSubmatch(d.LogGroup)
		rdsInstance := result[2]
		arn = fmt.Sprintf("arn:aws:rds:%s:%s:db:%s", awsRegion, d.Owner, rdsInstance)
	} else if d.LogGroup != "/aws/lambda/lm" && strings.Contains(d.LogGroup, "/aws/lambda") {
		re1, _ := regexp.Compile(`aws/lambda/(.*)`)
		result := re1.FindStringSubmatch(d.LogGroup)
		lambdaName := result[1]
		arn = fmt.Sprintf("arn:aws:lambda:%s:%s:function:%s", awsRegion, d.Owner, lambdaName)
	} else {
		arn = fmt.Sprintf("arn:aws:ec2:%s:%s:instance/%s", awsRegion, d.Owner, d.LogStream)
	}

	handleFatalError("failed to parse cloudwatch event", err)

	for _, event := range d.LogEvents {

		if strings.TrimSpace(event.Message) != "" {
			lmEv := ingest.Log{
				Message:    event.Message,
				ResourceID: map[string]string{"system.aws.arn": arn},
				Timestamp:  time.Unix(0, event.Timestamp*1000000),
			}
			lmBatch = append(lmBatch, lmEv)
		}
	}

	return lmBatch
}
