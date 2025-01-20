package main

import "regexp"

var s3Regex = `("bucketName":")(?P<bucketName>[^/][^,][^"]*)|("ARN":")(?P<arn>[^/][^,][^"]*)`
var lambdaRegex = `("functionName":")(?P<functionName>[^/][^,][^"]*)|("resource":")(?P<resource>[^/][^,][^"]*)|("functionVersion":")(?P<functionVersion>[^/][^,][^"]*)`
var awsRegionRegex = `("awsRegion":")(?P<awsRegion>[^/][^,][^"]*)`
var awsEventSourceRegex = `("eventSource":")(?P<eventSource>[^/][^,][^"]*)`
var awsARNRegex = `("arn":")(?P<arn>[^/][^,][^"]*)`
var sqsRegex = `("queueName":")(?P<queueName>[^/][^,][^"]*)|("queueUrl":")(?P<queueUrl>[^/][^,][^"]*)`
var ec2Regex = `("instanceId":\s*")(?P<instanceId>[^/][^"]*)`
var cloudwatchResourceRegex = `("resources":\[")(?P<resources>[^/][^,][^"]*)`
var elbAccountId = `AWSLogs\/(.*)\/elasticloadbalancing`
var elbRegion = `\/elasticloadbalancing\/(.*?)\/`
var rdsInstanceRegex = `/aws/rds/(instance|cluster)/([^/]*)`
var eventSourceRegex = `("eventSource":")([^",]*)`
var kinesisFirehoseRegex = `("deliveryStreamName":"|"deliveryStreamName": "|:deliverystream/)([^/][^,][^"]*)`
var kinesisDataStreamRegex = `("streamName":"|"streamName": "|:stream/)([^/][^,][^"]*)`
var ecsStreamRegex = `("cluster":"|"cluster": "|:cluster/)([^/][^,][^"]*)`

func regexCompile(regexStr string) *regexp.Regexp {
	var regex, _ = regexp.Compile(regexStr)
	return regex
}
