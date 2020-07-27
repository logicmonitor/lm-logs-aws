package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

type GetContentFromS3Bucket func(string, string) string

type ResourceId struct {
	Arn string `json:"system.aws.arn"`
}

// Event expected by the HTTP ingestion API
type LMEvent struct {
	Message    string     `json:"msg"`
	Timestamp  time.Time  `json:"timestamp"`
	ResourceId ResourceId `json:"_lm.resourceId"`
}

// Lambda response struct
type Response struct {
	Ok      bool
	Message string
}

type Credentials struct {
	AccessID  string
	AccessKey string
}

type LMv1Token struct {
	AccessID  string
	Signature string
	Epoch     time.Time
}

func (t *LMv1Token) String() string {
	builder := strings.Builder{}
	append := func(s string) {
		if _, err := builder.WriteString(s); err != nil {
			panic(err)
		}
	}
	append("LMv1 ")
	append(t.AccessID)
	append(":")
	append(t.Signature)
	append(":")
	append(strconv.FormatInt(t.Epoch.UnixNano()/1000000, 10))

	return builder.String()
}

var lmHost, lambdaName, awsRegion string
var accessID, accessKey string
var keepTimestamp bool
var debug bool

func getContentsFromS3Bucket(bucketName string, fileName string) string {

	sesssion := session.Must(session.NewSession())
	s3Manager := s3.New(sesssion)
	s3ObjectOutput, err := s3Manager.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(fileName),
	})

	handleFatalError("could not get s3 logs object", err)

	return readCloserToString(s3ObjectOutput.Body)
}

func readCloserToString(body io.ReadCloser) string {
	buf := new(bytes.Buffer)
	buf.ReadFrom(body)
	return buf.String()
}

func generateLMv1Token(credentials *Credentials, method, resourcePath string, b []byte, epochMaybe string) *LMv1Token {

	var epoch string
	var epochTime time.Time

	if epochMaybe == "" {
		epochTime = time.Now()
		epoch = strconv.FormatInt(epochTime.UnixNano()/1000000, 10)
	} else {

		i, _ := strconv.ParseInt(epochMaybe, 10, 64)
		epochTime := time.Unix(i, 0)
		epoch = strconv.FormatInt(epochTime.UnixNano()/1000000, 10)

	}

	methodUpper := strings.ToUpper(method)

	h := hmac.New(sha256.New, []byte(credentials.AccessKey))

	writeOrPanic := func(bs []byte) {
		if _, err := h.Write(bs); err != nil {
			panic(err)
		}
	}
	writeOrPanic([]byte(methodUpper))
	writeOrPanic([]byte(epoch))
	writeOrPanic(b)
	writeOrPanic([]byte(resourcePath))

	hash := h.Sum(nil)
	hexString := hex.EncodeToString(hash)
	signature := base64.StdEncoding.EncodeToString([]byte(hexString))
	return &LMv1Token{
		AccessID:  credentials.AccessID,
		Signature: signature,
		Epoch:     epochTime,
	}
}

func handleFatalError(errStr string, err error) {
	if err != nil {
		log.Fatalf("%s: %s", errStr, err)
	}
}

func postBatch(lmBatch []*LMEvent) error {

	url := lmHost + "/rest/log/ingest"

	data, err := json.Marshal(lmBatch)
	handleFatalError("failed to marshal JSON payload", err)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))

	if debug {
		fmt.Printf("Sending request to %s with payload: %s\n", url, string(data))
	}

	lMv1Token := generateLMv1Token(
		&Credentials{
			AccessID:  accessID,
			AccessKey: accessKey,
		},
		"POST",
		"/log/ingest",
		data,
		"")

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf(lMv1Token.String()))

	var resp *http.Response
	resp, err = http.DefaultClient.Do(req)
	handleFatalError("request failed", err)

	if debug {
		fmt.Printf("Response body :%s\n", readCloserToString(resp.Body))
		fmt.Printf("Response status code :%d\n", resp.StatusCode)
	}
	_ = resp.Body.Close()

	if resp.StatusCode != 202 {
		log.Fatalf("Ingest service did not accepted the message")
	}

	lmBatch = make([]*LMEvent, 0)
	return nil
}

// Lambda handler
func handler(request interface{}) (Response, error) {

	lmBatch := processLogs(request)

	err := postBatch(lmBatch)
	if err != nil {
		fmt.Println(err)
		return Response{
			Ok:      false,
			Message: fmt.Sprintf("Failed to post batch: %s", err),
		}, err
	}

	return Response{
		Ok:      true,
		Message: fmt.Sprintf("%d events sent", len(lmBatch)),
	}, nil
}

func getSecretValue(secretArn string) string {
	sesssion := session.Must(session.NewSession())
	secManager := secretsmanager.New(sesssion)
	secretValueOutput, err := secManager.GetSecretValue(&secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretArn),
	})

	handleFatalError("error in extracting secret value", err)

	return *secretValueOutput.SecretString
}

func parseELBlogs(request events.S3Event, getContentsFromS3Bucket GetContentFromS3Bucket) []*LMEvent {
	bucketName := request.Records[0].S3.Bucket.Name
	key := request.Records[0].S3.Object.Key
	content := getContentsFromS3Bucket(bucketName, key)

	keySplit := strings.Split(key, "_")
	firstPart := strings.Split(keySplit[0], "/")

	accountId := firstPart[1]
	region := firstPart[3]
	name := keySplit[3]
	elbName := strings.ReplaceAll(name, ".", "/")
	allMessages := strings.Split(content, "\n")

	arn := fmt.Sprintf("arn:aws:elasticloadbalancing:%s:%s:loadbalancer/%s", region, accountId, elbName)

	lmBatch := make([]*LMEvent, 0)

	for _, message := range allMessages {

		lmEv := &LMEvent{
			Message:    message,
			ResourceId: ResourceId{arn},
			Timestamp:  request.Records[0].EventTime,
		}

		lmBatch = append(lmBatch, lmEv)
	}

	return lmBatch
}

func parseS3logs(request events.S3Event, getContentsFromS3Bucket GetContentFromS3Bucket) []*LMEvent {
	bucketName := request.Records[0].S3.Bucket.Name
	fileName := request.Records[0].S3.Object.Key

	content := getContentsFromS3Bucket(bucketName, fileName)

	originBucketName := strings.Split(content, " ")[1]

	arn := fmt.Sprintf("arn:aws:s3:::%s", originBucketName)

	lmBatch := make([]*LMEvent, 0)

	lmEv := &LMEvent{
		Message:    content,
		ResourceId: ResourceId{arn},
		Timestamp:  request.Records[0].EventTime,
	}

	lmBatch = append(lmBatch, lmEv)

	return lmBatch
}

func parseEventType(requests interface{}) string {

	data := requests.(map[string]interface{})

	if reflect.ValueOf(data).MapKeys()[0].String() == "awslogs" {
		return "cloudwatch"
	}

	if reflect.ValueOf(data).MapKeys()[0].String() == "Records" {

		event := convertToS3Event(requests)

		if strings.Contains(event.Records[0].S3.Object.Key, "elasticloadbalancing") {
			return "elb"
		}
		return "s3"
	}
	log.Fatalf("Could not extract event type")
	return ""
}

func parseCloudWatchLogs(request events.CloudwatchLogsEvent) []*LMEvent {

	lmBatch := make([]*LMEvent, 0)
	d, err := request.AWSLogs.Parse()
	arn := fmt.Sprintf("arn:aws:ec2:%s:%s:instance/%s", awsRegion, d.Owner, d.LogStream)

	handleFatalError("failed to parse cloudwatch event", err)

	for _, event := range d.LogEvents {

		lmEv := &LMEvent{
			Message:    event.Message,
			ResourceId: ResourceId{arn},
			Timestamp:  time.Unix(0, event.Timestamp*1000000),
		}

		lmBatch = append(lmBatch, lmEv)
	}

	return lmBatch
}

func convertToCloudwatchLogsEvent(m interface{}) events.CloudwatchLogsEvent {
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

func processLogs(data interface{}) []*LMEvent {
	lmBatch := []*LMEvent{}
	source := parseEventType(data)

	if debug {
		fmt.Printf("Event Recieved: %s\n", data)
		fmt.Printf("Source: %s\n", source)
	}

	switch source {
	case "cloudwatch":
		cloudWatchEvent := convertToCloudwatchLogsEvent(data)
		lmBatch = parseCloudWatchLogs(cloudWatchEvent)
	case "s3":
		s3Event := convertToS3Event(data)
		lmBatch = parseS3logs(s3Event, getContentsFromS3Bucket)
	case "elb":
		s3Event := convertToS3Event(data)
		lmBatch = parseELBlogs(s3Event, getContentsFromS3Bucket)
	}
	return lmBatch

}

func main() {
	lambdaName = os.Getenv("AWS_LAMBDA_FUNCTION_NAME")
	awsRegion = os.Getenv("AWS_REGION")

	keepTimestampString := os.Getenv("KEEP_TIMESTAMP")
	keepTimestamp = true
	if keepTimestampString == "false" {
		keepTimestamp = false
	}

	accessKey = getSecretValue(os.Getenv("LM_ACCESS_KEY_ARN"))
	accessID = getSecretValue(os.Getenv("LM_ACCESS_ID_ARN"))

	lmHost = os.Getenv("LM_HOST")
	if lmHost == "" {
		log.Fatalf("Missing LM_HOST env var")
	}

	if os.Getenv("DEBUG") == "true" {
		debug = true
	} else {
		debug = false
	}

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: false}
	fmt.Println("AWS lambda started")
	lambda.Start(handler)

}
