package main

import (
	"crypto/tls"
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/logicmonitor/lm-logs-sdk-go/ingest"
)

var awsRegion, scrubRegex, logSource, versionID, useSecretManager, accessID, accessKey, bearerToken, companyName string
var debug bool

func getCompany() string {
	if companyName != "" {
		return companyName
	}

	return ""
}

func SendLogs(logs []ingest.Log) {

	if len(logs) == 0 {
		return
	}

	lmIngest, err := ingest.NewLogIngester(getCompany(), accessID, accessKey, bearerToken, logSource, versionID)
	if err != nil {
		log.Fatalf("Error while setting up LM Log Ingestion client. Error : %s", err)
	}

	// Send logs to Logic Monitor
	ingestResponse, err := lmIngest.SendLogs(logs)
	handleFatalError("Request failed", err)

	if debug || !ingestResponse.Success {
		json, _ := json.Marshal(ingestResponse)
		log.Printf("Response: %s\n", string(json))
		log.Println(string(json))
	}
}

func ScrubLogsWithRegex(lmBatch []ingest.Log) {
	if scrubRegex != "" {
		reg := regexp.MustCompile(scrubRegex)
		for _, event := range lmBatch {
			log.Print(event.Message)
			event.Message = reg.ReplaceAllString(event.Message, "")
			log.Print(event.Message)
		}
	}
}

func ParseEventType(requests interface{}) string {
	data := requests.(map[string]interface{})

	_, ok := data["awslogs"] //cloudwatch logs
	if ok {
		return "cloudwatch"
	}

	_, ok = data["Records"] //s3 and elb logs
	if ok {
		event := convertToS3Event(requests)
		if strings.Contains(event.Records[0].S3.Object.Key, "elasticloadbalancing") {
			return "elb"
		}
		return "s3"
	}

	_, ok = data["source"] // cloudWatchEvents
	if ok {
		return "cloudwatchEvents"
	}

	log.Fatalf("Could not extract event type")
	return ""
}

func ExtractLogs(data interface{}) []ingest.Log {
	logs := []ingest.Log{}
	var err error
	source := ParseEventType(data)

	if debug {
		json, _ := json.Marshal(data)
		log.Printf("Event Recieved: %s\n", string(json))
		log.Printf("Source: %s\n", source)
	}

	switch source {
	case "cloudwatch":
		cloudWatchEvent := convertToCloudWatchLogsEvent(data)
		logs = parseCloudWatchLogs(cloudWatchEvent)
	case "s3":
		s3Event := convertToS3Event(data)
		logs = parseS3logs(s3Event, getContentsFromS3Bucket)
	case "elb":
		s3Event := convertToS3Event(data)
		logs, err = parseELBlogs(s3Event, getContentsFromS3Bucket)
		if err != nil {
			log.Printf("WARN failed to parse elb logs %s\n", err)
		}
	case "cloudwatchEvents":
		cloudwatchEvents := convertToCloudWatchEvent(data)
		logs = parseCloudWatchEvents(cloudwatchEvents)
	}
	return logs
}

// Lambda handler
func handler(request interface{}) {
	ExtractEnvironmentVariables()

	logs := ExtractLogs(request)
	ScrubLogsWithRegex(logs)
	SendLogs(logs)
}

func main() {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: false}
	lambda.Start(handler)
}
