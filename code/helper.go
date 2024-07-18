package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

func ExtractEnvironmentVariables() {
	awsRegion = os.Getenv("AWS_REGION")

	useSecretManager = os.Getenv("USE_SECRET_MANAGER")

	if os.Getenv("DEBUG") == "true" {
		debug = true
	} else {
		debug = false
	}

	if useSecretManager == "true" {
		if debug {
			log.Println("Using Secrets Manager to store LM credentials")
		}
		accessKey = getSecretValue(os.Getenv("LOGICMONITOR_ACCESS_KEY"))

		accessID = getSecretValue(os.Getenv("LOGICMONITOR_ACCESS_ID"))

		bearerToken = getSecretValue((os.Getenv("LOGICMONITOR_BEARER_TOKEN")))

	} else {
		if debug {
			log.Println("Using Environmental Variables to store LM credentials")
		}
		accessKey = os.Getenv("LOGICMONITOR_ACCESS_KEY")

		accessID = os.Getenv("LOGICMONITOR_ACCESS_ID")

		bearerToken = os.Getenv("LOGICMONITOR_BEARER_TOKEN")

	}

	if bearerToken != "" {
		bearerToken = fmt.Sprintf("Bearer %s", bearerToken)
	}

	companyName = os.Getenv("LM_ACCOUNT")
	defaultMetadata := os.Getenv("METADATA")

	metadataArray = strings.Split(defaultMetadata, ",")

	if companyName == "" {
		log.Fatalf("missing company name")
	}

	scrubRegex = os.Getenv("LM_SCRUB_REGEX")

	defaultJsonMetadataKeyString := os.Getenv("JSON_METADATA_KEYS")
	defaultJsonMetadataKeysRaw := strings.Split(defaultJsonMetadataKeyString, ",")
	for _, str := range defaultJsonMetadataKeysRaw {
		defaultJsonMetadataKeys = append(defaultJsonMetadataKeys, strings.Trim(str, " "))
	}
	addCWM, err := strconv.ParseBool(os.Getenv("ADD_CLOUDWATCH_METADATA"))
	if err == nil {
		addCloudWatchMetadata = addCWM
	}

	lmTenantID = os.Getenv("LM_TENANT_IDENTIFIER")
	resourceType = os.Getenv("RESOURCE_TYPE")

	ingestTimeoutStr, exists := os.LookupEnv("LOG_INGEST_TIMEOUT")

	if exists {
		ingestTimeout, err = strconv.Atoi(ingestTimeoutStr)
		if err != nil {
			if debug{
				log.Println("Error converting %s to integer: %v\n", ingestTimeout, err)
			}
			ingestTimeout = 0
		}
	}

	if ingestTimeout == 0 {
		if debug {	
			log.Println("Environmental variable LOG_INGEST_TIMEOUT not set. Using default as 10sec")
		}
		ingestTimeout = 10
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
