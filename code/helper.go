package main

import (
	"bytes"
	"io"
	"log"
	"os"
)

func ExtractEnvironmentVariables() {
	awsRegion = os.Getenv("AWS_REGION")

	accessKey = getSecretValue(os.Getenv("LM_ACCESS_KEY_ARN"))
	if accessKey == "" {
		log.Fatalf("missing LM_ACCESS_KEY_ARN env var")
	}

	accessID = getSecretValue(os.Getenv("LM_ACCESS_ID_ARN"))
	if accessID == "" {
		log.Fatalf("missing LM_ACCESS_ID_ARN env var")
	}

	lmHost = os.Getenv("LM_HOST")
	companyName = os.Getenv("LM_COMPANY_NAME")

	if lmHost == "" && companyName == "" {
		log.Fatalf("missing company name")
	}

	if os.Getenv("DEBUG") == "true" {
		debug = true
	} else {
		debug = false
	}

	scrubRegex = os.Getenv("LM_SCRUB_REGEX")

	logSource = "AWS"

	versionID = "0.0.1"
}

func readCloserToString(body io.ReadCloser) string {
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(body)
	return buf.String()
}

func handleFatalError(errStr string, err error) {
	if err != nil {
		log.Fatalf("%s: %s", errStr, err)
	}
}
