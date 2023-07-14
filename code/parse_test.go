package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/logicmonitor/lm-logs-sdk-go/ingest"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
)

func TestParseELBlogs(t *testing.T) {
	var metadataMap = map[string]string{"_integration": "aws", "_type": "elb.amazonaws.com"}

	t.Run("parse elb log without prefix", func(t *testing.T) {
		message := "2020-05-11T09:24:27.754579Z test 78.82.62.133:64107 172.40.0.85:80 0.00005 0.000852 0.000027 304 304 0 0 \"GET http://test-56808838.eu-west-1.elb.amazonaws.com:80/ HTTP/1.1\" \"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.138 Safari/537.36\" - -"
		fileName := "AWSLogs/123123123123/elasticloadbalancing/us-west-1/2020/06/02/123123123123_elasticloadbalancing_us-west-1_test_20200511T0925Z_34.242.46.46_4jtxqo72.txt"
		time, _ := time.Parse(time.RFC3339, "2020-04-08T15:08:34+02:00")
		record := events.S3EventRecord{
			S3: events.S3Entity{
				Bucket: events.S3Bucket{
					Name: "LogBucket",
				},
				Object: events.S3Object{
					Key: fileName,
				},
			},
			EventTime: time,
		}

		records := make([]events.S3EventRecord, 0)

		records = append(records, record)
		s3Event := events.S3Event{
			Records: records,
		}

		var getContentsFromS3BucketMock = func(bucket string, key string) string {
			assert.Equal(t, "LogBucket", bucket)
			assert.Equal(t, fileName, key)
			return message
		}

		//Execution
		lmEvents, _ := parseELBlogs(s3Event, getContentsFromS3BucketMock)

		//Assertion
		expectedLMEvent := ingest.Log{
			Message:    message,
			Timestamp:  time,
			ResourceID: map[string]string{"system.aws.arn": "arn:aws:elasticloadbalancing:us-west-1:123123123123:loadbalancer/test"},
			Metadata:   metadataMap,
		}

		assert.Equal(t, expectedLMEvent, lmEvents[0])
	})

	t.Run("parse elb log with prefix", func(t *testing.T) {
		message := "2020-05-11T09:24:27.754579Z test 78.82.62.133:64107 172.40.0.85:80 0.00005 0.000852 0.000027 304 304 0 0 \"GET http://test-56808838.eu-west-1.elb.amazonaws.com:80/ HTTP/1.1\" \"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.138 Safari/537.36\" - -"
		fileName := "logs/AWSLogs/123123123123/elasticloadbalancing/us-west-1/2020/06/02/123123123123_elasticloadbalancing_us-west-1_test_20200511T0925Z_34.242.46.46_4jtxqo72.txt"
		time, _ := time.Parse(time.RFC3339, "2020-04-08T15:08:34+02:00")
		record := events.S3EventRecord{
			S3: events.S3Entity{
				Bucket: events.S3Bucket{
					Name: "LogBucket",
				},
				Object: events.S3Object{
					Key: fileName,
				},
			},
			EventTime: time,
		}

		records := make([]events.S3EventRecord, 0)

		records = append(records, record)
		s3Event := events.S3Event{
			Records: records,
		}

		var getContentsFromS3BucketMock = func(bucket string, key string) string {
			assert.Equal(t, "LogBucket", bucket)
			assert.Equal(t, fileName, key)
			return message
		}

		//Execution
		lmEvents, _ := parseELBlogs(s3Event, getContentsFromS3BucketMock)

		//Assertion
		expectedLMEvent := ingest.Log{
			Message:    message,
			Timestamp:  time,
			ResourceID: map[string]string{"system.aws.arn": "arn:aws:elasticloadbalancing:us-west-1:123123123123:loadbalancer/test"},
			Metadata:   metadataMap,
		}

		assert.Equal(t, expectedLMEvent, lmEvents[0])
	})
}

func TestParseS3logs(t *testing.T) {
	// Data preparation
	var metadataMap = map[string]string{"_integration": "aws", "_type": "s3.aws"}

	time, _ := time.Parse(time.RFC3339, "2020-04-08T13:08:34+00:00")
	record := events.S3EventRecord{
		S3: events.S3Entity{
			Bucket: events.S3Bucket{
				Name: "LogBucket",
			},
			Object: events.S3Object{
				Key: "Key",
			},
		},
		EventTime:   time,
		EventSource: "s3.aws",
	}

	records := make([]events.S3EventRecord, 0)

	records = append(records, record)
	s3Event := events.S3Event{
		Records: records,
	}

	var getContentsFromS3BucketMock = func(bucket string, key string) string {
		assert.Equal(t, "LogBucket", bucket)
		assert.Equal(t, "Key", key)
		return "a OriginBucket c"
	}

	//Execution
	lmEvents := parseS3logs(s3Event, getContentsFromS3BucketMock)

	//Assertion

	expectedlmEvent := ingest.Log{
		Message:    "a OriginBucket c",
		Timestamp:  time,
		ResourceID: map[string]string{"system.aws.arn": "arn:aws:s3:::OriginBucket"},
		Metadata:   metadataMap,
	}

	assert.Equal(t, expectedlmEvent, lmEvents[0])
}

func TestParseCloudWatchlogs(t *testing.T) {
	var metadataMap = map[string]string{"_integration": "aws", "_type": "ec2.amazonaws.com"}

	cloudWatchEvent := events.CloudwatchLogsEvent{
		AWSLogs: events.CloudwatchLogsRawData{
			Data: "H4sIAAAAAAAAAE1Q22oCMRD9lZBnU5NMNpd9W1oV2lqQlVoQKdk1amBvZKNSpP/etKXQp4E5M+d2w60bR3t064/B4Rw/FOvifTkry2IxwxPcXzsX0lpKoQEgE8LwtG764yL05yEhU3sdp41tq72dNu0vVsbgbJtATyg7VFBnDIwTFVdVlS7GczXWwQ/R993cN9GFEedb3LTk0IerDXsXyHNi+XdGXp9eVpv7R7EqNnO8+1GZXVwXvz9v2O+TGGSgpGZSCK2l5mkYpoAplSlupDRGcQraUKMkVZQDaA5SMJkcRZ9aiLZNgVimJSS/TFBKJ3/tJPpiCAhpxCCnOgeB/ECY4kRQQgnnCu1PdeOToy3nnO1y9LZc56jsG1/7iPoOuXiiE+S7FPdiG6QYAG3HO/y5+/wC3OrpfYUBAAA=",
		},
	}

	lmEvents := parseCloudWatchLogs(cloudWatchEvent)

	time := time.Unix(0, 1586351314000*1000000)
	expectedLMEvent := ingest.Log{
		Message:    "Apr  8 13:08:34 ip-172-40-0-227 dhclient[2221]: XMT: Solicit on eth0, interval 71330ms.",
		Timestamp:  time,
		ResourceID: map[string]string{"system.aws.arn": "arn:aws:ec2::664833354492:instance/i-01fb3c5139e4b27bb"},
		Metadata:   metadataMap,
	}

	assert.Equal(t, expectedLMEvent, lmEvents[0])
}

func TestRDSLogs(t *testing.T) {
	var metadataMap = map[string]string{"_integration": "aws", "_type": "rds.amazonaws.com"}
	cloudWatchEvent := events.CloudwatchLogsEvent{
		AWSLogs: events.CloudwatchLogsRawData{
			Data: "H4sIAAAAAAAAALVSzY7aMBB+FctSpV0pwMSJQ2KEVNpSLqx6gKqHFVqZZEitTeLUNiC04t07G4T2AarePN+fZ8Z+4y16r2vcXnrkin9bbBcvT8vNZrFa8ojbc4eO4CxL8yRJZJoWguDG1itnjz0xE332E1f5iel80F2Jk0oHvdceR/Gktz7UDv2f5mbaBIe6JdeHZgxE+ePel870wdjuu2kCOs/VM1/rdl/pm+kloA+jZkD4bkhbnrAL78I3bioKTWQGEhIAKHIBcZpDMY2TOC9kAaIASGUihYRM5HI6lXlapELkpKIGgqE1BN3SRLEsMiKnWUpB0X09FC9AwAjyEaRMJCrNFB1+br8q9Vk9SzGVO7X+sVKMlb+xfO2t6QKjSBdMVyv2fgG/Rv/Wafz/Oi1t2zcYULGzswFZzPbHw4HegT3AGD49zhiwX4s1O5gGH/wj01WFVUSgw9ae7sfyUjZYzSjDBJzDOIaY+Yj5S1dSBbcq2KAbKhOQzM8Gcoj18zhije1qmu9DrU/oaKo7MGOVuf2zeSZlkrHXLxEjg2l1GCCZE8Svu+tfg3YiCdoCAAA=",
		},
	}

	lmEvents := parseCloudWatchLogs(cloudWatchEvent)

	time := time.Unix(0, 1596584764000*1000000)
	expectedLMEvent := ingest.Log{
		Message:    "2020-08-04 23:46:04 UTC::@:[5275]:LOG:  checkpoint starting: time",
		Timestamp:  time,
		ResourceID: map[string]string{"system.aws.arn": "arn:aws:rds::664833354492:db:database-1"},
		Metadata:   metadataMap,
	}

	assert.Equal(t, expectedLMEvent, lmEvents[0])
}

func TestRDSEnhancedLogs(t *testing.T) {
	var metadataMap = map[string]string{"_integration": "aws", "_type": "rds.amazonaws.com"}

	cloudWatchEvent := events.CloudwatchLogsEvent{
		AWSLogs: events.CloudwatchLogsRawData{
			Data: "H4sIAAAAAAAAAM1Z227bOBD9lYWeEy3vlPxmbJI2aLK5OOlitykKWaYdIrpFopwb8u87pGRbdtvsDdjyIYA4HA7PHM4c0shLkKumSRbq6qlSwSg4GF+Nv5weTibjd4fBXlA+FKoGsxAsopRyxmIC5qxcvKvLtoKZy4PJ2eRUmVqnTTczMbVKcpiaTffpeCzG14K8/yg+nL0nv4nja3n8qzhmF+DbtNMmrXVldFkc6cyouglGn4KTJJ/Oki7KF6Mas585S/DZhT9cqsJYx5dAz2AXygUSsYgJ5pIgKoUkjJOI45gzwUWMJEUcI4SJxJSgCMuIyIghAGA0JG+SHPLAPBZCYknAE+2tSIHwLzeBKha6UDfB6CY4/X1ycXIT7N0EuoCVRaqOD9zELDHJNGnUPtmavVRN2dYbr7cYcQvXkJw/QQTto2gf8StCR5yPCP7DuS2BK2ANnDCM2soucysQGhE0wtJ5FW3+8Zfz66Z3S6v22uhMPyemWwu5LVrYDj5RiCzs+n793Tw1RuVuGMHwIdGdH7F+s8xuF8uQ2u0bVdstQmETKE2SwYiE0gYxyo1QaAEUOlVd/FcYZWUyG0MiwPOpLlqjOkRlodYY5no5HMyNUsUmQK7ysn7qVj3U2qhpkt7ZaZi7bRfqHCI3R7VSu7bLZjnbtU3auuptaZLeKuvAKJRUtOWln200gpg1z7vYUDIRFUO3q54Ex2mRpKbLg8TQPtZW2YZLppmyR0OZtItnujY2GRJZjzypKodB4JhZwzoIJwxjMmAaIxxTB7PJkqmFTaSLOG3nc6gTG5NHhFnKTNLcNR1jTaZUpYuFPUfr/VzmU72iqm6LoptzpWDKHg0a7OtWTbMyveumbPzmIam68GsSh0sYinns1vXUbQy6WB902ZrNIRfKPJS1PdZPL9YLVGKepF2xK3OLXKHXj5YpzkJHg3l0NMY0pPz1s6O2uTs+60O4SvkwPZ9YLkNsCxiUZnZ81jgTVKot3GRd7xL3Hv0aFArrUNf3eT+2mFtorG7A7Orl4qJVrTpRm6xM5fowZOtwLn8LTy11n1E9a2Dkcupx2gJBURfzUt1PnsEQhSR2HlsQ3IJVGjgk9HVvJ10i17t/N11K/lm60ZvpkpDTt/Kda+iBXmi2csaCse2kCQmxGGZNQiJ30yYhle7Iq9unRqdJduD2+ruHj3fZcG0+YIP9JzZwv3rFRtfqAzoel7PFXx0+xqEU26cvol0eUNhVvuV38tT02YNQ246kEnTIhiiS/tZwe9rZI91pElyDA8u5gius6JuSOHV6XHliipF0opaXbWHOS+0cb4KfbS1P7b3YXWyrewHB3pStwm9C4zDirmJ7lAQu49j18/dhSk4p+yZQYIlvAxWcU/EtnFvwrJgijNhX8AiG1uhKqy5TeB6caHdxWl6XjQMjEI3JEO/Z5Kfe2SKwuyz0ShKrpIbA7l2A1nfZtd0ytUfUVYq9sVemnvp1gNptSiU8duyDIM903hWtI7GDRDCLCN7CBO+1fwuKiJDFO6hAz+TXqIhEyN2I34YF9UIiNkSVPzX32WwIh0Fi24jAgr8GhXknCztUrTD1YTpYmDNOt9mCYy7ct625/xciGWCM3sL446iLvYTFkZ+wsJ+wiEewBjXfPQw8wTWAxfyE9aaK/ThYwk9Y0k9YHsk85gNcPun8QCKET0I/xOWn0gtPlV74pPRDXH5KvfBT6oWfUi/8lPrIJ+kawPJJIgaw/OzEyM9OjPzsxMij180Qlk+Pmw2s2E+BiH162gzu6tjPVoz9bMXYz0sx9lMhYi8VIsJeXoqRVw+u9Y9Y+x/+t3F9fg3g709et82y8CAAAA==",
		},
	}

	logs := parseCloudWatchLogs(cloudWatchEvent)

	time := time.Unix(0, 1596671721000*1000000)
	expectedLMEvent := ingest.Log{
		Message:    "{\"engine\":\"MYSQL\",\"instanceID\":\"database-2\",\"instanceResourceID\":\"db-3AA6AU62HV6KOH2W6IU7IN6I4Q\",\"timestamp\":\"2020-08-05T23:55:21Z\",\"version\":1,\"uptime\":\"00:20:17\",\"numVCPUs\":1,\"cpuUtilization\":{\"guest\":0.0,\"irq\":0.0,\"system\":0.8,\"wait\":0.2,\"idle\":97.3,\"user\":1.6,\"total\":2.7,\"steal\":0.1,\"nice\":0.0},\"loadAverageMinute\":{\"one\":0.0,\"five\":0.0,\"fifteen\":0.0},\"memory\":{\"writeback\":0,\"hugePagesFree\":0,\"hugePagesRsvd\":0,\"hugePagesSurp\":0,\"cached\":435728,\"hugePagesSize\":2048,\"free\":100836,\"hugePagesTotal\":0,\"inactive\":294920,\"pageTables\":3476,\"dirty\":280,\"mapped\":61940,\"active\":524112,\"total\":1019328,\"slab\":42776,\"buffers\":25824},\"tasks\":{\"sleeping\":96,\"zombie\":0,\"running\":0,\"stopped\":0,\"total\":96,\"blocked\":0},\"swap\":{\"cached\":0,\"total\":4095996,\"free\":4095996,\"in\":0.0,\"out\":0.0},\"network\":[{\"interface\":\"eth0\",\"rx\":654.28,\"tx\":2893.35}],\"diskIO\":[{\"writeKbPS\":5.13,\"readIOsPS\":0.17,\"await\":0.71,\"readKbPS\":0.67,\"rrqmPS\":0.0,\"util\":0.04,\"avgQueueLen\":0.0,\"tps\":1.4,\"readKb\":40,\"device\":\"rdsdev\",\"writeKb\":308,\"avgReqSz\":8.29,\"wrqmPS\":0.0,\"writeIOsPS\":1.23},{\"writeKbPS\":27.4,\"readIOsPS\":0.17,\"await\":0.32,\"readKbPS\":0.67,\"rrqmPS\":0.0,\"util\":0.08,\"avgQueueLen\":0.0,\"tps\":2.53,\"readKb\":40,\"device\":\"filesystem\",\"writeKb\":1644,\"avgReqSz\":22.16,\"wrqmPS\":2.27,\"writeIOsPS\":2.37}],\"physicalDeviceIO\":[{\"writeKbPS\":5.13,\"readIOsPS\":1.17,\"await\":0.48,\"readKbPS\":4.67,\"rrqmPS\":0.0,\"util\":0.08,\"avgQueueLen\":0.0,\"tps\":1.67,\"readKb\":280,\"device\":\"xvdg\",\"writeKb\":308,\"avgReqSz\":11.76,\"wrqmPS\":0.68,\"writeIOsPS\":0.5}],\"fileSys\":[{\"used\":379496,\"name\":\"\",\"usedFiles\":210,\"usedFilePercent\":0.02,\"maxFiles\":1310720,\"mountPoint\":\"/rdsdbdata\",\"total\":20496340,\"usedPercent\":1.85},{\"used\":2172928,\"name\":\"\",\"usedFiles\":75334,\"usedFilePercent\":11.5,\"maxFiles\":655360,\"mountPoint\":\"/\",\"total\":10190104,\"usedPercent\":21.32}],\"processList\":[{\"vss\":760392,\"name\":\"OS processes\",\"tgid\":0,\"parentID\":0,\"memoryUsedPc\":3.67,\"cpuUsedPc\":0.02,\"id\":0,\"rss\":37452,\"vmlimit\":0},{\"vss\":2148212,\"name\":\"RDS processes\",\"tgid\":0,\"parentID\":0,\"memoryUsedPc\":26.49,\"cpuUsedPc\":1.47,\"id\":0,\"rss\":270036,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4745,\"rss\":154532,\"vmlimit\":\"unlimited\"},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.02,\"id\":4748,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4749,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4750,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4751,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4752,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.02,\"id\":4753,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4754,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4755,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4756,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4757,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4758,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.15,\"id\":4759,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.02,\"id\":4760,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.02,\"id\":4761,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4762,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.02,\"id\":4763,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"nam
		Timestamp:  time,
		ResourceID: map[string]string{"system.aws.arn": "arn:aws:rds::664833354492:db:database-2"},
		Metadata:   metadataMap,
	}

	assert.Equal(t, expectedLMEvent, logs[0])
}

func TestLambdaLogs(t *testing.T) {
	var metadataMap = map[string]string{"_integration": "aws", "_type": "lambda.amazonaws.com"}

	cloudWatchEvent := events.CloudwatchLogsEvent{
		AWSLogs: events.CloudwatchLogsRawData{
			Data: "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
		},
	}

	logs := parseCloudWatchLogs(cloudWatchEvent)

	time := time.Unix(0, 1598517709043*1000000)
	expectedLMEvent := ingest.Log{
		Message:    "START RequestId: b8cf7efa-a997-4a31-a1ff-881feee2273e Version: $LATEST\n",
		Timestamp:  time,
		ResourceID: map[string]string{"system.aws.arn": "arn:aws:lambda::197152445587:function:observatory-worker"},
		Metadata:   metadataMap,
	}

	assert.Equal(t, expectedLMEvent, logs[0])
}

func TestEC2FlowLogs(t *testing.T) {
	var metadataMap = map[string]string{"_integration": "aws", "_type": "ec2.amazonaws.com"}
	cloudWatchEvent := events.CloudwatchLogsEvent{
		AWSLogs: events.CloudwatchLogsRawData{
			Data: "H4sIAAAAAAAAAL3RUUsjMRAA4L8S8ly3mcnMJPGtaE8OkRPaN5Fju6ayaHfL7mo5xP/ubEW04MPB0XubZIaZfJkXu8l9X97n5Z9ttqf2fLac/b6aLxazi7md2HbX5E6vIQVgJGKOQa8f2/uLrn3aamZa7vpprnDa5GHXdg8/myF367LK72WLocvlRutyU5+4AOuVphBdFOfBn5RVlbeDlvZPq77q6u1Qt82P+lF79Pb0xq73YVNusr3d95s/52YYUy+2vtO2XhxJAgdMCQlJILFEx8BBopCTSCDCTJQCMnrGyJAk6sihVvpQblQBAuJT8szOucnHl2h7fbOEVYCYGaG6W5MP5utfmG9ZhrEASAUiFOINuAI8FUHPZIi8IR4jMRCNSEzmc/xHSEBmdnY2v16aX5f2dfJv2nRc7YHvkP4uHc2qZeNFoyNrgzuyVkIRoFBmcof08bloENWKJsajQ+E/rvVAPRL3Vl2qob+A3r6+AaSOCSloBAAA",
		},
	}

	logs := parseCloudWatchLogs(cloudWatchEvent)

	time := time.Unix(0, 1616399355000*1000000)
	expectedLMEvent := ingest.Log{
		Message:    "i-067b718e521cdf437 197152445587 eni-071fbace220860313 52.119.221.63 10.134.7.224 443 45224 6 18 6689 1616399355 1616399414 ACCEPT OK",
		Timestamp:  time,
		ResourceID: map[string]string{"system.aws.arn": "arn:aws:ec2::197152445587:instance/i-067b718e521cdf437"},
		Metadata:   metadataMap,
	}

	assert.Equal(t, expectedLMEvent, logs[0])
}

func TestNATFlowLogs(t *testing.T) {
	var metadataMap = map[string]string{"_integration": "aws", "_type": "natGateway.amazonaws.com"}
	cloudWatchEvent := events.CloudwatchLogsEvent{
		AWSLogs: events.CloudwatchLogsRawData{
			Data: "H4sIAAAAAAAAAL2Wb2/aMBDGv4qV1zT4zr6z3XeoZWiapk2Cd1M1ZdStopWAknSoqvrdd4FSAdW2F3Mi/og4jn0/7rnz85ytYtMU93HxtInZZXY9WUy+f57O55PZNBtl620VaxmG4IDQWiLvZPhhfT+r148buTMuts24KtpZ0cZt8TSuYrtd1z8/Vm2s74pl3M+et3UsVjI9VuUFMGi6o+KiWC7jppUZzeOPZlmXm7ZcVx/KB3m0yS6/ZUV2s3t6+itWbTfynJW3sohh7YJnSxoMQgiBGdEY0jaANj44MoHlI5dA4LwJ5INxLDu1pfC2xUpClzCcd06W0lqPDv+DLI/qGFcdh6xA52BsDjmwOVxQDl4rMB4VsWZQrEBZGXnb4PBT5qjJ1dX060J9+ZS9jP6Px/XLg2CDMs4DDwXk0wFhjgZOgSw565W1RnCYFUo40DtSSJkjRDxDAnL4igROYZCvnom8Tkf0KrRjvI5lT9VpjrB3HOgJZy/AHQ450oIjAMFy70CYAIgkegg5uJzsmeIEyGjkwfJjUufnhG2PMmwB2TQJkjebvBPUeX7YGLc7hFCZgNQ7EPUOhKTdcBXUr0/ogEC2sQOdqj6hTfhLh8MBBZfQJ0Aewikdy6GtyAY/WIISeoSzBLFEjNKwnRRQ199Aud7LJyQxCOxFXZR7zs07XyqC09bCYEAJLcIf+oFFdkPJLaQwCOcn6lvvlurhoU12SOgR3veD7tWJTYRn/01y8/Ibd82p5VIPAAA=",
		},
	}

	logs := parseCloudWatchLogs(cloudWatchEvent)

	time := time.Unix(0, 1617877079000*1000000)
	expectedLMEvent := ingest.Log{
		Message:    "2 197152445587 eni-16105f5a 10.134.1.163 10.134.5.180 1382 56061 6 1 40 1617877079 1617877138 ACCEPT OK",
		Timestamp:  time,
		ResourceID: map[string]string{"system.aws.networkInterfaceId": "eni-16105f5a"},
		Metadata:   metadataMap,
	}

	assert.Equal(t, expectedLMEvent, logs[0])
}

func TestParseCloudfrontlogs(t *testing.T) {
	var metadataMap = map[string]string{"_integration": "aws", "_type": "cloudFront.amazonaws.com"}

	// Data preparation
	time, _ := time.Parse(time.RFC3339, "2020-04-08T13:08:34+00:00")
	record := events.S3EventRecord{
		S3: events.S3Entity{
			Bucket: events.S3Bucket{
				Name: "CloudfrontLogBucket",
			},
			Object: events.S3Object{
				Key: "Key",
			},
		},
		EventTime:   time,
		EventSource: "cloudFront.amazonaws.com",
	}

	records := make([]events.S3EventRecord, 0)

	records = append(records, record)
	s3Event := events.S3Event{
		Records: records,
	}

	//Creating a gzip file
	logMsg := "Test the Cloudfront logs"

	f, _ := os.Create("file.gz")
	defer os.Remove("file.gz")
	w := gzip.NewWriter(f)
	_, err := w.Write([]byte(logMsg))
	if err != nil {
		fmt.Println("Error in writing log in file")
	}
	w.Close()

	//Reading from gzip file
	f, _ = os.Open("file.gz")
	result := make([]byte, 100)
	_, err = f.Read(result)
	if err != nil {
		fmt.Println("Error in reading from file")
	}
	defer f.Close()

	var getContentsFromS3BucketMock = func(bucket string, key string) string {
		assert.Equal(t, "CloudfrontLogBucket", bucket)
		assert.Equal(t, "Key", key)
		return string(result)
	}

	lmEvents := parseS3logs(s3Event, getContentsFromS3BucketMock)

	expectedlmEvent := ingest.Log{
		Message:    "Test the Cloudfront logs",
		Timestamp:  time,
		ResourceID: map[string]string{"system.aws.arn": "arn:aws:s3:::CloudfrontLogBucket"},
		Metadata:   metadataMap,
	}

	assert.Equal(t, expectedlmEvent, lmEvents[0])
}

func TestParseCloudtrailLogs(t *testing.T) {
	cloudWatchEvent := events.CloudwatchLogsEvent{
		AWSLogs: events.CloudwatchLogsRawData{
			Data: "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
		},
	}

	var metadataMap = map[string]string{"_integration": "aws", "_type": "ec2.amazonaws.com"}

	logs := parseCloudWatchLogs(cloudWatchEvent)

	time := time.Unix(0, 1618555235714*1000000)
	expectedLMEvent := ingest.Log{
		Message:    "{\"eventVersion\":\"1.08\",\"userIdentity\":{\"type\":\"AssumedRole\",\"principalId\":\"AROAS3ZZTSSJZC36CRUZ4:LMAssumeRoleSession\",\"arn\":\"arn:aws:sts::197152445587:assumed-role/LogicMonitor_119/LMAssumeRoleSession\",\"accountId\":\"197152445587\",\"accessKeyId\":\"ASIAS3ZZTSSJVBZZRIN7\",\"sessionContext\":{\"sessionIssuer\":{\"type\":\"Role\",\"principalId\":\"AROAS3ZZTSSJZC36CRUZ4\",\"arn\":\"arn:aws:iam::197152445587:role/LogicMonitor_119\",\"accountId\":\"197152445587\",\"userName\":\"LogicMonitor_119\"},\"webIdFederationData\":{},\"attributes\":{\"mfaAuthenticated\":\"false\",\"creationDate\":\"2021-04-16T06:03:12Z\"}}},\"eventTime\":\"2021-04-16T06:27:05Z\",\"eventSource\":\"ec2.amazonaws.com\",\"eventName\":\"DescribeInstanceStatus\",\"awsRegion\":\"ap-northeast-1\",\"sourceIPAddress\":\"34.221.10.3\",\"userAgent\":\"aws-sdk-java/1.11.918 Linux/4.14.193-149.317.amzn2.x86_64 OpenJDK_64-Bit_Server_VM/11.0.3+7-LTS java/11.0.3 vendor/Amazon.com_Inc.\",\"errorCode\":\"Client.RequestLimitExceeded\",\"errorMessage\":\"Request limit exceeded.\",\"requestParameters\":{\"instancesSet\":{\"items\":[{\"instanceId\":\"i-0d345eec77c8a08b1\"}]},\"filterSet\":{},\"includeAllInstances\":false},\"responseElements\":null,\"requestID\":\"23b2ea29-b7b1-449f-a584-035f485f05cf\",\"eventID\":\"703a3ad3-3d3d-4bd1-8637-3692f850e857\",\"readOnly\":true,\"eventType\":\"AwsApiCall\",\"managementEvent\":true,\"eventCategory\":\"Management\",\"recipientAccountId\":\"197152445587\"}",
		Timestamp:  time,
		ResourceID: map[string]string{"system.aws.arn": "arn:aws:ec2::197152445587:instance/i-0d345eec77c8a08b1"},
		Metadata:   metadataMap,
	}

	assert.Equal(t, expectedLMEvent, logs[0])
}

func TestElbGzipLogs(t *testing.T) {
	var metadataMap = map[string]string{"_integration": "aws", "_type": "elb.amazonaws.com"}
	message := "2020-05-11T09:24:27.754579Z test 78.82.62.133:64107 172.40.0.85:80 0.00005 0.000852 0.000027 304 304 0 0 \"GET http://test-56808838.eu-west-1.elb.amazonaws.com:80/ HTTP/1.1\" \"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.138 Safari/537.36\" - -"
	fileName := "AWSLogs/123123123123/elasticloadbalancing/us-west-1/2020/06/02/123123123123_elasticloadbalancing_us-west-1_test_20200511T0925Z_34.242.46.46_4jtxqo72.gz"
	time, _ := time.Parse(time.RFC3339, "2020-04-08T15:08:34+02:00")
	record := events.S3EventRecord{
		S3: events.S3Entity{
			Bucket: events.S3Bucket{
				Name: "LogBucket",
			},
			Object: events.S3Object{
				Key: fileName,
			},
		},
		EventTime: time,
	}

	records := make([]events.S3EventRecord, 0)

	records = append(records, record)
	s3Event := events.S3Event{
		Records: records,
	}

	//Creating a gzip file
	logMsg := "2020-05-11T09:24:27.754579Z test 78.82.62.133:64107 172.40.0.85:80 0.00005 0.000852 0.000027 304 304 0 0 \"GET http://test-56808838.eu-west-1.elb.amazonaws.com:80/ HTTP/1.1\" \"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.138 Safari/537.36\" - -"

	f, _ := os.Create("file.gz")
	defer os.Remove("file.gz")
	w := gzip.NewWriter(f)
	_, err := w.Write([]byte(logMsg))
	if err != nil {
		fmt.Println("Error in writing log in file")
	}
	w.Close()

	//Reading from gzip file
	f, _ = os.Open("file.gz")
	result := make([]byte, 512)
	_, err = f.Read(result)
	if err != nil {
		fmt.Println("Error in reading from file")
	}
	defer f.Close()

	var getContentsFromS3BucketMock = func(bucket string, key string) string {
		assert.Equal(t, "LogBucket", bucket)
		assert.Equal(t, fileName, key)
		return string(result)
	}

	//Execution
	lmEvents, _ := parseELBlogs(s3Event, getContentsFromS3BucketMock)

	//Assertion
	expectedLMEvent := ingest.Log{
		Message:    message,
		Timestamp:  time,
		ResourceID: map[string]string{"system.aws.arn": "arn:aws:elasticloadbalancing:us-west-1:123123123123:loadbalancer/test"},
		Metadata:   metadataMap,
	}

	assert.Equal(t, expectedLMEvent, lmEvents[0])
}

//Test case for AWS kinesis logs from cloudtrail
func TestParseKinesisFirehoseLogs(t *testing.T) {
	cloudWatchEvent := events.CloudwatchLogsEvent{
		AWSLogs: events.CloudwatchLogsRawData{
			Data: "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
		},
	}

	var metadataMap = map[string]string{"_integration": "aws", "_type": "firehose.amazonaws.com"}

	logs := parseCloudWatchLogs(cloudWatchEvent)

	localTime := time.Local
	timeValue := time.Date(2021, time.April, 26, 11, 15, 15, 228000000, time.Local)
	if strings.Contains(localTime.String(), "UTC") { //Test case is running at system with time.Local as UTC
		timeValue = time.Date(2021, time.April, 26, 5, 45, 15, 228000000, time.Local)
	}

	expectedLMEvent := ingest.Log{
		Message:    "{\"eventVersion\":\"1.08\",\"userIdentity\":{\"type\":\"AssumedRole\",\"principalId\":\"AROAS3ZZTSSJZC36CRUZ4:LMAssumeRoleSession\",\"arn\":\"arn:aws:sts::197152445587:assumed-role/LogicMonitor_119/LMAssumeRoleSession\",\"accountId\":\"197152445587\",\"accessKeyId\":\"ASIAS3ZZTSSJ5UWDCK2J\",\"sessionContext\":{\"sessionIssuer\":{\"type\":\"Role\",\"principalId\":\"AROAS3ZZTSSJZC36CRUZ4\",\"arn\":\"arn:aws:iam::197152445587:role/LogicMonitor_119\",\"accountId\":\"197152445587\",\"userName\":\"LogicMonitor_119\"},\"webIdFederationData\":{},\"attributes\":{\"mfaAuthenticated\":\"false\",\"creationDate\":\"2021-04-26T05:23:11Z\"}}},\"eventTime\":\"2021-04-26T05:30:50Z\",\"eventSource\":\"firehose.amazonaws.com\",\"eventName\":\"DescribeDeliveryStream\",\"awsRegion\":\"ap-northeast-1\",\"sourceIPAddress\":\"34.214.159.46\",\"userAgent\":\"aws-sdk-java/1.11.918 Linux/4.14.193-149.317.amzn2.x86_64 OpenJDK_64-Bit_Server_VM/11.0.3+7-LTS java/11.0.3 vendor/Amazon.com_Inc.\",\"errorCode\":\"ResourceNotFoundException\",\"errorMessage\":\"Firehose firehosedelievery under account 197152445587 not found.\",\"requestParameters\":{\"deliveryStreamName\":\"firehosedelievery\"},\"responseElements\":null,\"requestID\":\"dd1d624c-66ab-9056-841d-300639f2b022\",\"eventID\":\"f25e9886-5059-4b07-90d1-d29a965c0a0f\",\"readOnly\":true,\"eventType\":\"AwsApiCall\",\"managementEvent\":true,\"eventCategory\":\"Management\",\"recipientAccountId\":\"197152445587\"}",
		Timestamp:  timeValue,
		ResourceID: map[string]string{"system.aws.arn": "arn:aws:firehose::197152445587:deliverystream/firehosedelievery"},
		Metadata:   metadataMap,
	}

	assert.Equal(t, expectedLMEvent, logs[4])
}

func TestKinesisFirehoseErrorLog(t *testing.T) {
	var metadataMap = map[string]string{"_integration": "aws", "_type": "firehose.amazonaws.com"}
	cloudWatchEvent := events.CloudwatchLogsEvent{
		AWSLogs: events.CloudwatchLogsRawData{
			Data: "H4sIAAAAAAAAADWPQW7CMBBFrxJ5jZTYHtsxu0gNbNpVsqtQ5RITrJI48piiCnH3Di0s/fxm5v8rmzyiG33/s3i2Zi9N33y8tV3XbFu2YvEy+0SYW8OVAFCqNoRPcdymeF7op3QXLL/C7DHgISR/jOjLwWW3eTz+9S4n7ybyBySA50/cp7DkEOdNOGWfkK3fmduz3Z/dfvs539GVhYGGpOYcVAWmri1UWlAUIQ0IS7kMKC2VEbbWQmqj7shaY4WGik7lQAWzmygr19yCrKCSgtvVszit78koXuNYHGIqHl2KZxl2291+AT15qVomAQAA",
		},
	}

	logs := parseCloudWatchLogs(cloudWatchEvent)

	localTime := time.Local
	timeValue := time.Date(2021, time.April, 26, 15, 16, 43, 219000000, time.Local)
	if strings.Contains(localTime.String(), "UTC") { //Test case is running at system with time.Local as UTC
		timeValue = time.Date(2021, time.April, 26, 9, 46, 43, 219000000, time.Local)
	}

	expectedLMEvent := ingest.Log{
		Message:    "Test Log for kinesis firehose",
		Timestamp:  timeValue,
		ResourceID: map[string]string{"system.aws.arn": "arn:aws:firehose::197152445587:deliverystream/dataFirehose"},
		Metadata:   metadataMap,
	}

	assert.Equal(t, expectedLMEvent, logs[0])
}

func TestKinesisDataStreamLog(t *testing.T) {
	cloudWatchEvent := events.CloudwatchLogsEvent{
		AWSLogs: events.CloudwatchLogsRawData{
			Data: "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
		},
	}

	var metadataMap = map[string]string{"_integration": "aws", "_type": "kinesis.amazonaws.com"}

	logs := parseCloudWatchLogs(cloudWatchEvent)

	localTime := time.Local
	timeValue := time.Date(2021, time.April, 27, 14, 25, 50, 324000000, time.Local)
	if strings.Contains(localTime.String(), "UTC") { //Test case is running at system with time.Local as UTC
		timeValue = time.Date(2021, time.April, 27, 8, 55, 50, 324000000, time.Local)
	}

	expectedLMEvent := ingest.Log{
		Message:    "{\"eventVersion\":\"1.08\",\"userIdentity\":{\"type\":\"AssumedRole\",\"principalId\":\"AROAS3ZZTSSJTJMESL5PU:LMAssumeRoleSession\",\"arn\":\"arn:aws:sts::197152445587:assumed-role/BhushanPuriPortal/LMAssumeRoleSession\",\"accountId\":\"197152445587\",\"accessKeyId\":\"ASIAS3ZZTSSJV4QL3KY6\",\"sessionContext\":{\"sessionIssuer\":{\"type\":\"Role\",\"principalId\":\"AROAS3ZZTSSJTJMESL5PU\",\"arn\":\"arn:aws:iam::197152445587:role/BhushanPuriPortal\",\"accountId\":\"197152445587\",\"userName\":\"BhushanPuriPortal\"},\"webIdFederationData\":{},\"attributes\":{\"mfaAuthenticated\":\"false\",\"creationDate\":\"2021-04-27T08:34:28Z\"}}},\"eventTime\":\"2021-04-27T08:39:15Z\",\"eventSource\":\"kinesis.amazonaws.com\",\"eventName\":\"ListTagsForStream\",\"awsRegion\":\"ap-northeast-1\",\"sourceIPAddress\":\"34.220.47.95\",\"userAgent\":\"aws-sdk-java/1.11.918 Linux/4.14.193-149.317.amzn2.x86_64 OpenJDK_64-Bit_Server_VM/11.0.3+7-LTS java/11.0.3 vendor/Amazon.com_Inc.\",\"requestParameters\":{\"streamName\":\"kinesisTestSream\"},\"responseElements\":null,\"requestID\":\"f85f8e09-9fda-a448-a15e-41fa3263205e\",\"eventID\":\"d815bd8e-1016-46a6-8f68-608a42b9b7d3\",\"readOnly\":true,\"eventType\":\"AwsApiCall\",\"managementEvent\":true,\"eventCategory\":\"Management\",\"recipientAccountId\":\"197152445587\"}",
		Timestamp:  timeValue,
		ResourceID: map[string]string{"system.aws.arn": "arn:aws:kinesis::197152445587:stream/kinesisTestSream"},
		Metadata:   metadataMap,
	}

	assert.Equal(t, expectedLMEvent, logs[1])
}

func TestECSLog(t *testing.T) {
	cloudWatchEvent := events.CloudwatchLogsEvent{
		AWSLogs: events.CloudwatchLogsRawData{
			Data: "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
		},
	}
	var metadataMap = map[string]string{"_integration": "aws", "_type": "ecs.amazonaws.com"}
	logs := parseCloudWatchLogs(cloudWatchEvent)

	localTime := time.Local
	timeValue := time.Date(2021, time.April, 30, 14, 17, 41, 663000000, time.Local)
	if strings.Contains(localTime.String(), "UTC") { //Test case is running at system with time.Local as UTC
		timeValue = time.Date(2021, time.April, 30, 8, 47, 41, 663000000, time.Local)
	}

	expectedLMEvent := ingest.Log{
		Message:    "{\"eventVersion\":\"1.08\",\"userIdentity\":{\"type\":\"AssumedRole\",\"principalId\":\"AROAS3ZZTSSJQUWVOXM6Y:LMAssumeRoleSession\",\"arn\":\"arn:aws:sts::197152445587:assumed-role/lmngstockholm/LMAssumeRoleSession\",\"accountId\":\"197152445587\",\"accessKeyId\":\"ASIAS3ZZTSSJ57ASTCJ6\",\"sessionContext\":{\"sessionIssuer\":{\"type\":\"Role\",\"principalId\":\"AROAS3ZZTSSJQUWVOXM6Y\",\"arn\":\"arn:aws:iam::197152445587:role/lmngstockholm\",\"accountId\":\"197152445587\",\"userName\":\"lmngstockholm\"},\"webIdFederationData\":{},\"attributes\":{\"mfaAuthenticated\":\"false\",\"creationDate\":\"2021-04-30T08:30:58Z\"}}},\"eventTime\":\"2021-04-30T08:39:02Z\",\"eventSource\":\"ecs.amazonaws.com\",\"eventName\":\"ListTagsForResource\",\"awsRegion\":\"us-west-1\",\"sourceIPAddress\":\"34.219.122.65\",\"userAgent\":\"aws-sdk-java/1.11.918 Linux/4.14.193-149.317.amzn2.x86_64 OpenJDK_64-Bit_Server_VM/11.0.3+7-LTS java/11.0.3 vendor/Amazon.com_Inc.\",\"requestParameters\":{\"resourceArn\":\"arn:aws:ecs:us-west-1:197152445587:cluster/CVTestCluster\"},\"responseElements\":{\"tags\":[{\"key\":\"testKey\",\"value\":\"testValue\"}]},\"requestID\":\"ad778dc0-1c40-4599-a1d6-5bbbee2d17c2\",\"eventID\":\"f53bf60c-3b8b-4304-9a80-cacd7b02a275\",\"readOnly\":true,\"eventType\":\"AwsApiCall\",\"managementEvent\":true,\"eventCategory\":\"Management\",\"recipientAccountId\":\"197152445587\"}",
		Timestamp:  timeValue,
		ResourceID: map[string]string{"system.aws.arn": "arn:aws:ecs::197152445587:cluster/CVTestCluster"},
		Metadata:   metadataMap,
	}

	assert.Equal(t, expectedLMEvent, logs[44])
}

func TestELBlowLogs(t *testing.T) {
	var metadataMap = map[string]string{"_integration": "aws", "_type": "networkInterfaceId.amazonaws.com"}

	cloudWatchEvent := events.CloudwatchLogsEvent{
		AWSLogs: events.CloudwatchLogsRawData{
			Data: "H4sIAAAAAAAAAK2QT2sCMRDFv0rI2T+TZJLselvsVkopLeitSInrVJbq7rIbK0X87h2VUsFCD+0hTPIyefN+2csNdV1Y0eyjITmSN9kse3nIp9NsksuerHcVtSyr1CurEa1NPMvrejVp623DN8Ow64a0Xgwriru6fburIrWvoaBz2zS2FDbcR1XZh8KBNgtXLCnx4Cj0Q1FQE7m12y66oi2bWNbVbblmj06OnmWQ85NN/k5VPCp7WS7ZzTiVeK+tA6VTo3zKudBodCk4MAa9TzC1YDAxFhAQ8Vgd8qRYMnEMGw6vnNZeGW8SAOh9/QTba3EJLH7MLvj1QCEvbbkmQsFAGRzwQYOwimWhtEqdcEIJRPE97XKbjcf500w83stD729w9l/hLmGuSE9YZ8QTHPwONz98As9B+hNrAgAA",
		},
	}

	logs := parseCloudWatchLogs(cloudWatchEvent)

	localTime := time.Local
	timeValue := time.Date(2021, time.June, 03, 15, 18, 58, 000000000, time.Local)
	if strings.Contains(localTime.String(), "UTC") { //Test case is running at system with time.Local as UTC
		timeValue = time.Date(2021, time.June, 03, 9, 48, 58, 000000000, time.Local)
	}

	expectedLMEvent := ingest.Log{
		Message:    "2 197152445587 eni-0c6023b6cde8706ea 162.142.125.148 10.134.5.120 51125 12196 6 1 44 1622713738 1622713738 ACCEPT OK",
		Timestamp:  timeValue,
		ResourceID: map[string]string{"system.aws.networkInterfaceId": "eni-0c6023b6cde8706ea"},
		Metadata:   metadataMap,
	}

	assert.Equal(t, expectedLMEvent, logs[0])
}

func TestRDSFlowLogs(t *testing.T) {
	var metadataMap = map[string]string{"_integration": "aws", "_type": "rds.amazonaws.com"}

	cloudWatchEvent := events.CloudwatchLogsEvent{
		AWSLogs: events.CloudwatchLogsRawData{
			Data: "H4sIAAAAAAAAAG2Qy2rDMBBFf8VoHWO9LGmyM9QNXbRdOLsSiiLLwTR+ICk1JeTfO2kJbaEzm+HO5czjTAYfoz347cfsyZrcVdvq9bFummpTkxWZltEHlJk0RoLSwKhG+TgdNmE6zdgp7BKL0MZi9GmZwtvDmHzorPPftiYFbwf0+bHPKTjlulYp7oTpWtnm1jk/J7TG0z660M+pn8b7/oiMSNYvxJLdF6Z+92O6KmfSt0gTihsGXHKGMKpZKTiDUgKlnDHQknIhQVAqFIDkoI00TJTK4KTU48XJDrg8U1yB0LqkGKvbJxDPs98HZ//unuV/8od1Kw0F1J+ery8ll93lE2T4ipxrAQAA",
		},
	}

	logs := parseCloudWatchLogs(cloudWatchEvent)

	localTime := time.Local
	timeValue := time.Date(2021, time.July, 22, 12, 39, 10, 000000000, time.Local)
	if strings.Contains(localTime.String(), "UTC") { //Test case is running at system with time.Local as UTC
		timeValue = time.Date(2021, time.July, 22, 07, 9, 10, 000000000, time.Local)
	}

	expectedLMEvent := ingest.Log{
		Message:    "2 148849679107 eni-09c6cfd662c38fd4d - - - - - - - 1626937750 1626937809 - NODATA",
		Timestamp:  timeValue,
		ResourceID: map[string]string{"system.aws.networkInterfaceId": "eni-09c6cfd662c38fd4d"},
		Metadata:   metadataMap,
	}

	assert.Equal(t, expectedLMEvent, logs[0])
}

func TestFargateLog(t *testing.T) {
	var metadataMap = map[string]string{"_integration": "aws", "_type": "fargate.amazonaws.com"}
	cloudWatchEvent := events.CloudwatchLogsEvent{
		AWSLogs: events.CloudwatchLogsRawData{
			Data: "H4sIAAAAAAAAAH3SS2/bMAwA4L8S+Nyk4ksUcyu2pqed0p2GopBjuTCaF2ynxVD0v4/JMGDdksIXixTET6Teqk0ZhvxU7n/uSzWvvt7c3zx+u10ub+5uq6tq97otvYeBU2KLahDUw+vd012/O+w9c51fh+s29095LL8zy7EveeOptt9tpu36ULbjtO7G6fOhLrOX3M9802y1246589MH/z1sx9I/NqXNh/X4eFpP21YlNyVrxiZKSzVLHepmldu2ZiyQAVMgjK2J1ivQhoiaOgOrNGw1H6s4aDjUw6rv9mO32y66tdcZqvmPKlcPJ+zti+uOkbeqa9xMMaZEaoRECsEXkgxO1xdRQ9BggckAlRFVIxCnGJJXGjtv5Zg33hWILBApoRLD1Z8W+/GCYT5Z9N1kUeoJpEmQueg86OT7/ZcJBsTq/epfCUs0PhICRg7Ry2JgDyEnEwLExMEEWZUloFyQJGL8KIFzkvSZJKYACjEoR4qmJpYEOEZNKUVvGBpZPN3ZP8YLEr8RfZTgOYl9JrFAoBbEmC0EBp+Hp5BMQRwDUYRjUjgOjZnOSyj8J6EzEgifSBiOQzGBUwv8KaI/FxKL4hOz48iiEEUMvof8oVyQ+CvijxI+J4G/JA/vvwBKpFa1vAMAAA==",
		},
	}
	logs := parseCloudWatchLogs(cloudWatchEvent)
	localTime := time.Local
	timeValue := time.Date(2022, time.February, 18, 11, 27, 07, 341000000, time.Local)
	if strings.Contains(localTime.String(), "UTC") { //Test case is running at system with time.Local as UTC
		timeValue = time.Date(2022, time.February, 18, 5, 57, 07, 341000000, time.Local)
	}
	expectedLMEvent := ingest.Log{
		Message:    "520: Fri Feb 18 05:57:07 UTC 2022",
		Timestamp:  timeValue,
		ResourceID: map[string]string{"system.aws.accountid": "148849679107", "system.cloud.category": "AWS/LMAccount"},
		Metadata:   metadataMap,
	}
	assert.Equal(t, expectedLMEvent, logs[0])
}

func TestParseCloudtrailLogsS3(t *testing.T) {
	cloudWatchEvent := events.CloudwatchLogsEvent{
		AWSLogs: events.CloudwatchLogsRawData{
			Data: "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",
		},
	}

	var metadataMap = map[string]string{"_integration": "aws", "_type": "s3.amazonaws.com"}
	logs := parseCloudWatchLogs(cloudWatchEvent)

	expectedLMEvent := ingest.Log{
		Message:    "{\"eventVersion\":\"1.08\",\"userIdentity\":{\"type\":\"AWSService\",\"invokedBy\":\"cloudtrail.amazonaws.com\"},\"eventTime\":\"2023-03-03T07:30:04Z\",\"eventSource\":\"s3.amazonaws.com\",\"eventName\":\"GetBucketAcl\",\"awsRegion\":\"us-east-1\",\"sourceIPAddress\":\"cloudtrail.amazonaws.com\",\"userAgent\":\"cloudtrail.amazonaws.com\",\"requestParameters\":{\"bucketName\":\"aws-cloudtrail-logs-700010466334-8d075b05\",\"Host\":\"aws-cloudtrail-logs-700010466334-8d075b05.s3.us-east-1.amazonaws.com\",\"acl\":\"\"},\"responseElements\":null,\"additionalEventData\":{\"SignatureVersion\":\"SigV4\",\"CipherSuite\":\"ECDHE-RSA-AES128-GCM-SHA256\",\"bytesTransferredIn\":0,\"AuthenticationMethod\":\"AuthHeader\",\"x-amz-id-2\":\"La8vQCxEf9pcqy/H8Y7Rs7aULfw0Qkc0EI+uKOFTyuMu8of/2a/yvPO6hKck3V5YaGneBCVwzkw=\",\"bytesTransferredOut\":568},\"requestID\":\"TAQNDGTZC32834P4\",\"eventID\":\"2514d9c5-5365-4b24-ac86-241dea825ad6\",\"readOnly\":true,\"resources\":[{\"accountId\":\"700010466334\",\"type\":\"AWS::S3::Bucket\",\"ARN\":\"arn:aws:s3:::aws-cloudtrail-logs-700010466334-8d075b05\"}],\"eventType\":\"AwsApiCall\",\"managementEvent\":true,\"recipientAccountId\":\"700010466334\",\"sharedEventID\":\"59523b9c-953d-4110-b4d5-b8209621af88\",\"eventCategory\":\"Management\"}",
		Timestamp:  time.Date(2023, time.March, 3, 7, 30, 27, 87000000, time.Local),
		ResourceID: map[string]string{"system.aws.arn": "arn:aws:s3:::aws-cloudtrail-logs-700010466334-8d075b05"},
		Metadata:   metadataMap,
	}

	assert.Equal(t, expectedLMEvent, logs[0])
}

func TestParseCloudtrailLogsLambda(t *testing.T) {
	cloudWatchEvent := events.CloudwatchLogsEvent{
		AWSLogs: events.CloudwatchLogsRawData{
			Data: "H4sIAAAAAAAAAO1W227bRhD9F6JPhdbZO3f5xtpyYMBJg0hogUaGMNxdqkQpUiUpG47hf++QlBTZadIAcR8CCCBAcOfM7czskR6idWhbWIX5/SZESXSRztPlm+lslr6eRpOovqtCg8fcUCmFotRwisdlvXrd1NsNWl7BXfvKlfXWdw0U5WicdU2A9TO/5XmPmveo5bYlAdqOsKVEj3abta4pNl1RV5dF2YWmjZIP0S5qAeWyhGztYdkDo5shxfQ2VF0Pe4gKj5lELLmOpTGMcqm5EpZTFhsrtdZGCM0lfnLJqbUqZpzHylKtdN9NVyAHHayxHaZjw2MhNeXCTvbcYPiHRRT6jL9hbVjlIkoWETujZhFNFtG2Dc2VR2vR3aMFsR2yOWDS32ez0NwWLgzIorqt/wr+l/vBiG20Z7CGj3WFLJ65er2IHie7THOsakBxrIVQfMyc0UTZhLI/hmADbFZvGzcCS1hnHp4H3APfwi7e1VDDYEDQ+7Da93MYymBrh8BX71LvG6ThiwXv+k9XmONroCb8vUWa30GDhfQjHpnKt5Xr534oD5oqQb9k7CY5FJUc71Kyd0uu31xjvsu6uYPGh+bAsoPePN/PYViXo77SpnqaDYN8KRfayKpf9+TZsic/H0d0rt7uKDh2H0eKFG7qqg3TMqz7xUVYtS3LfgbeF32pUA41XkAHT5k5XrnvIyf56TqdT2fzfUnDQK4uhsh54ILTmBJ8A5GgGMlCLglkTINVIc6M+7RMOyePN8tBTok01hOpLSfAjSZKecd1iLVTsBs++F+rsl/7HMo2jIwMvPVUfMB2YeTvyv8Lg5OnNypJrsf+k8tdpwMkff/2RTbo8eZwBw9J79p0U5xDWQ6Z1lChLPSD3O3VUVeu2BR4ln69nfZPaIKfHnGpPfM8t4544WMireDEOgZEcaUUGCmspZ8GcA5dWNXNqCPjzjziVL9PDNmPJYYMlTDh8UkMT2L4wmIIUsiMgiI2Y3gXKbPEWEqJ9tZncaBZsPxzMbTC61woQplHBRWZIUZLSix3woAGShU7ieG3iWGQmeDgBQGhGZFK4Q+LCJTElFrAL0H1/y2G/IcTQ5EwcxLDkxi+sBhy77V2sSeGcUski2MCNJYkMEc1eJcbDp+JofKZNbkIRHmESgiBmDjkuKqUO68soCSexPDbxFABBAU0JxZ5J5Ljv+yM4bU3sUI/a5yh9r/E8ObxH0p2zy7qDwAA",
		},
	}

	var metadataMap = map[string]string{"_integration": "aws", "_type": "lambda.amazonaws.com"}
	logs := parseCloudWatchLogs(cloudWatchEvent)
	expectedLMEvent := ingest.Log{
		Message:    "{\"eventVersion\":\"1.08\",\"userIdentity\":{\"type\":\"AWSService\",\"invokedBy\":\"logs.amazonaws.com\"},\"eventTime\":\"2023-03-08T10:59:01Z\",\"eventSource\":\"lambda.amazonaws.com\",\"eventName\":\"Invoke\",\"awsRegion\":\"us-east-1\",\"sourceIPAddress\":\"logs.amazonaws.com\",\"userAgent\":\"logs.amazonaws.com\",\"requestParameters\":{\"functionName\":\"arn:aws:lambda:us-east-1:280443500820:function:LMLogsForwarder\",\"invocationType\":\"Event\",\"sourceArn\":\"arn:aws:logs:us-east-1:280443500820:log-group:/aws/cloudtrail:*\",\"sourceAccount\":\"280443500820\"},\"responseElements\":null,\"additionalEventData\":{\"functionVersion\":\"arn:aws:lambda:us-east-1:280443500820:function:LMLogsForwarder:$LATEST\"},\"requestID\":\"fe232070-e23a-4a51-bef4-ab16a95e7b8c\",\"eventID\":\"d833caf0-489d-4692-a286-55dc26e76c5a\",\"readOnly\":false,\"resources\":[{\"accountId\":\"280443500820\",\"type\":\"AWS::Lambda::Function\",\"ARN\":\"arn:aws:lambda:us-east-1:280443500820:function:LMLogsForwarder\"}],\"eventType\":\"AwsApiCall\",\"managementEvent\":false,\"recipientAccountId\":\"280443500820\",\"sharedEventID\":\"6d1d2f9c-d3d7-4932-9c1a-52555a843990\",\"eventCategory\":\"Data\"}",
		Timestamp:  time.Date(2023, time.March, 8, 11, 4, 20, 239000000, time.Local),
		ResourceID: map[string]string{"system.aws.arn": "arn:aws:lambda:us-east-1:280443500820:function:LMLogsForwarder"},
		Metadata:   metadataMap,
	}

	assert.Equal(t, expectedLMEvent, logs[0])
}

func TestCloudTrailLogsEC2(t *testing.T) {
	var s = "{\"owner\":\"123456678\",\"logGroup\":\"/aws/cloudtrail\",\"logStream\":\"123456678_CloudTrail_us-west-2_3\",\"subscriptionFilters\":[\"testsubslambda\"],\"messageType\":\"DATA_MESSAGE\",\"logEvents\":[{\"id\":\"\",\"timestamp\":123456789,\"message\":\"{\\\"eventVersion\\\":\\\"1.08\\\",\\\"userIdentity\\\":{\\\"type\\\":\\\"AssumedRole\\\",\\\"principalId\\\":\\\"AROAUCS54HEKDKHAECU5N:pooja.choudhary@logicmonitor.com\\\",\\\"arn\\\":\\\"arn:aws:sts::123456678:assumed-role/AWSReservedSSO_LM-Developer-Policy_c0615b7abc4ebbd6/pooja.choudhary@logicmonitor.com\\\",\\\"accountId\\\":\\\"123456678\\\",\\\"accessKeyId\\\":\\\"ASIAUCS54HEKDHXBQTVJ\\\",\\\"sessionContext\\\":{\\\"sessionIssuer\\\":{\\\"type\\\":\\\"Role\\\",\\\"principalId\\\":\\\"AROAUCS54HEKDKHAECU5N\\\",\\\"arn\\\":\\\"arn:aws:iam::123456678:role/aws-reserved/sso.amazonaws.com/AWSReservedSSO_LM-Developer-Policy_c0615b7abc4ebbd6\\\",\\\"accountId\\\":\\\"123456678\\\",\\\"userName\\\":\\\"AWSReservedSSO_LM-Developer-Policy_c0615b7abc4ebbd6\\\"},\\\"webIdFederationData\\\":{},\\\"attributes\\\":{\\\"creationDate\\\":\\\"2023-05-23T05:07:33Z\\\",\\\"mfaAuthenticated\\\":\\\"false\\\"}}},\\\"eventTime\\\":\\\"2023-05-23T05:36:11Z\\\",\\\"eventSource\\\":\\\"ec2.amazonaws.com\\\",\\\"eventName\\\":\\\"DescribeInstanceAttribute\\\",\\\"awsRegion\\\":\\\"us-west-2\\\",\\\"sourceIPAddress\\\":\\\"49.207.217.191\\\",\\\"userAgent\\\":\\\"AWSInternal\\\",\\\"requestParameters\\\":{\\\"instanceId\\\":\\\"i-0d51cd459226160ac\\\",\\\"attribute\\\":\\\"disableApiTermination\\\"},\\\"responseElements\\\":null,\\\"requestID\\\":\\\"468e0670-30d9-4ef7-ab75-648c03d63371\\\",\\\"eventID\\\":\\\"78294249-5828-4dde-8b53-49f81d715b95\\\",\\\"readOnly\\\":true,\\\"eventType\\\":\\\"AwsApiCall\\\",\\\"managementEvent\\\":true,\\\"recipientAccountId\\\":\\\"123456678\\\",\\\"eventCategory\\\":\\\"Management\\\",\\\"sessionCredentialFromConsole\\\":\\\"true\\\"}\"}]}"
	var data events.CloudwatchLogsData
	err := json.Unmarshal([]byte(s), &data)
	if err != nil {
		fmt.Println("error in unmarshal")
	}
	logs := parseCloudTrailLogs(data)
	var metadataMap = map[string]string{"_integration": "aws", "_type": "ec2.amazonaws.com"}
	expectedLMEvent := ingest.Log{
		Message:    "{\"eventVersion\":\"1.08\",\"userIdentity\":{\"type\":\"AssumedRole\",\"principalId\":\"AROAUCS54HEKDKHAECU5N:pooja.choudhary@logicmonitor.com\",\"arn\":\"arn:aws:sts::123456678:assumed-role/AWSReservedSSO_LM-Developer-Policy_c0615b7abc4ebbd6/pooja.choudhary@logicmonitor.com\",\"accountId\":\"123456678\",\"accessKeyId\":\"ASIAUCS54HEKDHXBQTVJ\",\"sessionContext\":{\"sessionIssuer\":{\"type\":\"Role\",\"principalId\":\"AROAUCS54HEKDKHAECU5N\",\"arn\":\"arn:aws:iam::123456678:role/aws-reserved/sso.amazonaws.com/AWSReservedSSO_LM-Developer-Policy_c0615b7abc4ebbd6\",\"accountId\":\"123456678\",\"userName\":\"AWSReservedSSO_LM-Developer-Policy_c0615b7abc4ebbd6\"},\"webIdFederationData\":{},\"attributes\":{\"creationDate\":\"2023-05-23T05:07:33Z\",\"mfaAuthenticated\":\"false\"}}},\"eventTime\":\"2023-05-23T05:36:11Z\",\"eventSource\":\"ec2.amazonaws.com\",\"eventName\":\"DescribeInstanceAttribute\",\"awsRegion\":\"us-west-2\",\"sourceIPAddress\":\"49.207.217.191\",\"userAgent\":\"AWSInternal\",\"requestParameters\":{\"instanceId\":\"i-0d51cd459226160ac\",\"attribute\":\"disableApiTermination\"},\"responseElements\":null,\"requestID\":\"468e0670-30d9-4ef7-ab75-648c03d63371\",\"eventID\":\"78294249-5828-4dde-8b53-49f81d715b95\",\"readOnly\":true,\"eventType\":\"AwsApiCall\",\"managementEvent\":true,\"recipientAccountId\":\"123456678\",\"eventCategory\":\"Management\",\"sessionCredentialFromConsole\":\"true\"}",
		Timestamp:  time.Date(1970, time.January, 2, 10, 17, 36, 789000000, time.Local),
		ResourceID: map[string]string{"system.aws.arn": "arn:aws:ec2::123456678:instance/i-0d51cd459226160ac"},
		Metadata:   metadataMap,
	}
	assert.Equal(t, expectedLMEvent, logs[0])
}

func TestCloudTrailLogsSQS(t *testing.T) {
	var s = "{\"owner\":\"123456678\",\"logGroup\":\"/aws/cloudtrail\",\"logStream\":\"123456678_CloudTrail_us-west-2_3\",\"subscriptionFilters\":[\"testsubslambda\"],\"messageType\":\"DATA_MESSAGE\",\"logEvents\":[{\"id\":\"\",\"timestamp\":123456789,\"message\":\"{\\\"eventVersion\\\":\\\"1.08\\\",\\\"userIdentity\\\":{\\\"type\\\":\\\"AssumedRole\\\",\\\"principalId\\\":\\\"AROAUCS54HEKDKHAECU5N:pooja.choudhary@logicmonitor.com\\\",\\\"arn\\\":\\\"arn:aws:sts::123456678:assumed-role/AWSReservedSSO_LM-Developer-Policy_c0615b7abc4ebbd6/pooja.choudhary@logicmonitor.com\\\",\\\"accountId\\\":\\\"123456678\\\",\\\"accessKeyId\\\":\\\"ASIAUCS54HEKPPHDNOGC\\\",\\\"sessionContext\\\":{\\\"sessionIssuer\\\":{\\\"type\\\":\\\"Role\\\",\\\"principalId\\\":\\\"AROAUCS54HEKDKHAECU5N\\\",\\\"arn\\\":\\\"arn:aws:iam::123456678:role/aws-reserved/sso.amazonaws.com/AWSReservedSSO_LM-Developer-Policy_c0615b7abc4ebbd6\\\",\\\"accountId\\\":\\\"123456678\\\",\\\"userName\\\":\\\"AWSReservedSSO_LM-Developer-Policy_c0615b7abc4ebbd6\\\"},\\\"webIdFederationData\\\":{},\\\"attributes\\\":{\\\"creationDate\\\":\\\"2023-05-03T08:34:17Z\\\",\\\"mfaAuthenticated\\\":\\\"false\\\"}}},\\\"eventTime\\\":\\\"2023-05-03T09:16:01Z\\\",\\\"eventSource\\\":\\\"sqs.amazonaws.com\\\",\\\"eventName\\\":\\\"CreateQueue\\\",\\\"awsRegion\\\":\\\"us-west-2\\\",\\\"sourceIPAddress\\\":\\\"49.207.235.15\\\",\\\"userAgent\\\":\\\"AWSInternal\\\",\\\"requestParameters\\\":{\\\"attribute\\\":{\\\"Policy\\\":\\\"{\\\\\\\"Version\\\\\\\":\\\\\\\"2012-10-17\\\\\\\",\\\\\\\"Id\\\\\\\":\\\\\\\"__default_policy_ID\\\\\\\",\\\\\\\"Statement\\\\\\\":[{\\\\\\\"Sid\\\\\\\":\\\\\\\"__owner_statement\\\\\\\",\\\\\\\"Effect\\\\\\\":\\\\\\\"Allow\\\\\\\",\\\\\\\"Principal\\\\\\\":{\\\\\\\"AWS\\\\\\\":\\\\\\\"123456678\\\\\\\"},\\\\\\\"Action\\\\\\\":[\\\\\\\"SQS:*\\\\\\\"],\\\\\\\"Resource\\\\\\\":\\\\\\\"arn:aws:sqs:us-west-2:123456678:TestPoojaNew\\\\\\\"}]}\\\",\\\"ReceiveMessageWaitTimeSeconds\\\":\\\"0\\\",\\\"SqsManagedSseEnabled\\\":\\\"true\\\",\\\"DelaySeconds\\\":\\\"0\\\",\\\"KmsMasterKeyId\\\":\\\"\\\",\\\"RedrivePolicy\\\":\\\"\\\",\\\"MessageRetentionPeriod\\\":\\\"345600\\\",\\\"MaximumMessageSize\\\":\\\"262144\\\",\\\"VisibilityTimeout\\\":\\\"30\\\",\\\"RedriveAllowPolicy\\\":\\\"\\\"},\\\"tags\\\":{\\\"test\\\":\\\"true\\\"}},\\\"responseElements\\\":{\\\"queueUrl\\\":\\\"https://sqs.us-west-2.amazonaws.com/123456678/TestPoojaNew\\\"},\\\"requestID\\\":\\\"7c41e87a-a828-5717-8eb2-b4b680b92479\\\",\\\"eventID\\\":\\\"85297ed4-f144-4afb-84ce-63e9e58556b4\\\",\\\"readOnly\\\":false,\\\"eventType\\\":\\\"AwsApiCall\\\",\\\"managementEvent\\\":true,\\\"recipientAccountId\\\":\\\"123456678\\\",\\\"eventCategory\\\":\\\"Management\\\",\\\"sessionCredentialFromConsole\\\":\\\"true\\\"}\"}]}"
	var data events.CloudwatchLogsData
	err := json.Unmarshal([]byte(s), &data)
	if err != nil {
		fmt.Println("error in unmarshal")
	}
	logs := parseCloudTrailLogs(data)
	var metadataMap = map[string]string{"_integration": "aws", "_type": "sqs.amazonaws.com"}
	expectedLMEvent := ingest.Log{
		Message:    "{\"eventVersion\":\"1.08\",\"userIdentity\":{\"type\":\"AssumedRole\",\"principalId\":\"AROAUCS54HEKDKHAECU5N:pooja.choudhary@logicmonitor.com\",\"arn\":\"arn:aws:sts::123456678:assumed-role/AWSReservedSSO_LM-Developer-Policy_c0615b7abc4ebbd6/pooja.choudhary@logicmonitor.com\",\"accountId\":\"123456678\",\"accessKeyId\":\"ASIAUCS54HEKPPHDNOGC\",\"sessionContext\":{\"sessionIssuer\":{\"type\":\"Role\",\"principalId\":\"AROAUCS54HEKDKHAECU5N\",\"arn\":\"arn:aws:iam::123456678:role/aws-reserved/sso.amazonaws.com/AWSReservedSSO_LM-Developer-Policy_c0615b7abc4ebbd6\",\"accountId\":\"123456678\",\"userName\":\"AWSReservedSSO_LM-Developer-Policy_c0615b7abc4ebbd6\"},\"webIdFederationData\":{},\"attributes\":{\"creationDate\":\"2023-05-03T08:34:17Z\",\"mfaAuthenticated\":\"false\"}}},\"eventTime\":\"2023-05-03T09:16:01Z\",\"eventSource\":\"sqs.amazonaws.com\",\"eventName\":\"CreateQueue\",\"awsRegion\":\"us-west-2\",\"sourceIPAddress\":\"49.207.235.15\",\"userAgent\":\"AWSInternal\",\"requestParameters\":{\"attribute\":{\"Policy\":\"{\\\"Version\\\":\\\"2012-10-17\\\",\\\"Id\\\":\\\"__default_policy_ID\\\",\\\"Statement\\\":[{\\\"Sid\\\":\\\"__owner_statement\\\",\\\"Effect\\\":\\\"Allow\\\",\\\"Principal\\\":{\\\"AWS\\\":\\\"123456678\\\"},\\\"Action\\\":[\\\"SQS:*\\\"],\\\"Resource\\\":\\\"arn:aws:sqs:us-west-2:123456678:TestPoojaNew\\\"}]}\",\"ReceiveMessageWaitTimeSeconds\":\"0\",\"SqsManagedSseEnabled\":\"true\",\"DelaySeconds\":\"0\",\"KmsMasterKeyId\":\"\",\"RedrivePolicy\":\"\",\"MessageRetentionPeriod\":\"345600\",\"MaximumMessageSize\":\"262144\",\"VisibilityTimeout\":\"30\",\"RedriveAllowPolicy\":\"\"},\"tags\":{\"test\":\"true\"}},\"responseElements\":{\"queueUrl\":\"https://sqs.us-west-2.amazonaws.com/123456678/TestPoojaNew\"},\"requestID\":\"7c41e87a-a828-5717-8eb2-b4b680b92479\",\"eventID\":\"85297ed4-f144-4afb-84ce-63e9e58556b4\",\"readOnly\":false,\"eventType\":\"AwsApiCall\",\"managementEvent\":true,\"recipientAccountId\":\"123456678\",\"eventCategory\":\"Management\",\"sessionCredentialFromConsole\":\"true\"}",
		Timestamp:  time.Date(1970, time.January, 2, 10, 17, 36, 789000000, time.Local),
		ResourceID: map[string]string{"system.aws.arn": "arn:aws:sqs::123456678:TestPoojaNew"},
		Metadata:   metadataMap,
	}
	assert.Equal(t, expectedLMEvent, logs[0])
}

func TestParseCloudtrailLogsS3ForARN(t *testing.T) {
	var s = "{\"owner\":\"123456678\",\"logGroup\":\"/aws/cloudtrail\",\"logStream\":\"123456678_CloudTrail_us-west-2_3\",\"subscriptionFilters\":[\"testsubslambda\"],\"messageType\":\"DATA_MESSAGE\",\"logEvents\":[{\"id\":\"\",\"timestamp\":123456789,\"message\":\"{\\\"eventVersion\\\":\\\"1.08\\\",\\\"userIdentity\\\":{\\\"type\\\":\\\"AWSService\\\",\\\"invokedBy\\\":\\\"cloudtrail.amazonaws.com\\\"},\\\"eventTime\\\":\\\"2023-03-03T07:30:04Z\\\",\\\"eventSource\\\":\\\"s3.amazonaws.com\\\",\\\"eventName\\\":\\\"GetBucketAcl\\\",\\\"awsRegion\\\":\\\"us-east-1\\\",\\\"sourceIPAddress\\\":\\\"cloudtrail.amazonaws.com\\\",\\\"userAgent\\\":\\\"cloudtrail.amazonaws.com\\\",\\\"requestParameters\\\":{\\\"Host\\\":\\\"aws-cloudtrail-logs-700010466334-8d075b05.s3.us-east-1.amazonaws.com\\\",\\\"acl\\\":\\\"\\\"},\\\"responseElements\\\":null,\\\"additionalEventData\\\":{\\\"SignatureVersion\\\":\\\"SigV4\\\",\\\"CipherSuite\\\":\\\"ECDHE-RSA-AES128-GCM-SHA256\\\",\\\"bytesTransferredIn\\\":0,\\\"AuthenticationMethod\\\":\\\"AuthHeader\\\",\\\"x-amz-id-2\\\":\\\"La8vQCxEf9pcqy/H8Y7Rs7aULfw0Qkc0EI+uKOFTyuMu8of/2a/yvPO6hKck3V5YaGneBCVwzkw=\\\",\\\"bytesTransferredOut\\\":568},\\\"requestID\\\":\\\"TAQNDGTZC32834P4\\\",\\\"eventID\\\":\\\"2514d9c5-5365-4b24-ac86-241dea825ad6\\\",\\\"readOnly\\\":true,\\\"resources\\\":[{\\\"accountId\\\":\\\"700010466334\\\",\\\"type\\\":\\\"AWS::S3::Bucket\\\",\\\"ARN\\\":\\\"arn:aws:s3:::aws-cloudtrail-logs-700010466334-8d075b05\\\"}],\\\"eventType\\\":\\\"AwsApiCall\\\",\\\"managementEvent\\\":true,\\\"recipientAccountId\\\":\\\"700010466334\\\",\\\"sharedEventID\\\":\\\"59523b9c-953d-4110-b4d5-b8209621af88\\\",\\\"eventCategory\\\":\\\"Management\\\"}\"}]}"
	var data events.CloudwatchLogsData
	err := json.Unmarshal([]byte(s), &data)
	if err != nil {
		fmt.Println("error in unmarshal")
	}
	logs := parseCloudTrailLogs(data)

	var metadataMap = map[string]string{"_integration": "aws", "_type": "s3.amazonaws.com"}

	expectedLMEvent := ingest.Log{
		Message:    "{\"eventVersion\":\"1.08\",\"userIdentity\":{\"type\":\"AWSService\",\"invokedBy\":\"cloudtrail.amazonaws.com\"},\"eventTime\":\"2023-03-03T07:30:04Z\",\"eventSource\":\"s3.amazonaws.com\",\"eventName\":\"GetBucketAcl\",\"awsRegion\":\"us-east-1\",\"sourceIPAddress\":\"cloudtrail.amazonaws.com\",\"userAgent\":\"cloudtrail.amazonaws.com\",\"requestParameters\":{\"Host\":\"aws-cloudtrail-logs-700010466334-8d075b05.s3.us-east-1.amazonaws.com\",\"acl\":\"\"},\"responseElements\":null,\"additionalEventData\":{\"SignatureVersion\":\"SigV4\",\"CipherSuite\":\"ECDHE-RSA-AES128-GCM-SHA256\",\"bytesTransferredIn\":0,\"AuthenticationMethod\":\"AuthHeader\",\"x-amz-id-2\":\"La8vQCxEf9pcqy/H8Y7Rs7aULfw0Qkc0EI+uKOFTyuMu8of/2a/yvPO6hKck3V5YaGneBCVwzkw=\",\"bytesTransferredOut\":568},\"requestID\":\"TAQNDGTZC32834P4\",\"eventID\":\"2514d9c5-5365-4b24-ac86-241dea825ad6\",\"readOnly\":true,\"resources\":[{\"accountId\":\"700010466334\",\"type\":\"AWS::S3::Bucket\",\"ARN\":\"arn:aws:s3:::aws-cloudtrail-logs-700010466334-8d075b05\"}],\"eventType\":\"AwsApiCall\",\"managementEvent\":true,\"recipientAccountId\":\"700010466334\",\"sharedEventID\":\"59523b9c-953d-4110-b4d5-b8209621af88\",\"eventCategory\":\"Management\"}",
		Timestamp:  time.Date(1970, time.January, 2, 10, 17, 36, 789000000, time.Local),
		ResourceID: map[string]string{"system.aws.arn": "arn:aws:s3:::aws-cloudtrail-logs-700010466334-8d075b05"},
		Metadata:   metadataMap,
	}

	assert.Equal(t, expectedLMEvent, logs[0])
}

func TestCloudWatchEventsS3(t *testing.T) {
	var s = "{\"account\":\"280443500820\",\"detail\":{\"additionalEventData\":{\"AuthenticationMethod\":\"AuthHeader\",\"CipherSuite\":\"ECDHE-RSA-AES128-GCM-SHA256\",\"SignatureVersion\":\"SigV4\",\"bytesTransferredIn\":0,\"bytesTransferredOut\":0,\"x-amz-id-2\":\"UJNZntIyly1aAdiFVSPCwh14QRbyGaVQsbTjtgaNHRYgy0gl4p8c2Uu8Iu498icXe/3JHf9VFHo=\"},\"awsRegion\":\"us-west-2\",\"eventCategory\":\"Data\",\"eventID\":\"b9d131b3-70c7-4f74-a203-bddda20b7469\",\"eventName\":\"HeadBucket\",\"eventSource\":\"s3.amazonaws.com\",\"eventTime\":\"2023-07-13T09:52:54Z\",\"eventType\":\"AwsApiCall\",\"eventVersion\":\"1.08\",\"managementEvent\":false,\"readOnly\":true,\"recipientAccountId\":\"280443500820\",\"requestID\":\"J9ZE535Z4A4RVWZF\",\"requestParameters\":{\"Host\":\"sagemaker-studio-280443500820-6jk579yrjdw.s3.us-west-2.amazonaws.com\",\"bucketName\":\"sagemaker-studio-280443500820-6jk579yrjdw\"},\"resources\":[{\"ARNPrefix\":\"arn:aws:s3:::sagemaker-studio-280443500820-6jk579yrjdw/\",\"type\":\"AWS::S3::Object\"},{\"ARN\":\"arn:aws:s3:::sagemaker-studio-280443500820-6jk579yrjdw\",\"accountId\":\"1234567\",\"type\":\"AWS::S3::Bucket\"}],\"responseElements\":null,\"sourceIPAddress\":\"10.54.148.148\",\"tlsDetails\":{\"cipherSuite\":\"ECDHE-RSA-AES128-GCM-SHA256\",\"clientProvidedHostHeader\":\"sagemaker-studio-280443500820-6jk579yrjdw.s3.us-west-2.amazonaws.com\",\"tlsVersion\":\"TLSv1.2\"},\"userAgent\":\"[aws-sdk-java/1.12.498Linux/5.10.178-162.673.amzn2.x86_64OpenJDK_64-Bit_Server_VM/17.0.7+7-LTSjava/17.0.7vendor/Amazon.com_Inc.cfg/retry-mode/legacy]\",\"userIdentity\":{\"accessKeyId\":\"ASIAUCS54HEKCWHITL6Z\",\"accountId\":\"1234567\",\"arn\":\"arn:aws:sts::280443500820:assumed-role/aws-test-pooja-role/LMAssumeRoleSession\",\"principalId\":\"AROAUCS54HEKLBZ4YGEZZ:LMAssumeRoleSession\",\"sessionContext\":{\"attributes\":{\"creationDate\":\"2023-07-13T09:06:17Z\",\"mfaAuthenticated\":\"false\"},\"sessionIssuer\":{\"accountId\":\"1234567\",\"arn\":\"arn:aws:iam::280443500820:role/aws-test-pooja-role\",\"principalId\":\"AROAUCS54HEKLBZ4YGEZZ\",\"type\":\"Role\",\"userName\":\"aws-test-pooja-role\"}},\"type\":\"AssumedRole\"},\"vpcEndpointId\":\"vpce-051f8c152e5ab9b9d\"},\"detail-type\":\"AWS API Call via CloudTrail\",\"id\":\"0cc13242-9685-6b5a-4881-e970d28cc19c\",\"region\":\"us-west-2\",\"resources\":[],\"source\":\"aws.s3\",\"time\":\"2023-07-13T09:52:54Z\",\"version\":\"0\"}"
	var data events.CloudWatchEvent
	err := json.Unmarshal([]byte(s), &data)
	if err != nil {
		fmt.Println("error in unmarshal")
	}
	logs := parseCloudWatchEvents(data)
	var metadataMap = map[string]string{"_integration": "aws", "_type": "s3.amazonaws.com"}
	expectedLMEvent := ingest.Log{
		Message:    "{\"additionalEventData\":{\"AuthenticationMethod\":\"AuthHeader\",\"CipherSuite\":\"ECDHE-RSA-AES128-GCM-SHA256\",\"SignatureVersion\":\"SigV4\",\"bytesTransferredIn\":0,\"bytesTransferredOut\":0,\"x-amz-id-2\":\"UJNZntIyly1aAdiFVSPCwh14QRbyGaVQsbTjtgaNHRYgy0gl4p8c2Uu8Iu498icXe/3JHf9VFHo=\"},\"awsRegion\":\"us-west-2\",\"eventCategory\":\"Data\",\"eventID\":\"b9d131b3-70c7-4f74-a203-bddda20b7469\",\"eventName\":\"HeadBucket\",\"eventSource\":\"s3.amazonaws.com\",\"eventTime\":\"2023-07-13T09:52:54Z\",\"eventType\":\"AwsApiCall\",\"eventVersion\":\"1.08\",\"managementEvent\":false,\"readOnly\":true,\"recipientAccountId\":\"280443500820\",\"requestID\":\"J9ZE535Z4A4RVWZF\",\"requestParameters\":{\"Host\":\"sagemaker-studio-280443500820-6jk579yrjdw.s3.us-west-2.amazonaws.com\",\"bucketName\":\"sagemaker-studio-280443500820-6jk579yrjdw\"},\"resources\":[{\"ARNPrefix\":\"arn:aws:s3:::sagemaker-studio-280443500820-6jk579yrjdw/\",\"type\":\"AWS::S3::Object\"},{\"ARN\":\"arn:aws:s3:::sagemaker-studio-280443500820-6jk579yrjdw\",\"accountId\":\"1234567\",\"type\":\"AWS::S3::Bucket\"}],\"responseElements\":null,\"sourceIPAddress\":\"10.54.148.148\",\"tlsDetails\":{\"cipherSuite\":\"ECDHE-RSA-AES128-GCM-SHA256\",\"clientProvidedHostHeader\":\"sagemaker-studio-280443500820-6jk579yrjdw.s3.us-west-2.amazonaws.com\",\"tlsVersion\":\"TLSv1.2\"},\"userAgent\":\"[aws-sdk-java/1.12.498Linux/5.10.178-162.673.amzn2.x86_64OpenJDK_64-Bit_Server_VM/17.0.7+7-LTSjava/17.0.7vendor/Amazon.com_Inc.cfg/retry-mode/legacy]\",\"userIdentity\":{\"accessKeyId\":\"ASIAUCS54HEKCWHITL6Z\",\"accountId\":\"1234567\",\"arn\":\"arn:aws:sts::280443500820:assumed-role/aws-test-pooja-role/LMAssumeRoleSession\",\"principalId\":\"AROAUCS54HEKLBZ4YGEZZ:LMAssumeRoleSession\",\"sessionContext\":{\"attributes\":{\"creationDate\":\"2023-07-13T09:06:17Z\",\"mfaAuthenticated\":\"false\"},\"sessionIssuer\":{\"accountId\":\"1234567\",\"arn\":\"arn:aws:iam::280443500820:role/aws-test-pooja-role\",\"principalId\":\"AROAUCS54HEKLBZ4YGEZZ\",\"type\":\"Role\",\"userName\":\"aws-test-pooja-role\"}},\"type\":\"AssumedRole\"},\"vpcEndpointId\":\"vpce-051f8c152e5ab9b9d\"}",
		Timestamp:  time.Date(2023, time.July, 13, 15, 22, 54, 0, time.Local),
		ResourceID: map[string]string{"system.aws.arn": "arn:aws:s3:::sagemaker-studio-280443500820-6jk579yrjdw"},
		Metadata:   metadataMap,
	}
	assert.Equal(t, expectedLMEvent, logs[0])
}

func TestCloudWatchEventsLambda(t *testing.T) {
	var s = "{\"account\":\"280443500820\",\"detail\":{\"additionalEventData\":{\"functionVersion\":\"arn:aws:lambda:us-west-2:280443500820:function:LMLogsForwarder:$LATEST\"},\"awsRegion\":\"us-west-2\",\"eventCategory\":\"Data\",\"eventID\":\"e34805e1-c975-40ff-96e9-8fb01c4bd5ec\",\"eventName\":\"Invoke\",\"eventSource\":\"lambda.amazonaws.com\",\"eventTime\":\"2023-07-13T09:52:58Z\",\"eventType\":\"AwsApiCall\",\"eventVersion\":\"1.08\",\"managementEvent\":false,\"readOnly\":false,\"recipientAccountId\":\"280443500820\",\"requestID\":\"84fe1797-0fa3-4c28-a84b-318328f0ff1f\",\"requestParameters\":{\"functionName\":\"arn:aws:lambda:us-west-2:280443500820:function:LMLogsForwarder\",\"invocationType\":\"Event\",\"sourceAccount\":\"280443500820\",\"sourceArn\":\"arn:aws:logs:us-west-2:280443500820:log-group:/aws/events/cloudwatchEventsTest:*\"},\"resources\":[{\"ARN\":\"arn:aws:lambda:us-west-2:280443500820:function:LMLogsForwarder\",\"accountId\":\"280443500820\",\"type\":\"AWS::Lambda::Function\"}],\"responseElements\":null,\"sharedEventID\":\"779c751a-2152-4151-93b9-254d90b5819a\",\"sourceIPAddress\":\"logs.amazonaws.com\",\"userAgent\":\"logs.amazonaws.com\",\"userIdentity\":{\"invokedBy\":\"logs.amazonaws.com\",\"type\":\"AWSService\"}},\"detail-type\":\"AWS API Call via CloudTrail\",\"id\":\"e8954201-f3c4-54cd-9665-dee6ab22bf5a\",\"region\":\"us-west-2\",\"resources\":[],\"source\":\"aws.lambda\",\"time\":\"2023-07-13T09:52:58Z\",\"version\":\"0\"}"
	var data events.CloudWatchEvent
	err := json.Unmarshal([]byte(s), &data)
	if err != nil {
		fmt.Println("error in unmarshal")
	}
	logs := parseCloudWatchEvents(data)
	var metadataMap = map[string]string{"_integration": "aws", "_type": "lambda.amazonaws.com"}
	expectedLMEvent := ingest.Log{
		Message:    "{\"additionalEventData\":{\"functionVersion\":\"arn:aws:lambda:us-west-2:280443500820:function:LMLogsForwarder:$LATEST\"},\"awsRegion\":\"us-west-2\",\"eventCategory\":\"Data\",\"eventID\":\"e34805e1-c975-40ff-96e9-8fb01c4bd5ec\",\"eventName\":\"Invoke\",\"eventSource\":\"lambda.amazonaws.com\",\"eventTime\":\"2023-07-13T09:52:58Z\",\"eventType\":\"AwsApiCall\",\"eventVersion\":\"1.08\",\"managementEvent\":false,\"readOnly\":false,\"recipientAccountId\":\"280443500820\",\"requestID\":\"84fe1797-0fa3-4c28-a84b-318328f0ff1f\",\"requestParameters\":{\"functionName\":\"arn:aws:lambda:us-west-2:280443500820:function:LMLogsForwarder\",\"invocationType\":\"Event\",\"sourceAccount\":\"280443500820\",\"sourceArn\":\"arn:aws:logs:us-west-2:280443500820:log-group:/aws/events/cloudwatchEventsTest:*\"},\"resources\":[{\"ARN\":\"arn:aws:lambda:us-west-2:280443500820:function:LMLogsForwarder\",\"accountId\":\"280443500820\",\"type\":\"AWS::Lambda::Function\"}],\"responseElements\":null,\"sharedEventID\":\"779c751a-2152-4151-93b9-254d90b5819a\",\"sourceIPAddress\":\"logs.amazonaws.com\",\"userAgent\":\"logs.amazonaws.com\",\"userIdentity\":{\"invokedBy\":\"logs.amazonaws.com\",\"type\":\"AWSService\"}}",
		Timestamp:  time.Date(2023, time.July, 13, 15, 22, 58, 0, time.Local),
		ResourceID: map[string]string{"system.aws.arn": "arn:aws:lambda:us-west-2:280443500820:function:LMLogsForwarder"},
		Metadata:   metadataMap,
	}
	assert.Equal(t, expectedLMEvent, logs[0])
}

func TestCloudWatchEventsEC2Lanuch(t *testing.T) {
	var s = "{\"version\":\"0\",\"id\":\"1681ab87-4a09-459f-95a2-7fa09403c4b7\",\"detail-type\":\"EC2InstanceLaunchUnsuccessful\",\"source\":\"aws.autoscaling\",\"account\":\"123456789012\",\"time\":\"2015-11-11T21:42:36Z\",\"region\":\"us-east-1\",\"resources\":[\"arn:aws:autoscaling:us-east-1:123456789012:autoScalingGroup:528ffce5-ef9f-4c1d-8d18-5d005b4a438c:autoScalingGroupName/sampleBrokenASG\",\"arn:aws:ec2:us-east-1:123456789012:instance/\"],\"detail\":{\"StatusCode\":\"Failed\",\"AutoScalingGroupName\":\"brokenASG\",\"ActivityId\":\"06076c51-4874-487d-b15b-7895a713ab55\",\"Details\":{\"AvailabilityZone\":\"us-east-1e\",\"SubnetID\":\"subnet-16c5df2c\"},\"RequestId\":\"06076c51-4874-487d-b15b-7895a713ab55\",\"EndTime\":\"2015-11-11T21:42:36.000Z\",\"EC2InstanceId\":\"\",\"StartTime\":\"2015-11-11T21:42:36.698Z\",\"Cause\":\"At2015-11-11T21:42:09ZauserrequestupdateofAutoScalingGroupconstraintstomin:0,max:10,desired:2changingthedesiredcapacityfrom0to2.At2015-11-11T21:42:35Zaninstancewasstartedinresponsetoadifferencebetweendesiredandactualcapacity,increasingthecapacityfrom0to2.\"}}"
	var data events.CloudWatchEvent
	err := json.Unmarshal([]byte(s), &data)
	if err != nil {
		fmt.Println("error in unmarshal")
	}
	logs := parseCloudWatchEvents(data)

	assert.Len(t, logs, 0)
}
