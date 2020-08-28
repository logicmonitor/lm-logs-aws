package main

import (
	"bytes"
	"io"
	"log"
	"os"
)

func ExtractEnvironmentVariables() {
	lambdaName = os.Getenv("AWS_LAMBDA_FUNCTION_NAME")
	awsRegion = os.Getenv("AWS_REGION")

	accessKey = getSecretValue(os.Getenv("LM_ACCESS_KEY_ARN"))
	if accessKey == "" {
		log.Fatalf("Missing LM_ACCESS_KEY_ARN env var")
	}

	accessID = getSecretValue(os.Getenv("LM_ACCESS_ID_ARN"))
	if accessID == "" {
		log.Fatalf("Missing LM_ACCESS_ID_ARN env var")
	}

	lmHost = os.Getenv("LM_HOST")
	if lmHost == "" {
		log.Fatalf("Missing LM_HOST env var")
	}

	if os.Getenv("DEBUG") == "true" {
		debug = true
	} else {
		debug = false
	}

	scrubRegex = os.Getenv("LM_SCRUB_REGEX")
}

func readCloserToString(body io.ReadCloser) string {
	buf := new(bytes.Buffer)
	buf.ReadFrom(body)
	return buf.String()
}

func handleFatalError(errStr string, err error) {
	if err != nil {
		log.Fatalf("%s: %s", errStr, err)
	}
}
