package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

type GetContentFromS3Bucket func(string, string) string

func getSecretValue(secretArn string) string {
	var objmap map[string]interface{}

	if len(secretArn) < 1 {
		return ""
	}
	secrets_extension_endpoint := "http://localhost:2773/secretsmanager/get?secretId=" + secretArn

	req, err := http.NewRequest("GET", secrets_extension_endpoint, nil)
	if err != nil {
		log.Println("Error creating HTTP request to secrets extension: ", err)
		return ""
	}
	req.Header.Add("X-Aws-Parameters-Secrets-Token", os.Getenv("AWS_SESSION_TOKEN"))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Error sending HTTP request to secrets extension: ", err)
		return ""
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading HTTP response body:", err)
		return ""
	}

	if err := json.Unmarshal(body, &objmap); err != nil {
		log.Println(err)
	}

	str := fmt.Sprintf("%v", objmap["SecretString"])
	return str
}

func getContentsFromS3Bucket(bucketName string, fileName string) string {

	s3ObjectOutput, err := s3Manager.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(fileName),
	})

	handleFatalError("could not get s3 logs object", err)

	return readCloserToString(s3ObjectOutput.Body)
}

func convertToCloudWatchLogsEvent(m interface{}) events.CloudwatchLogsEvent {
	data, _ := marshalEvent(m)

	var result events.CloudwatchLogsEvent
	var err = json.Unmarshal(data, &result)
	handleFatalError("failed to unmarshal s3 event", err)

	return result
}

func convertToS3Event(m interface{}) events.S3Event {
	data, _ := marshalEvent(m)
	var result events.S3Event
	var err = json.Unmarshal(data, &result)
	handleFatalError("failed to unmarshal s3 event", err)

	return result
}

func convertToCloudWatchEvent(m interface{}) events.CloudWatchEvent {
	data, _ := marshalEvent(m)

	var result events.CloudWatchEvent
	var err = json.Unmarshal(data, &result)
	handleFatalError("failed to unmarshal cloudWatch event", err)
	return result
}

func marshalEvent(m interface{}) ([]byte, error) {
	data, err := json.Marshal(m)
	handleFatalError("failed to marshal s3 event", err)
	return data, err
}
