package main

import (
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

type GetContentFromS3Bucket func(string, string) string

func getSecretValue(secretArn string) string {
	session := session.Must(session.NewSession())
	secManager := secretsmanager.New(session)
	secretValueOutput, err := secManager.GetSecretValue(&secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretArn),
	})

	handleFatalError("error in extracting secret value", err)

	return *secretValueOutput.SecretString
}

func getContentsFromS3Bucket(bucketName string, fileName string) string {

	session := session.Must(session.NewSession())
	s3Manager := s3.New(session)
	s3ObjectOutput, err := s3Manager.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(fileName),
	})
	handleFatalError("could not get s3 logs object", err)

	return readCloserToString(s3ObjectOutput.Body)
}

func convertToCloudWatchLogsEvent(m interface{}) events.CloudwatchLogsEvent {
	data, err := json.Marshal(m)
	handleFatalError("failed to marshal s3 event", err)

	var result events.CloudwatchLogsEvent
	err = json.Unmarshal(data, &result)
	handleFatalError("failed to unmarshal s3 event", err)

	return result
}

func convertToS3Event(m interface{}) events.S3Event {
	data, err := json.Marshal(m)
	handleFatalError("failed to marshal s3 event", err)

	var result events.S3Event
	err = json.Unmarshal(data, &result)
	handleFatalError("failed to unmarshal s3 event", err)

	return result
}
