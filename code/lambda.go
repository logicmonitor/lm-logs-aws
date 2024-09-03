package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/logicmonitor/lm-data-sdk-go/api/logs"
	"github.com/logicmonitor/lm-data-sdk-go/model"
	"github.com/logicmonitor/lm-data-sdk-go/utils"
)

var awsRegion, scrubRegex, useSecretManager, accessID, accessKey, bearerToken, companyName string
var ingestTimeout int
var debug bool
var sessionNew *session.Session
var s3Manager *s3.S3

func SendLogs(logInput []model.LogInput, lmLog *logs.LMLogIngest) {
	if len(logInput) == 0 {
		return
	}

	ingestResponse, err := lmLog.SendLogs(context.Background(), logInput)
	if err != nil {
		fmt.Println("Error in sending log to LM : ", err)
	}

	handleFatalError("Request failed", err)

	if debug || !ingestResponse.Success {
		json, _ := json.Marshal(ingestResponse)
		log.Printf("Response: %s\n", string(json))
		log.Println(string(json))
	}
}

func ScrubLogsWithRegex(lmBatch []model.LogInput) {
	if scrubRegex != "" {
		reg := regexp.MustCompile(scrubRegex)
		for _, event := range lmBatch {
			log.Printf("%s", fmt.Sprintf("%s", event.Message))
			event.Message = reg.ReplaceAllString(fmt.Sprintf("%s", event.Message), "")
			log.Printf("%s", fmt.Sprintf("%s", event.Message))
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

func ExtractLogs(data interface{}) []model.LogInput {
	logs := []model.LogInput{}
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
	sessionNew = session.Must(session.NewSession())
	s3Manager = s3.New(sessionNew)
	log := ExtractLogs(request)
	ScrubLogsWithRegex(log)

	client := Client()
	auth := utils.AuthParams{AccessID: accessID,
		AccessKey:            accessKey,
		BearerToken:          bearerToken}

	options := []logs.Option{
		logs.WithLogBatchingDisabled(),
		logs.WithAuthentication(auth),
		logs.WithUserAgent("lm-logs-aws"),
		logs.WithHTTPClient(client),
	}

	lmLog, err := logs.NewLMLogIngest(context.Background(), options...)
	if err != nil {
		fmt.Println("Error in initilaizing log ingest ", err)
		return
	}
	SendLogs(log, lmLog)

}

func Client() *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: false, MinVersion: tls.VersionTLS12}
	clientTransport := (http.RoundTripper)(transport)
	return &http.Client{Transport: clientTransport, Timeout: time.Duration(ingestTimeout) * time.Second}
}

func main() {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: false}
	lambda.Start(handler)
}
