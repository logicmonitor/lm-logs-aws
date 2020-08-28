package main

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseEventTypeCloudWatch(t *testing.T) {

		data := map[string]interface{}{"data": "H4sIAAAAAAAAAE1QyW7CMBT8FctnDH7enVtUlkurHuBSIVQ5xIVI2Wo7RQjx7zWtKvU6M+/NcsOdj9Gd/O46elzgZbkr319W2225WeEZHi69DxkGq0EyIaQ0OsPtcNqEYRozs3CXuGhdV9Vu0Xa/3DYF77pMNoQ6drQuX1e1rhm3kBVxquIxNGNqhn7dtMmHiIs9bjvyMYSLC7UP5NNNLpHn/OufmCyXopSvbPNGt2t8+PFaffk+Pe5vuKmzJZeK56SSKauMklpKyoAzQ3VGIPcwihnBGAUpKAWhuVFcGZpzpSZvkVyXa4G02nKrFAdgs7+N8vtyOiFGEaiC6oJr1IwEKAFBiTIEFIrxXO+BWmCHAj0Nfe+Pj+To2A7R16i6ImHnjMu54XPgCo1DSEgwrRXaj3m1KZ0P+H64fwN9VdQGmAEAAA=="}
		event := map[string]interface{}{
			"awslogs": data,
		}

		assert.Equal(t, "cloudwatch", ParseEventType(event))
}

func TestParseEventTypeELB(t *testing.T) {

	eventJson := `{
    "Records": [
        {
            "awsRegion": "us-west-1",
            "eventName": "ObjectCreated:CompleteMultipartUpload",
            "eventSource": "aws:s3",
            "eventTime": "2020-08-24T13:01:16.519Z",
            "eventVersion": "2.1",
            "requestParameters": {
                "sourceIPAddress": "2600:1f1c:770:4e00:8d7a:1e08:be28:f21c"
            },
            "responseElements": {
                "x-amz-id-2": "RWIPFmzXaBw8Ns0Ivm89wtI5yEJAj/NVt6EWEgaYlz8m39K0xLXK5OZ7Xq60ELgJiZGPFCncm8GOi316i4YU7XLq8Izhba32",
                "x-amz-request-id": "6EE763CD70885405"
            },
            "s3": {
                "bucket": {
                    "arn": "arn:aws:s3:::elb-logs-to-lambda",
                    "name": "elb-logs-to-lambda",
                    "ownerIdentity": {
                        "principalId": "ATRBEOLXIUBLG"
                    }
                },
                "configurationId": "N2FkMzIzZGMtNzFiZC00YmQ1LWJmODctZGM3NzNkYjE2NTYy",
                "object": {
                    "eTag": "b115fb720ee4772071c1df3a04272574-1",
                    "key": "AWSLogs/197152445587/elasticloadbalancing/us-west-1/2020/08/24/197152445587_elasticloadbalancing_us-west-1_elb-lambda-ahsan_20200824T1300Z_52.52.117.168_44ffpjs8.log",
                    "sequencer": "005F43BA1D9A322EE3",
                    "size": 761
                },
                "s3SchemaVersion": "1.0"
            },
            "userIdentity": {
                "principalId": "AWS:AIDAIVHUBDT47DQDIKXZK"
            }
        }
    ]
}`

	var event map[string]interface{}
	json.Unmarshal([]byte(eventJson), &event)

	assert.Equal(t, "elb", ParseEventType(event))
}

func TestParseEventTypeS3(t *testing.T) {

	eventJson := `{
    "Records": [
        {
            "awsRegion": "us-west-1",
            "eventName": "ObjectCreated:Put",
            "eventSource": "aws:s3",
            "eventTime": "2020-08-24T13:07:17.337Z",
            "eventVersion": "2.1",
            "requestParameters": {
                "sourceIPAddress": "212.107.128.230"
            },
            "responseElements": {
                "x-amz-id-2": "3kBzX5K7qfxRsXxo2F2PiaRYOxFp25lMfVvekCEwmLWhNVD/idKXZgNsVKK3AAzvXhiMrYrt6m+EmtGwz4Qzz/LXtxcXoj++",
                "x-amz-request-id": "C339F992C35D6FDB"
            },
            "s3": {
                "bucket": {
                    "arn": "arn:aws:s3:::lm-image-logs",
                    "name": "lm-image-logs",
                    "ownerIdentity": {
                        "principalId": "ATRBEOLXIUBLG"
                    }
                },
                "configurationId": "MjRkMWE5OWMtNDkzMS00ZWMzLWIyYTUtNjg5NmFiYmI4MTA4",
                "object": {
                    "eTag": "89093af23759ce21c4f091482adb241b",
                    "key": "2020-05-23-00-42-05-AEF59C5C81F1AF84.txt",
                    "sequencer": "005F43BB87120B8977",
                    "size": 583
                },
                "s3SchemaVersion": "1.0"
            },
            "userIdentity": {
                "principalId": "AWS:AROAJBCKFEZ5YY3BVJ5VI:muhammad.ahsan@logicmonitor.com"
            }
        }
    ]
}`

	var event map[string]interface{}
	json.Unmarshal([]byte(eventJson), &event)

	assert.Equal(t, "s3", ParseEventType(event))
}