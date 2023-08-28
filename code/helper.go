package main

import (
	"bytes"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

func ExtractEnvironmentVariables() {
	awsRegion = os.Getenv("AWS_REGION")

	useSecretManager = os.Getenv("USE_SECRET_MANAGER")

	if useSecretManager == "true" {
		if debug {
			log.Println("Using Secrets Manager to store LM credentials")
		}
		accessKey = getSecretValue(os.Getenv("LM_ACCESS_KEY"))

		accessID = getSecretValue(os.Getenv("LM_ACCESS_ID"))

		bearerToken = getSecretValue((os.Getenv("LM_BEARER_TOKEN")))

	} else {
		if debug {
			log.Println("Using Environmental Variables to store LM credentials")
		}
		accessKey = os.Getenv("LM_ACCESS_KEY")

		accessID = os.Getenv("LM_ACCESS_ID")

		bearerToken = os.Getenv("LM_BEARER_TOKEN")

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

	versionID = "1.3.0"

	defaultJsonMetadataKeyString := os.Getenv("JSON_METADATA_KEYS")
	defaultJsonMetadataKeysRaw := strings.Split(defaultJsonMetadataKeyString, ",")
	for _, str := range defaultJsonMetadataKeysRaw {
		defaultJsonMetadataKeys = append(defaultJsonMetadataKeys, strings.Trim(str, " "))
	}
	addCWM, err := strconv.ParseBool(os.Getenv("ADD_CLOUDWATCH_METADATA"))
	if err == nil {
		addCloudWatchMetadata = addCWM
	}
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

func mergeMaps(m1 map[string]interface{}, m2 map[string]interface{}) {
	for k, v := range m2 {
		m1[k] = v
	}
}
