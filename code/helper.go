package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func ExtractEnvironmentVariables() {
	awsRegion = os.Getenv("AWS_REGION")

	useSecretManager = os.Getenv("USE_SECRET_MANAGER")

	if useSecretManager == "true" {
		fmt.Println("Using Secrets Manager to store LM credentials")
		accessKey = getSecretValue(os.Getenv("LM_ACCESS_KEY"))
		if accessKey == "" {
			log.Fatalf("missing LM_ACCESS_KEY env var")
		}

		accessID = getSecretValue(os.Getenv("LM_ACCESS_ID"))
		if accessID == "" {
			log.Fatalf("missing LM_ACCESS_ID env var")
		}
	} else {
		fmt.Println("Using Environmental Variables to store LM credentials")
		accessKey = os.Getenv("LM_ACCESS_KEY")
		if accessKey == "" {
			log.Fatalf("missing LM_ACCESS_KEY env var")
		}

		accessID = os.Getenv("LM_ACCESS_ID")
		if accessID == "" {
			log.Fatalf("missing LM_ACCESS_ID env var")
		}
	}

	lmHost = os.Getenv("LM_HOST")
	companyName = os.Getenv("LM_COMPANY_NAME")
	defaultMetadata := os.Getenv("METADATA")

	metadataArray = strings.Split(defaultMetadata, ",")

	if lmHost == "" && companyName == "" {
		log.Fatalf("missing company name")
	}

	if os.Getenv("DEBUG") == "true" {
		debug = true
	} else {
		debug = false
	}

	scrubRegex = os.Getenv("LM_SCRUB_REGEX")

	logSource = "lm-logs-aws"

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
