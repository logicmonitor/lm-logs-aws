package main

import (
	"compress/gzip"
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
		}

		assert.Equal(t, expectedLMEvent, lmEvents[0])
	})
}

func TestParseS3logs(t *testing.T) {
	// Data preparation
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
		EventTime: time,
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
	}

	assert.Equal(t, expectedlmEvent, lmEvents[0])
}

func TestParseCloudWatchlogs(t *testing.T) {
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
	}

	assert.Equal(t, expectedLMEvent, lmEvents[0])
}

func TestParseCloudWatchlogsEks(t *testing.T) {
	cloudWatchEvent := events.CloudwatchLogsEvent{
		AWSLogs: events.CloudwatchLogsRawData{
			Data: "H4sIAAAAAAAAADWQzWrDMBCEX0XoXMerf9ulh0DT0ENPaU9NKLKtuCa2ZSS5IZS+ezehFQuCnZ0d9vumo4vRdu71Mjta0cf16/rjZbPbrbcbekf9eXIB27wAKYUCKDhge/DdNvhlRiW355i7U8ztKcMvQynmzbDEhMbb5C4FZ0ccPS21y2Lz6dplcCGTzjJQdaFKbnQrRFkKwxk/cq1Aa2XQHZc6NqGfU++np37AlZFW7xRzhnH2IdmBHm4Zmy83pav2TfsWo4QBxbguGWghgSkjFWYVEgzeoQyXjCvDGBdSG45ppsQeCI6ZqUciyY54HNOaSaGFLgDg7p8Urn9GDoowWYGuoFhJyTUAuT3GyNE2yYfLqvOVMOZA9vRtsvXgSPLk/3wy+/aeTB6rdZHY4EhwXX+l5trrYPp05A/jPTnbPvVTt6dX28P+j+QFxTFvPDqmmCExXcpGH2U2qXoc95T+HH5+AUTjfWfgAQAA",
		},
	}

	lmEvents := parseCloudWatchLogs(cloudWatchEvent)

	time := time.Unix(0, 1661436368000*1000000)
	expectedLMEvent := ingest.Log{
		Message:    "I0825 14:06:08.442600      11 factory.go:377] \"Unable to schedule pod; no nodes are registered to the cluster; waiting\" pod=\"kube-system/coredns-657694c6f4-n5bmm\"",
		Timestamp:  time,
		ResourceID: map[string]string{"system.aws.arn": "arn:aws:lambda::280443500820:function:ak-eks-logs"},
	}

	assert.Equal(t, expectedLMEvent, lmEvents[0])
}

func TestRDSLogs(t *testing.T) {
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
	}

	assert.Equal(t, expectedLMEvent, lmEvents[0])
}

func TestRDSEnhancedLogs(t *testing.T) {
	cloudWatchEvent := events.CloudwatchLogsEvent{
		AWSLogs: events.CloudwatchLogsRawData{
			Data: "H4sIAAAAAAAAAM1Z227bOBD9lYWeEy3vlPxmbJI2aLK5OOlitykKWaYdIrpFopwb8u87pGRbdtvsDdjyIYA4HA7PHM4c0shLkKumSRbq6qlSwSg4GF+Nv5weTibjd4fBXlA+FKoGsxAsopRyxmIC5qxcvKvLtoKZy4PJ2eRUmVqnTTczMbVKcpiaTffpeCzG14K8/yg+nL0nv4nja3n8qzhmF+DbtNMmrXVldFkc6cyouglGn4KTJJ/Oki7KF6Mas585S/DZhT9cqsJYx5dAz2AXygUSsYgJ5pIgKoUkjJOI45gzwUWMJEUcI4SJxJSgCMuIyIghAGA0JG+SHPLAPBZCYknAE+2tSIHwLzeBKha6UDfB6CY4/X1ycXIT7N0EuoCVRaqOD9zELDHJNGnUPtmavVRN2dYbr7cYcQvXkJw/QQTto2gf8StCR5yPCP7DuS2BK2ANnDCM2soucysQGhE0wtJ5FW3+8Zfz66Z3S6v22uhMPyemWwu5LVrYDj5RiCzs+n793Tw1RuVuGMHwIdGdH7F+s8xuF8uQ2u0bVdstQmETKE2SwYiE0gYxyo1QaAEUOlVd/FcYZWUyG0MiwPOpLlqjOkRlodYY5no5HMyNUsUmQK7ysn7qVj3U2qhpkt7ZaZi7bRfqHCI3R7VSu7bLZjnbtU3auuptaZLeKuvAKJRUtOWln200gpg1z7vYUDIRFUO3q54Ex2mRpKbLg8TQPtZW2YZLppmyR0OZtItnujY2GRJZjzypKodB4JhZwzoIJwxjMmAaIxxTB7PJkqmFTaSLOG3nc6gTG5NHhFnKTNLcNR1jTaZUpYuFPUfr/VzmU72iqm6LoptzpWDKHg0a7OtWTbMyveumbPzmIam68GsSh0sYinns1vXUbQy6WB902ZrNIRfKPJS1PdZPL9YLVGKepF2xK3OLXKHXj5YpzkJHg3l0NMY0pPz1s6O2uTs+60O4SvkwPZ9YLkNsCxiUZnZ81jgTVKot3GRd7xL3Hv0aFArrUNf3eT+2mFtorG7A7Orl4qJVrTpRm6xM5fowZOtwLn8LTy11n1E9a2Dkcupx2gJBURfzUt1PnsEQhSR2HlsQ3IJVGjgk9HVvJ10i17t/N11K/lm60ZvpkpDTt/Kda+iBXmi2csaCse2kCQmxGGZNQiJ30yYhle7Iq9unRqdJduD2+ruHj3fZcG0+YIP9JzZwv3rFRtfqAzoel7PFXx0+xqEU26cvol0eUNhVvuV38tT02YNQ246kEnTIhiiS/tZwe9rZI91pElyDA8u5gius6JuSOHV6XHliipF0opaXbWHOS+0cb4KfbS1P7b3YXWyrewHB3pStwm9C4zDirmJ7lAQu49j18/dhSk4p+yZQYIlvAxWcU/EtnFvwrJgijNhX8AiG1uhKqy5TeB6caHdxWl6XjQMjEI3JEO/Z5Kfe2SKwuyz0ShKrpIbA7l2A1nfZtd0ytUfUVYq9sVemnvp1gNptSiU8duyDIM903hWtI7GDRDCLCN7CBO+1fwuKiJDFO6hAz+TXqIhEyN2I34YF9UIiNkSVPzX32WwIh0Fi24jAgr8GhXknCztUrTD1YTpYmDNOt9mCYy7ct625/xciGWCM3sL446iLvYTFkZ+wsJ+wiEewBjXfPQw8wTWAxfyE9aaK/ThYwk9Y0k9YHsk85gNcPun8QCKET0I/xOWn0gtPlV74pPRDXH5KvfBT6oWfUi/8lPrIJ+kawPJJIgaw/OzEyM9OjPzsxMij180Qlk+Pmw2s2E+BiH162gzu6tjPVoz9bMXYz0sx9lMhYi8VIsJeXoqRVw+u9Y9Y+x/+t3F9fg3g709et82y8CAAAA==",
		},
	}

	logs := parseCloudWatchLogs(cloudWatchEvent)

	time := time.Unix(0, 1596671721000*1000000)
	expectedLMEvent := ingest.Log{
		Message:    "{\"engine\":\"MYSQL\",\"instanceID\":\"database-2\",\"instanceResourceID\":\"db-3AA6AU62HV6KOH2W6IU7IN6I4Q\",\"timestamp\":\"2020-08-05T23:55:21Z\",\"version\":1,\"uptime\":\"00:20:17\",\"numVCPUs\":1,\"cpuUtilization\":{\"guest\":0.0,\"irq\":0.0,\"system\":0.8,\"wait\":0.2,\"idle\":97.3,\"user\":1.6,\"total\":2.7,\"steal\":0.1,\"nice\":0.0},\"loadAverageMinute\":{\"one\":0.0,\"five\":0.0,\"fifteen\":0.0},\"memory\":{\"writeback\":0,\"hugePagesFree\":0,\"hugePagesRsvd\":0,\"hugePagesSurp\":0,\"cached\":435728,\"hugePagesSize\":2048,\"free\":100836,\"hugePagesTotal\":0,\"inactive\":294920,\"pageTables\":3476,\"dirty\":280,\"mapped\":61940,\"active\":524112,\"total\":1019328,\"slab\":42776,\"buffers\":25824},\"tasks\":{\"sleeping\":96,\"zombie\":0,\"running\":0,\"stopped\":0,\"total\":96,\"blocked\":0},\"swap\":{\"cached\":0,\"total\":4095996,\"free\":4095996,\"in\":0.0,\"out\":0.0},\"network\":[{\"interface\":\"eth0\",\"rx\":654.28,\"tx\":2893.35}],\"diskIO\":[{\"writeKbPS\":5.13,\"readIOsPS\":0.17,\"await\":0.71,\"readKbPS\":0.67,\"rrqmPS\":0.0,\"util\":0.04,\"avgQueueLen\":0.0,\"tps\":1.4,\"readKb\":40,\"device\":\"rdsdev\",\"writeKb\":308,\"avgReqSz\":8.29,\"wrqmPS\":0.0,\"writeIOsPS\":1.23},{\"writeKbPS\":27.4,\"readIOsPS\":0.17,\"await\":0.32,\"readKbPS\":0.67,\"rrqmPS\":0.0,\"util\":0.08,\"avgQueueLen\":0.0,\"tps\":2.53,\"readKb\":40,\"device\":\"filesystem\",\"writeKb\":1644,\"avgReqSz\":22.16,\"wrqmPS\":2.27,\"writeIOsPS\":2.37}],\"physicalDeviceIO\":[{\"writeKbPS\":5.13,\"readIOsPS\":1.17,\"await\":0.48,\"readKbPS\":4.67,\"rrqmPS\":0.0,\"util\":0.08,\"avgQueueLen\":0.0,\"tps\":1.67,\"readKb\":280,\"device\":\"xvdg\",\"writeKb\":308,\"avgReqSz\":11.76,\"wrqmPS\":0.68,\"writeIOsPS\":0.5}],\"fileSys\":[{\"used\":379496,\"name\":\"\",\"usedFiles\":210,\"usedFilePercent\":0.02,\"maxFiles\":1310720,\"mountPoint\":\"/rdsdbdata\",\"total\":20496340,\"usedPercent\":1.85},{\"used\":2172928,\"name\":\"\",\"usedFiles\":75334,\"usedFilePercent\":11.5,\"maxFiles\":655360,\"mountPoint\":\"/\",\"total\":10190104,\"usedPercent\":21.32}],\"processList\":[{\"vss\":760392,\"name\":\"OS processes\",\"tgid\":0,\"parentID\":0,\"memoryUsedPc\":3.67,\"cpuUsedPc\":0.02,\"id\":0,\"rss\":37452,\"vmlimit\":0},{\"vss\":2148212,\"name\":\"RDS processes\",\"tgid\":0,\"parentID\":0,\"memoryUsedPc\":26.49,\"cpuUsedPc\":1.47,\"id\":0,\"rss\":270036,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4745,\"rss\":154532,\"vmlimit\":\"unlimited\"},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.02,\"id\":4748,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4749,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4750,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4751,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4752,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.02,\"id\":4753,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4754,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4755,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4756,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4757,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4758,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.15,\"id\":4759,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.02,\"id\":4760,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.02,\"id\":4761,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4762,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.02,\"id\":4763,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.02,\"id\":4764,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4765,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4766,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4767,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4780,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4782,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4784,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4785,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4786,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4788,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4789,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4790,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4791,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.02,\"id\":4795,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4796,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4797,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4798,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4799,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4814,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4867,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.05,\"id\":5100,\"rss\":154532,\"vmlimit\":0}]}",
		Timestamp:  time,
		ResourceID: map[string]string{"system.aws.arn": "arn:aws:rds::664833354492:db:database-2"},
	}

	assert.Equal(t, expectedLMEvent, logs[0])
}

func TestLambdaLogs(t *testing.T) {
	cloudWatchEvent := events.CloudwatchLogsEvent{
		AWSLogs: events.CloudwatchLogsRawData{
			Data: "H4sIAAAAAAAAAMVdXXMct5X9K1OsfditEjj4/tCbvZZTWxXvJpayeYhcLnQ32ssKJSpDyl7Fpf++57Rkr5kYHXQPlHHJtkgO+9wGDs499wI98+PVq3J/n78rL969KVdPr7747MVn33717Pnzz37z7OrJ1d0Pr8sJ31YpKKetdS4GfPv27rvfnO7evsFPjvmH++NtfjVM+Xg33JfT9/nh7vRO/HB3+jN+dXnt84dTya/wYi21PMp41OH4p3/57Wcvnj1/8Y1xxYxj0EOKk5WDHvQcozWzC3YayzDjEvdvh/vxdPPm4ebu9Zc3tw/ldH/19E9Xv11QP1z8299+9du77+6/vDv9kE9TOf3+sz989uLqmwX+2ffl9QN/48ermwlRGOdtVMYrbYP00ZmkvFTJeKdxj8Y6naRlDFanGJIPPukopUQkDzcYrof8CneuXIpOhSDxYvPkp2HE5Z+/+OzrF4evy1/e4qX/MT09DHGcQ5mzyCkFYbNRIqt5FjGquZSidTDl8N+4Kdze08PHcXn5+ur9k78L2CFO6REZ4sIfk4K3CF4nJa330kmfrAxeemeidtWAXfhlwAj09O7w48urz/PD+D/l/uXVU3whtR+HyScxFKeFVRZ/c1MSWnO25DCnqQjlhfF4/Z9eXrnZWjvYQWjpZ9ylKyKNbhTSxRQHn/Mch+PPBHBHZY/tEMf8+u5Vvn337QNoen3/cLp5/d31NFzf/tW+vHryqcGn8v3NWL69ma5vXj/882ALeUvUf/b9fqTGPx33Z7I+GudvnpCMYym4rPC2ZGHNxMXksyiDCTrP1pjihRQqnsHFRoRPQsVW7L5MbEXtTcRW3M48bIVdoWEYlIlSi1lawIYCJY9gtZ5G75RxMdsRVxBmLw/xl2aI7kTcBN6PiZtge1JxE3BHLm7CrZNRGWnk7CYBAith4xhEMtEKOYxTmfFPKoNwVri0k4zq2A7RkQ+bYHvyYRNwRz5swl3hw1hmXZIRJoxF2GEA7owv7ahnPRoXosQV/H4+2GM7RNdMtQG2b6raANw1V23ArfNBJ++Ds0GoAZnNZvz+kPF7MqOMwGWsyrhC3M8HeWyH6JovNsD2zRcbgLvmiw24K3wYdQxZo+5EeQgqlSTyNFm4oKmgHnVTCFoY8Mru44NMx3aIfnzYBtuRD9uA+/FhG+4KH0pwKsxBOCmRdPKIwmxIUkx5kqrEKUmTaIh30kHpYzNCR3XYgtpTHLbgdtSGLbB1KhhtYo5TEFlp5BvPfJNzhqSUCb4kFRVnXAGU2s2FdoiuZNgA25cNG4C70mED7gofjNR6jvgtMAi1sva83ISLhGKsS8M85DPaLVCwZoSuiaIdtW+eaMftmibaYVeo4OyQptmIMPpR2HnQsKJxFEEZo7MMOcfzqgpzbIfoKA2bYHtKwybgjtKwCXeFD342Y4DXKKXgIm6UYhhzED4HNSqpFGoUYXCRvdsC6tgO0bXrsAG2b9dhA3DXrsMG3BU+xGSQU4KQEYnFzhIXUXoQUjrnwLFofBYGpcnelih42wzRVx/aYTvrQztwX31ox13hQy6hzGkS46y1sEPUYjBWCWNcSeMw+EGPZ/EBvG2G6KsP7bCd9aEduK8+tOOu8GEo8xxVhOfES62MhZvfXgQz2awGN+rBnZMvaHOaIfqayXbYzm6yHbivnWzHXeHDGEdr2dWc/ICkY0Eq7Ywok3UDKhRtRnNWVxI61gzRN1+0w3bOF+3AffNFO26dD1YbXeA0xGyCQ41SjIhFR1HiJINxQbNqQY2y3042I3TNFu2ofZNFO27XXNEOu0aFUasBlSpbVyAVAsgxGzFMbHa7SckQzjpxpI/tEF27UBtg+3ahNgB37UJtwF3hg5+1c1MQIQcLXFjRqAYrQsAlQo5ThOyc03qQx3aIrhtYG2D7bmBtAO66gbUBd4UPQbpSJid4NlTYkrKINjoRlIqluGmIeTyLD+BtM0RffWiH7awP7cB99aEdd4UPOafoDAvU2QjrI0WmOJFRrkxZyeim6RzrYI7NCF2NZDtqXx/ZjtvVRrbDrlBhthEkUuATPAgPb4tkeckwFD8OYwjGntWVRJjNEH3J0A7bmQ3twH3p0I67xgeXdZFBpMmiVHVhFFFPUcw52eTGrN143oYm42yF6MyHZtjefGgG7syHZtw6H5xNoYSYhBvGSdikDc9KDEKZyU0yD6MZzFldSXtsh+h6Nm4DbN+zcRuAu56N24C7woeolZEhC6u5M4qvxICaBFcKk3R2MrM963kCdWxG6Np1aEft23Vox+3adWiHXaFCKlPxLuJVCtcAjURKchB2LlaVqHOM55x1QPHTjPAJTvS3Y/esb9tR+5a37bhdq9t22BUa5nGclVRiHFwQdtIStifPQhtj5kniGkGedZpfH9shuha3G2D7FrcbgLsWtxtwV/gwjNroBLNTchFW6iSGjEJIp2JjKLHMozpHlsyxGaGrf21H7Wtf23G7utd22BUqlCGM0hrhgyRsUCLleRJGhWG0I7zQ5M566gyJtBmir11ph+3sV9qB+xqWdtw6H7xXxY5pErMvGvpSRjHoOIlBRh7jGmOJ7pyD2/LYjPAJHEs7dk/H0o7a17G043Z1LO2wKzTExc2svQhDTpA2PpQyjFGEpGY1jGmMppxVUyPMZohPQcR28K5MbIftTMV24L5cbMddISMfmAOyGMMshXXZi8ElK9JQwqyjUdx/Psc+I85miL58aIftzId24L58aMet8yGoqHUyTphsvbB+UtC2kIXxgx58zOMMT3bOWQJ5bIfoyocNsH35sAG4Kx824K7wQWcnqSp8sx0U5hp8KtKAR8kbtg2LNud4aJmO7RBdjyFugO17DHEDcNdjiBtw63yIbnBxykHIUiySjkeNrvCl1TFq46xOVp9TXutjM0LXZks7at9eSztu11ZLO+wKFfIwjfgjZrtsKEwo0UNIgFXTrAetDC59zolUdWyH6Fpeb4DtW15vAO5aXm/AXePD6OYJCWXyGf4jewNOaW4q6BlwRft4XrsFtG2G6CsO7bCd1aEduK88tOOu8GGaVYjKiCgny2POVkQ/BRGNBq7K0sKcntFuAW1bEfqqQzNqZ3Foxu2rDc2wdSokqdzsZi3YqxN2NEbkKJFzXMxuGrKx+bxjBPrYDtFVGjbA9pWGDcBdpWED7gofYpwHPXpR+Grr+L5NWjlRxjxl5JyJT1DulwaY3WaErjVFO2rfkqIdt2tF0Q67QoXJQ2D4Oqcz+JSyyBAVMQ45xyHreUDxesa765CyzRB9paEdtrM0tAP3lYZ23DofspbTGOWM/JL4cNTAKjUoYdRoZUrDOPnzTqTqYztE9+74JvCeZNwA25eMG4C7knED7goZjcLLxyK008AdoW2DK7OY5ZCl8kOI5az3coaGtkN0zVQbYPumqg3AXXPVBtwVPliTrHJOTEVC4ZJMAnnOihy8HqU2Qx7OetMGxtkM0ZcP7bCd+dAO3JcP7bgrfPBzTGUuoBIs8Id3fpiZ9mLJIasQeLjtHPOiju0QXYvcDbB9q9wNwF3L3A24K3yIsxt0DsJl1EU2ZyWyyTNqJZ9nPaUyWX3Obip52wzRVx/aYTvrQztwX31ox13hw2Ssm7QS02BgQka+x6DxkxhHbYdg/QyynfU4jTy2Q3TdTd0A23c3dQNw193UDbh1PgwqR8f3pJaF71+uZy9Sskl4ZczkkomjVGdtmehjO0TX+mIDbN/6YgNw1/piA+4KH2KExqBIKZMvwobsRAa1xDDAhnipIh3qOfqgju0QXf3DBti+/mEDcFf/sAG3zocxxSyzDWJUihcBbswWTkQP0c0jZAbfOedJfXVsh+jKhw2wffmwAbgrHzbgrvEBl+FbPBQ1oUiRMCHJyEH4qGcZSjB8A4Az+ACf0w7R1U9ugO3rJzcAd/WTG3BX+DD64opKgnswwsYEEzJGXM6WsQQV53k6rzkqj+0Qn+Do8AbwnmZ2A2xfM7sBuKuZ3YBbJ+Mk8TJVsjAjH9dSfKDYR1xThUkHjTrJn3fewxzbIbo+a7UBtu/DVhuAuz5ttQF3hQ/ZRCcnJ2JypNI8iSHC9HhtgrIlmwnl8znNMMTZDNGXD+2wnfnQDtyXD+24a3wYZ8fP2ZDRGojMkEXmxuAYk4I9zmqyZz1uxTAbETqzoRW1NxlacTtzoRV2hQqjVymMKI5SKdwZnkXSqJCSjH7kcywyqrOeKkCYzRB9ydAO25kN7cB96dCOu8aHPKo5TmIcRvgPzXfUDxrXHAcdzGj9EMeznjpinK0QnfnQDNubD83AnfnQjLvCh4JLp2kUZojwH0NC5lGO+2+mKOtVmeZwzjvSyWMzwieoatqxexY17ah9a5p23K4lTTtsnYZFmlT8nEBWPlHp5wGyJicxecXPirLTZKaz0pQ6tkN0bb9tgO3bftsA3LX9tgF3hQ96ct4aJQBrYX3CKJLVVqjBz9M4IKJy1htl6mMzQtfNmnbUvns17bhdt2raYVeoMIQ8DfxoWW34RIRJYvDSiMlOyHI+RDmed9JDHtshPkGO2gDeM0ltgO2bpTYAd01TG3DrZJxnG8I4TvDgbuZb+Q0i6jCJNBc4LidLUsMZuiTTsRmh66ZAO2rfPYF23K5bAu2wFSq8B+4fy/D7t+X0DlP9I3suD/m+POCLtoiW0Oeb24dyuidZcInx7vbtq9fLFR7f7fLauzfllB9u7j68YLx7/ZBvXt8vP/o+374t9x8o97tTeZNP5fBwdzjdPeSHcri9++4AoHK9xP0Y5xGHfgWm/OVvAFRQznzznmsBw3Li/SqXonYh6CB/+oey/Xr6+DOnfHKPfkbxXC5/Kri/gqsvKL/4iiN6f7dcHn+7O03ltPzCF8+e//vy2l/cw8srzsbtzasbvtwtCPmWF3k4vS0/v/bj+Pz9Hf8No5fv/cr4P2LC8p1fyQUfN4venhD8h9h//jV8ychupp+Cfv/xot/eF9z2/YeXD2/HP5eH+48v/u509/bNt8O7j7/CX3893r6dyrfl1ZsHfnvOt/fl/fv3L19jFH7E1a+eXhnk3KiMV9pJlSQSMDJvwh+TgrfKS52UtN5LJ32yMnjpnYnaXz35/2h/mroQZJIuPPlpRHD5L+5+eH17lyfc7sEuzLrvgB4uih4vip4uiW7kRdHVDnTXDV1f9N7NRe/dXvTe3UXRL6p1Zo/W9Zv3i2qd2aN13e7dXlTr7B6t64d+Ua2ze7SuH/pFtc5eVOvsRbXOXlTr7EW1zl7U17mLap27qNa5PVrXjXXuolrnLqp17qJa5y6qde6iWuf2aF0/9Itqnb+o1vmLap2/qK/zF9U6f1Gt83u0rtuK8xfVOn/Rfp2/qK/zF9W6cFGtCxfVunBRrQsX1bqwR+u6qU24qK8LF9W6cFGtCxfVunBRrYsX1bp4Ua2LF9W6eFGtixf1dfGiWhcvqnWxu9YlG5JWhsMqbVKo1WzQRgUdXEJcCj+1KjkfTJCxqnX6MfpHxMP0MYoyHW5eH+S1lsFadaiFgiLZaNyxxzBgHFSK+KbG3z2PMGipmOhctBLjU138OthfhvK7092ILzgO928KAlHeXuM2k1GHrz4//nooRkqDOwZitD7iP4qBKeddwMxIz+UfEqZE22BTVQV11E2jgjtXtVEx0pqorLOIxqpgnEkp4tIS1RXmy0qHFOQDhibZVE8HuKP1UQnqmh99avzaqAQdpXNKJa+1N8ZhKBBDQPzaYl5gglDvJoywNHWuGGWaRsVFbWRtVJTCxTwmOxowWDkDSG5bKuBzejBHDjPjtI86VcXSmEcT9OVy3ufAs0KH13cPjIOxVSLQVqagNZZJhB4Hvu2vxJRYJzVGZJkmzBN+mlK9+YMBbxqMwLefrw4GVigC0FheuGbUYCz5G6W2ESriUIwiEPwI91tXEGNtUygxGV8PxUmsYoWVg6WipTNKUsqgKIhOKgSmTJIYGMxOvQ0MLfwHa5hagjVsV9iqHNCj9tARkMB7B12JLmGCXHTW2+BBVgmFQaj1nAKS76cIBB0kjJgEp5U0VinLDxII1uH+ncdP2JZEzYjpqg+GVy3zYjDYPtTmRSslg+T/sf4c/sfPMrDUU3zl+bbloKmHnkHmVigS3Pq8qHjNtJXcyrzoELBuEGtYRgJ8TT5IDMmyZhGZR5JYJBe8rS4cCF/TqGDWQ1VFoIo6QEsgYEx4QXq10BfjwDEBPQwIIyFoeuUYkFVuN0Ug7wGrEkggKtYtNNaApiipuIyRA7meE14KhtQpgsTQNBggvdLVwcDdWuoYlkoCRSAauHtvQgB1NATe4UX4L0yJq9ebYFVTKJjj4KuhWOYST0m3UEgoBUYC/gfzg2yjkPLAFqnhnZAH6qNizpgX+CBcBuYMOVFG2qO0pD9vMCRQDU918+COlfVGI1b7+npx4Ro5Kqiwsl6sU0ktkm4QANKdwSxC2rFykOcgHsj8mjEqELk+GLEp62K4oZm1ecGUIcPxmCjI4INHLtD0HtA5ZLsAZ8ZkGPA/51N9vSS7e15gOeByokI2gwIw9Rl4D4nbh84iI2CEIK0mYWR0vS3h2sQDZIf9rg0G5g0xBBAAw07HAWJymSSnJMYiYAnBIimIB1xIdTDc46W7aTCwQEBBzUwGx6FBkYilA+NKj8VyIjj8UNEXWlWt2WDomgYjBSzL2mBsCqW6XhpDAQlUXTy8QsEE0bCYEgv344DLBUyeJPp5+EYkMhhorKvqenHujHmBd4VOwP8jmYAbEjUdYuYCBWEchB3KDvJAa129e+Ue5/1tEcAZI19oeHKYQ84DwkFy0/wmBsQoOhMH34hyq5pjYW2bpsMgq1Q1I9ABIpvBcUnKeOCocz4sn5tf3DOwJJJPMnXngTS8ezBQutJtECjapaylH4xIrVjBqKwT35sfeTey1KoWc0g9TYMBYZT1wViKNxRREHESFXIaIJ0YFNw9LI8zFHmLvEPXXA3FpH+QVhyKygCVXEkrkaf+DeU08sYNRTPAGRup+a4GqH4jLgCa0IVUKeIfl7iuMiqoylAL1UYFdZNnmmUhC811CUICaiY+j6EoK/gDPaecunrnC36tZYJQt0KTq6Ewl1gKDK66mFINV8j2B4pJRaeOJMsyRkPqq1wBs3azNVkQkj0YOCRP9wXDg4gxMBbsAGNgB/B9OjNUd/UI9nufhHpaO1LU057SCnvmNwZkqFjwK+yNBCpsNccG1TYdHhmhVrZYiVoCQgaz5akWCM1APFFKoZq0iBFcBY3T0h+RVfGA3O0dDKYMLDdAwLf4gLu2ML9usabI8RRWlDOIBVYw1MUjBNOyTDDVKBVrg0ErgVyCBOuMTR7uR1LSUF0aeKGQUBFDbBEMijpZZ0Zs0rGAqZe1ppRlxYQZYb0AJ2xjhPOi0sMhIPdgjjBbNMlgC8S/HkpT2RKoDPVQWLNhZQYPswZyQj0C25gYIjoxjBAKOFTfS7TVrmGU+ymCi8OGOpZqWBecFuQ3DYFFneDgiKEbXEwoYOALqjYsarM/AlotVkrILGzToq4AvGZHSLPAjnyOBPULpmulc4oBapoOUK4+HdpDtR3yKMYbFIVwIOtJNqfY3HUwiuBrSoqNunovPT5uAG0bDOgSsgYoA/nC2jU0oTBES7Wg2HnB2uXMwAvWz0FG19S7DZThWjaxcMLIHQjIaDYJaXdQo4AvkTwNXEfIwH5pctfFA2TePRgARtGvqY4J643lCTKxggkFLTEaND8JzoxCVs8mMTQVjxGiXG3vY9Rxl4FddO4pIJUHOFwfyRUsXbxIUUKQuAIMan0wwn5m4N9ASw6Rwu1iuaCcdezvK2uRaJDQLZtP4Klc2VmNab9UIFnAiNMRB+ZWqjkcIGOALWZmjQlJBoIVYRWrtQlsUdN0oPDxtYoNd7y0P21aOoGB1Tz4SqcOV4xBoAONXDu0vtXUCi/fFIrjdauhKPaJkUqQyUAGJFw6YR/48J2DjEl2NjDskUVbVTOSairqUfZBieuhsE+O1QENc0tfMC4NSczQElxizZYkXchKhyO19eMQCga9GopmW8fzcmzbYggYHdiq2cmGF2Ijm81r+oEqVyC2baE4a+ujYtjYT54tMEzOUmJHx5JIRzIa7hDrx+AFHL9qKGdkOFhyTjBSKEQc9w1ywoVrVvOWbQMQlruEoAykrU4R0+R9cBnYh+pgYHk4CimMOsQcSxmTAhvrbLJae0SjDaxIYiGl6hSxZwxGYo3EHpzm1RR34Xj/3JLCsqYj0WzaqmVTrhpBW7MFYmNcrA7GllDqg/E4lJo9jmxBVWsFFAP0XYEdBcirTnSnhkslOc36Ae4a682xZZpWRqXJBGFUUt0eOywMyUpJM53A/YEWLKvgiKMBV5jxkHsSTFKoH9FF/bubIlgSxi77+drCd/gP5wsMOz7s5X/oSyLx0J7UO+hI1y2Dwf5Jqq4XJhKsW8VNHkjXspnBAgauUGIxJ1SVIAidK6auLh5xvwlyEHHJXUlkFhmlXSp4fFNj3aKQWnYtYZYjyVp/Hgsqtz8ClKiYCpivCCnnjr1cCpTAjQTYDY4O5go2DWV/nRCPG+e1ZZJYE1WTLQxwYi8uATTxyIdkukXWBRM8xFxxiwG01PBG9Q56Sm3MsLQ0tVC8RGYjMJyo1AYTgIIhcCZ4BIUuzMG/Y7bgB2unO1HY7W+2gHkK0gUB835ZisgnLKoVd3scnKJj3WBhEHX9nAcjaLLHkCEQvzoYEAaNalkhszsH28UPVsIkcK9FMrVqmNdEm4iiodJBRyhtfix5Xd8sJjsc+5IgEqScjp30xXpJ3JPEUgFVWF2z1S/ro9LWAEIooXrOA2Vt5OLxWDsek6Mp4M5Se3ns1JnlxANKKR6LWaGI2Z9s4a5wl9wlZomEL2CNNXd92AFkVxN1BP5qaIVUfTDc7j0vG3ieH/Acf4eUwXIscK+c+/hsZBvYZ2gLW5O1JhgieFxF/vp0qGtJ51JN9wFzDlJYElEtUwNLJlFOQg8c91vYYqfQmvrxDobS4EkRCvx3qpKUFYtfpIFmS7LtAXZAviJsGXQVXoybop5JrpZWuFHXsHQRCup3WyVpgPNM4AC8uGRtibExPFMCzxPg88BWVDHGap6hq72vA0NpcB4MRUNGqqEgfcB/LicZkFvZ5KfWw/6gemAJFXj8w8G6hupDIgylYekyFIe8UQ8FQkoLjBy7VJNQMc1lynzPHX6eEoMzAXHCioqEhlp3CQU3WQ0lslXMNsxyWgOrxMAOeYlIoPGse5ctXDZHUs2lsoPYYJgZClShKvM8bcmWC9iLRcTPWkS5zeM50DE4JKR/HpKUzDqq5lIZShtXeNCnSlu2PdgsxQwFbnQrLm1W6qh9kXAWh8SDMAZCU3uAEKGktlAscmzVlEQ6MpMMaicMA4bAfjg/xd4u1njkmGKRIfUwUddDaSi7GYrB8HYJZWVUdm8fM7+wHcV9Ocg9dygDLRsi4MFKxcOxrCDYYLd1OVFqf5sq0pTi/j3WBM8WQj4Uxh1+cdnqRUaySIA8rMyPDK5H0NCQwXSwPqgu3Q2hqLqKNIeCi9SZwYMDLrA1bBZ/RsWnOWL/0hCEtYxOIfGkXTUU3eDQGAqkqKqtcfGmMnGvOUoDaBt4dcW9F6RBSQPp3bITUntChaH4f1xPIBTuatQniH0d77gTpHiSPEIseN6UZ6ODW7ozDnxNmDdVOzameDKxaVQiS49aKEmy8c+jg3TRhnaeJ+zgXlFKYOUiEaPK5MkdXX2vGYbidy8cyqldqk3HQ1nWMtGiqIOEQlP0Ujs5ZhseiKgvHJua5gU6EKr+aEsotTej+LtQqvOCKrq+J8SCVIbEyhN4PBtGQmGl8cy6CXE5FWu4qYsZrC+cliaRuqZlqO+h8uiX5vshevozHqvjCR1S1VPSLQ9uIT0EHgiptc4YimucoFA9Ks3jN4mnYFj5I4DEvRoemE5IOY5WEXaE53Yg5K72rBND2Z9oeEqAzpXHkHgwGDFIPtWABY0qh20ixYPnyGU8116NwK8/cGH9Nep62L36YRQMhKc1s+wrqA/LWEb23CH3lNqo4BIt+3s8cleNJOxOeUi0ltmWjhBOyHGrkE13fo9PfdAOOYliAzQJtfcnYwRNK5eHzqsrF/CLA/TcIHQ8Ic2uKrOdWo7tsIFmSB3PVlJdUePutgj7DagTNN8fSLN1yINrcdkuA10MTIemjHjPh2RWiJGaki5Gur6F6rijmJYHfxTLGsuuKswz1guqGO6+8sknx6cvuHFVD2X3KuHWpF9OyTsWDh7MpDuLXC8aVRUtQWQFqpFeau90wQiahJTjXE3720KpS8b+s3RQRajmh+1kWB4ejebWC+WLyu55vACWA4vZsYdUi0Dv96ZuOQekg17OnGiYC+i1Qh3FB6L4EDgPTlqs1hS4bOsRNCk4tM1XH81yy+ES7oAlHjZXSKRsRbFikktnU3pNzwiqcsLqoeznpuJDNcjmioU9z85ZNrvZ3XNYI4nbhGzJsH9maidN2XrdPx00Wjxui/wNdWCHEOoAUqCu93y/ocCDMfjaLUf66xGcMQYO8gHJisvFaCrwX55c5PkoHujDBPEYKmL0K4Swu10fLgA/gXHnMyth2axeuoV8/ggvcsiwfC8Ufk9Xz4ghAndGBAmmzlvHTQbDzo7hqU3MQgAgt3wCzyNHyzadrqYM7XcfGuR5EuRpxe4skiaPc6I4sdzv4OkCw/1ByBMfnOSJymoEYXebdCmaaWFwgyAkN75hGBTYoPjEhOaBcEwOMjwoUnvfO8Xm9v4I4CUXW2n53BkS15KkFGvX9OGJ9bA8JICloGpbK4gg7W5WOz6fSUnyXnkYbom5gADx6cgADkrDkhBWT2uWsNUxMPKMCHhiOPH5WB7vAXRiYzouBxcczRW3yCGNlOnaWQ7Fo/K7I+ARvLjsC7AYpQQbwC3boLAJhk9bsf/HM4K+3tYy+w8rYLhBBR4U4O4Rn2Hi+UTLip5b1MiR3P9bOm08d1SNYP/DqcvhGR5i4akZyUPmqK/5mFCQ0bHAdWzlJ+oFoqjqgbFnRBCWc2eJR5HZq9fcT+NeG5QKixBrgI1QHhjhBz7UI9iviUxLTAEQJZ7E5G4nd14dm/SaJzJ57guOHVPkVniwf+OGz2wve+2UQiw+aAN0CV/A/dulyuPiQOHMk7u1Ezw8F7V/LXDnP6KC4ClDJ5fD7Z7PAfMAgOVjfDzv5ZctT7vCg7B/LdjlYVfukIDrkGIPmY7LQUzPQ7lh2e7VbMjyJGA1grR79x8X5mPhkqcOWbgExccusDIsB4HFFA9ocnsxcIuiHsEZPNgSQV0PHkewfLjFAa96uH96QG5J+poPsnh1WD4x4f54f1gexoUALYX34V/5QQiHnz/F4d+eHO6G+3L6Pj/cnd5dP+eFfnxx95Bvny2//5TbSPHJYfnW5+8eyv1TRdsNV/nk8MXbD59C8dXN7e0NfsB1evjj3enP5YSvnhy+pJt+inz46598sG1E6rx8PCLP/vOLw9flL2/xwv+Ynh6GOM6hzFnklAI/UUaJrOZZgOtzKQXVoukxX/Uc9ji6r5/97r++frE5wIefhvrpgc9fXjt9eHX/8uFzDDyKlF/80Eq5/OSr8grzeXh+89cCXkgZMfv4bv7fw8ef/OG+ABwJYvkBR+Cb9/8HR5LhhLbrAAA=",
		},
	}

	logs := parseCloudWatchLogs(cloudWatchEvent)

	time := time.Unix(0, 1598517709043*1000000)
	expectedLMEvent := ingest.Log{
		Message:    "START RequestId: b8cf7efa-a997-4a31-a1ff-881feee2273e Version: $LATEST\n",
		Timestamp:  time,
		ResourceID: map[string]string{"system.aws.arn": "arn:aws:lambda::197152445587:function:observatory-worker"},
	}

	assert.Equal(t, expectedLMEvent, logs[0])
}

func TestEC2FlowLogs(t *testing.T) {
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
	}

	assert.Equal(t, expectedLMEvent, logs[0])
}

func TestNATFlowLogs(t *testing.T) {
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
	}

	assert.Equal(t, expectedLMEvent, logs[0])
}

func TestParseCloudfrontlogs(t *testing.T) {
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
		EventTime: time,
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
	}

	assert.Equal(t, expectedlmEvent, lmEvents[0])
}

func TestParseCloudtrailLogs(t *testing.T) {
	cloudWatchEvent := events.CloudwatchLogsEvent{
		AWSLogs: events.CloudwatchLogsRawData{
			Data: "H4sIAAAAAAAAAO2da5Ncx42m/4qCX9do5QVAIvmNQ2k8HN28JKUVNXIokDdNz5JsbXfTsseh/77Ibkq+FIchuuoEq8spKaJLda516hw89SbyBf5870W/utLv+9M//dDv3b/30YOnD7777OMnTx789uN7v7l38ePLfmlv+5w8BUQiSfb284vvf3t58eoHW/Kh/nj1YX1+8apdX+r589uFT64vu774u+2+ezjXejrX+k5/gJcXl9f/2fXqGrxtdPWqXNXL8x+uzy9e/uv58+t+eXXv/n/c03u/v9nhx3/oL6/nO3++d95sv5FdxiwpO4/Ods7IKVCwt0VSypmi8xwy+ogpsaC9muuLHen63D7xtb6wk/fshcg2o+TxNz9fCdv9n7+91+cRv7LTsBP69t79b+/5Myff3vvNt/deXfXLR82Wnl//yZbYutd27W7WeXB19epFb48vnvebVX+4PH9Zz3/Q54/a7fLHXzx4Er/55umTJ//+zcPIDx9/+Q3e//Sz2+3mZk/sHG6OaFvr5e2R7e99u8r3r66v7t//6yt6X2+PB5e25YefXnx/Xj+7eHl+fXH5nff5w/9xv7VevHp5/fqc/nqHPy+2lT/pf/r5pJ88+uWkv/qXb755/Ojz2xWvbvf58OLldf/j9e2leP3eIztwv/y7q/Nul+WNl+BcX/zdJXjjR/81H3N+jZ/ri9tT293BT7bKj708av/aW7/UeV9+pNc6P9JcpNfXl+fl1XW/uv2QL4Y+eGX3s90VVa/77TGHPr+6/cTVnofXu7g9YHDBg0Pw/NTxfRfv+/CNHfSnue+bW+/p+Ys3rRnSfUff3OzzZrUnF68u6+2KvYYzfaH/ffHSrtRZvXjxl7V++Zwf9fmUlf7opT0CL2t/cq3Xr65uL9ePV4/79z/f7X/7gN5+3TeHevS7B61d2rd8s1rEsxD8mXdn8ZeL+uB7O+LtTn68gqv2f+G/9A/6oa3lz7KXDz49f/nqjx/imbf/cgSP+Sz6ZKf+3y/D2R+Fv2P84Isf+st//+gTewn/cn793ZN++Yd++d1Xn31o+7Bj/a8Enz598sHtfm/e+cA+Zru4/PDBzQWYn/67Ry/r2e0luLy8uHx40W4vwcPn53Z+Z4/7/3tlceDT8xfn1x//sXb7lttf1v7sNhTc3ra3a37wfK76QX+97u2uL2+X/U4v7QrPoHV7N5y/vrxXT/rr5+L8ur+YC//jr5a+vjPPwbWI1HtNqYo6KXa9f/r9vBPGTSR8vZP5hj03z1+1/uD585+/wbnTm/vsp5vTufrh4uVV//h5fzEDpi17+er587+c6KOPbu+pWELXkKGk4gExD1ASBBdpoNBwVMdf7p7XGyUXNWqLEFtsgKV5EI4JIucwhFwXSq8virYvXj6fwfH68lX/5Y7+JUr+ePXgh/OH+vz5zeov9KVd6nm+N1H+77Z6aE/M9xeXf7rZ8rNfVn19IIsi89t88JZH/Sd7lN9AjZCMFMEHJCfZk7cl3jGJ+OBjzuyjs398IuPL/0wNWtT4FdR49skTeva7//P5osZBqeHzfS93kRo0g382HthfRwsc/zg4NOWKkVrPvocRaWtwcLJgbxERSk4MOLCBUDaO+OTaQOwV8w44QnQxZe0gBSNgDwW05AaFh+OEwp3dSYKDFzj2A8fXX3797NG/ffHpAsdhwSH3iX41OPh4wDHlhucbcKBf3NiDG8F3h9RCkxyUxtbcMDjlUkVMZmQBFFXIvnbIVEK2aIo02g43mqM4kmbISsaN4QkkYYDqW6g+ZFc5nyQ30uLGftx4+ujxV+Exr2GqgwuO4H4tN26kyZFwg8KZt//4LIZFjX+cGr3Q6KM2xOEkxLY1NRQ7q3MFQogKmEsBoxVDYjXJUzRFxB1qlBhTCM3ByNJgnqpJFMdAXqqpTW4lj5OkxluSG4saK7lxR5IbR0SNldy4m8kNnxvjUCjiPGCPDnJrFdi57Btni65hV2vUVFBbAuNLB1QmUNYETTR054dJkHaS1MiLGvtRg9I3n3zx1VdfLGockBp0H/F+xLtKjRn9KZ8hL27soTb8sB/63eJtR2YKW3OjYM3DdQeOjAPI3QjSsJiO4BGkRqU3qA2TQ5qG3YscydRG5gIl8jDqBZMaDccoJ5nb8G5xY3HjznPj2HIbixv7c4ND945y7bk5X+cg8LbcqNSMGK5DFD8AW0WQogUSE0vh6nLqu9yorlFj+8ayFECPCYwjCJTVMMROKJeT5IZf3FijVEfHjXcdpToybqxRqgNQY7RaDRiaR21Rt6ZGHM1j12FhHyugugElYTPh0H1zqRaU3Sm43vse5/wp5hRto5KhCJNBpzhXhveu+pOkRljU2I8ajM8+/xo/5UWN95cRPypq0JnndDZn4npZ3Ngju2FhCgl7wUbSSt2aG2gE0FE6FKczvV0jqAYCCTFapGw6otvhhvjmWXqExtloMSRCtmALVV0VRJd74JPkRlzc2FNtfPLFV/To6SeLG++RG3hE3JijVOGM41lYo1R7zcB1USUmTtixivDmc6lq86H2Ai3nAijcQFJHYIuGxXeqg3azG5m5ZAkMlGIw2NAAKUWhVC84h7BuTvwEuYGLG/txY2U3jiC7sbLip8cNKaOkVnlgdTkLbs2NrFVKR4FUXAesmkBoFAieJIjaubRdx58OKczdQ2lzNhVF44bvHXIbXAIpC54mN5ZVfE9uPHv48PGjb367shuH1xv+13LjmJwbN9yYVvFo7FjmjT3AUZ3E0YKvyQkP3dzy18ZI2U+XuOvTKp48qEV/kKYGjuR6CbIDjso9DltgCgPJVMoc5xISUx2c4xgmN15vdGrgWFbxPcERv37GX3/89f9e4FjgmKuRgQNnfRFnf/MCxz8ODipt1BKiMkfMI26eGS/RJW4EdRrGMY0OkuZ8KouFoccYnNMdcASKxQsbOIwu02Bu4MhEQOplKFsM7niS4Fhe8X0VxypO9f6LUx0XOFZxqsOkxodgb8qolHuQzY1/3aXaanczu1EAs2ugKSTQ2kYy+eOd0A44SHyhoREMFLZlzcEUR0Sw95WamPAsepLgWHbxBY67D45V1fAEwcGRLZxFlwu75mLfvKphJG6CDTylaeKbhXGrZ6jFxepiw9J2FYcr3fWeK0SiahvNUoghJRguh1FGnXH3JMGxHON7gmPNxT2COVVHBI41F/dA3ChhOFckeYlcmmxeDVdm34k2BmSZKY4+M9xjOOgWDJ1nphF3uYFZaRRCUB4B0NUGU4AAF84WjFMc6SRz42E5xvfkxqpq+P65cVQpjlXV8CCJ8Z59GMzdArb2sHmdER09kH0MCNKTAUAcZHYElJJPNQfF2HcT4xLHqLlBDKqA3r6bwoWgupEZCw/DzUlSY/nF96TGqqF+BDXUj4gaq4b6gbgRCakkbz/2XdfaN3eMJ4vyESmaXJAOWOKAEjrajabOtRDFy67aiE2HF6cgOIe26hylqpWht9KTKgdq7SS5sRzje3JjOTiWg2M5ODbgRu4cynAUVJwPM9Jsy42B7LumDL7NqVFFGDJOR0bvvVC1kBjqDje4K0sgBQpjdgiMDJp6hphsk+GTseMke2+E5Rhf3Lj73EiLGyfHDQ0WvVAHUhvk8uaOcWkeZ1wEkT6z4qKzyAiBeHKBfEPfd8ep/FAjhwZoXWaFKvIgrdrmYbjsYyuRw0lyYznGV1b8+LixKlQtbvhCzWKZcz1zQ93cv5G5B8kBoQfMgJUmQWoAZhM7FCuGvOsYj4lMWUxV4rwDJGYorkcI6DOVVpNvJzkNNyzH+J7cWBWqFjdWhaotZuFqiTlGDDH1XuLm41RFMKeRBDiEAjhUQIuSYcQCqmocEnf7NuXWO2odwFLGrKNu+oQ5AmvPyr6GXuNJcmMZxpfeuPvcWLNwT48bPtkP/JYoOE0lb1/ZMLXW4lCGceP9y7VCqVLBGU+GtJ5q2NUbiM2J4QV6xDlOJRmEknHDo/SKhQwXJ8mN5RffV2+s/hvvv//GMXFj9d84iFkch1OL1sVRcqG7zc3iTK6PIoBzdhR2riBuONCoPPuK4/C79XBjVa2tI8QS5tBWbKAtdQjcMFILntxJdm0Kyyy+JzXSwy+//OwZrVGqg8/CxV9tFvfxeKgxZYbIWXCmOvKyb+xl+kvCPuVK6muLm8uNHJ3POlMTRWen8BqhdKzAMYYROJeU/A44mjFFQkmmTLoH5EqQQ1VoPmot3Zf8Opd+auBYZvE9wbHsG+/fvuGPTG4s+8YhyhomlxuLhDhrjeDmZvHQUlTvA+RYkwmObAiYgsNx0GZ6QsZr7fA3tj/TQ8MQAaNWkypOEWRgh95U2TfvaznJhn9xmcWX4DhSbryD4DgibizBcTDfn+tOYi65ekK3/UgVBtFELYOJG2faoXVQrAGotGwUsUjqdgupS28hk2tQohuAownkRBWCdl8js+HjJAupx+UXX/mNYwTHO+U3jgkcK79xEGqEij0JtzjqEIx586z46HZbOYbIJQA2JCimdaCwUBO1iKm7bnGvbvTUB1BHBAx+1jTMAXzt1LwrLZWTdP3F5RZfcuMYqfFucuOI+sQuuXGw8lSYR5WeWUrT2DYfp4q1pCbFA3kmQMQA2mKHFDj0MLQ13ZUbvpfIPjUI1RiD4jKU3h3kohiTnX/mk5yGG5ddfMmNYwTHu8mNIwLHkhsHoUYfrgir8zGVHnraXG7YBZ8FRaBLLIDGfsgpFUg+ey3BO+m8Qw2eTWQHReh5TsIdXA0YxSRLUReQc7IzP0lqLLP4osaixqLGsVEjWgALYaBY6E2JNy+FmzX7UFMAbrPVHzUCRY6gxdvNHz2VtEsN8tFjwwY8Wp0+wQEyxoDWstYyYhtykqVw47KKr0GqY6TGGqT6pwdHi5hD8SWXRsx581q44jWRbx5UpnJog+Yk3AE+NGdxtbnqd8FhQkioExo4btp12Oa5Zg+jplZoULWNTxIcyyu+JzhWq7+tvOK/vtXfsYFjtfo7CDgkk9LwgxrnoJunxX0NRUZIUKPrgLUHyDffRfPSLOwj0xta/UkSpkQTNAaOZMixANuheeXu1XCTTrLVX1xm8T3B8YQ/i48pPF7gOGgx3Jjvx7s8TmX8wDVQtU8tXNTgyugyKoYeNk9vqMRRdAjkUQug42IECNGkA4daXLMn5Q3mjeazqQwEqoqGjZEhB5kGkFKU6kyL40liY7nFFzbuPjbCwsapYaOOaNGro6HDwlbum6uNOHqdI1S90PSKM4NWGVAbp+yj51beMJcqpqihJtAwG4s7L1AMdZBbqyqaZ0L9JLGxvOJ7YmO13nj/rTeODRur9cYhzBuDO836tDWlbCFp89lU3bUU0YOF0T694g0kdgTtmAKP4XLpO9wIY0QTGbNVB+uUGwRqDxV4xTpciIFjOkVu4PKK78mNNZtqzaZas6kOX5mqiu+UwjRUxExta2pUKaXNLLi2GsDCZYRcA0Gz01CWXjDuNvoTV6L2UIFanhN3i22epMOYEgXtR3lMp0mNZRRfSfFjpMZKiv/Tg0NScq4ylTS9fxQ3L00VpMQxExWUMmAgD0W7gIu5CbXixLcdcAz7toLzA3qpplGYB5SUEhRxPQ0LylpPcjYVLq/4GqY6PnC86zDVEYFjDVMdTHBgzYE4iq+h5s0nU3WM3BI2kDY7/U3tUCxyQjKgYOKMTna5ETVEX2K2jdg2atGDuOShW2gdjITsx0lyY1nFV1b8GLmxJlP9k2MjV+/n0JG0SjLy5o3+MMVaUxbIYRZDj6OAxFYgUAo0mqvltXL4a2wghxBjjeA4krHGO8g69zGlhhEvpnKSDWJxecX3xMaqoH4EFdSPqPXGqqB+sGGqxpXTbMHKTt3mw1Q820NZ3IOMxbhRfTK5oRMeiVrIHFLdzW/ENjSXYnRJtr79CreNukQwieSzlobUTrKkIS63+MpvHCU3Vn7jnx0cGjsjBZE+6wOOzRPjeVCxD5EhtjynR3WCTD4CBszDxeAFd3v9caU6HCIMNzuL6wSH0QfSaD1ioVTGaeY3llt8T3B89ckXX9Gjp6vMyHvsLH50giOccTwLK7+xDzeGdvu3pkyxaXObd96QhNUOw8B+5NtWf5kkgDKLx5pNeOx2FtfiRLQQhNIUcODrcS5KreSAKaE/SfsGLrP4vtxY03Df/zTcI+PGmoZ7gALqnEOgrkTRQtb2XvFiAifMESflAEiDDRjKULUndNiCjl2vuE+hxJYUuosM2KhDwZxumswGNrnk8ST7NeHyiq+s+PFR412z4sdHjZUV31dsDE7i6sijtBrz2BobIwsnIiOADAcYiaBQMBZkC5Y1OcWedrCRZwdAlAHF9TpHtgZki7gQIxKT4nSCnCQ2lld8T2wwPvv8a/yUFzbWINUNNujMc5rY8F4WN/aoaEgYCqpmP8sCts1nU4n0VnsQoMwJsJpy0BAqjJxGdg3tSXlDDfWcXCRWGKnyLLzOIGUWtfKSU/Otx1ZPkRu0vOKLG4sbixvHx43cYs+dOVR7wWPzPn880ki5mVbArsaNQLPMVIHgKZfSOIa6yw07y9EbByjJd0DfKmhybP/LFVPV5vtJusVpucUXNxY3FjeOjxvctc5w1goZP9zmtak4+zKr7kIZzhkCegAdzk+AhVJ95z52K6iHkoaF1gFKs6ZhTxUES4bs8nTGhalYTpIbyyy+JzeWe2O5N5Z7Ywtu5GjASJhi4HrTwGJbbpTi7IDjpl5IAywzaVFLguy9Sm4lOdqdhNsCM5k2Ads0AkbDjhhtQFKtYlFWBU+y1x8ts/hKix8fN941Le6OixsrLX6IEuqk3kKR/fiPjf3m6Y1Uou+OBFK66RFe5nTa0kGbuNmlNjDtmsXZO6p5ztmt3jTKsFcau0CIVYYXC8j5DskNct5eUGLMaPdwRCeBnaecUZgR0Vji2M99vCW9wQsby/R3N0x/R1RCfZn+DqY3inROA3m05qL6zafh6gglUoDkhoHDNQeq9l24XLh3Cqq8O5/KO9NF3leYF8D0RlUooQ5oIQ+JVF2O8STB8Zb8xgLHMv3dkfzGkemNZfo7xHyqkLtr3juuo2vefJyqjljJJAYMiskQEE07pFZnCyeXNTDX163i/6Y6VQ/ZdJGHSNNhTqogFbsJDnSjScvZ36G8+Dtw4y35jcWNVQz3jhTD9UfGjVUM9wA9mwq2ERQHucIj9s0HqiIOzZUhlOEBW2coYsrDFfJYinBJu9wYKWdR10FGEkCLnaDDlEolihIjlqZ3yL/xDtx4S35jcWPlN+5IfuOIBqpWfuNAw1R+9odVkxrNs3bdHBuSXAzqgBAVMFODXHyBGHugVpJkibvYYGPabPLUxuiAGCpIHGLoEUc+qRLdoWK474CNtxTDXdhYw1R3ZJjqyLCxhqkO0XvDIzdTAKqxxYqbD1PlTBrsOOB8IVMOEiEn7RBHSK3XWEdrO9zIhoPhfIeS7IbEZMTJo/SbeiOjZ49e20ly4y3FcBc3Vl78juTFjyi9sfLiB6szInnkhlRrU+ewbg6OlofWJtDCnFKbnIDwrFEVQkshFR/HeEN+I47kawXNscKMvLOmFUFxI8ROaPvyJwmOtxTDXeBY41RrnGqNU70fbHgJjYP2EhPGHjevheulSaWhUPycTpVMNRSXBLJvmWeGnqvsYMOFKEyYwOsogDPHoVUqjFIsjrrIVE4zvfGWWrgLGystfkfS4kekN1Za/EDcSOhcsFhdOgqNsDk3nD0PyrOSrYU+wCCz8eu0YtSCyi7kOna50UMc058IOSQy2KQKharC9M+lWu0DhNOUG2+phru4seTGHZEbR8aNJTcOMJsqJorqnPe+qvPby41AI3LNEEPRiY0CmaNhoI2eUlDKYzctLllT1qygUitg5WHYSMPuzMYJxd6UO+QWfwdsvKUa7sLGkht3RG6sWbinx42KwddUMRFjzYM3rzIy1A+sBSiVOUyVEXJzGYo9IYKDObbd6lQtY594gCLdzam7AYq2AC03bEECaR2nyA1ZdvE9ubFaNm2SFn+3lk3Hl91YLZv2nUzlCscSsoaRZfvO4tkOE4pHGFNoYBeTELUylNirSvRKfjcnXgYNTtogjNlLVpNCFiZILXX03SUaJzlIJcsrvic1Vi3cI5iEe0xqY9XCPVBjcZbiauQgrNG7zdWGlp4bm9AQPxhw6PT8NYLQcvE8Sqh+t0Fs1VhK5FnMKlVALASiEzspiqOeUm90ktxYXvGV3Dg+brxrcuOYuLGSG4cpMdJHmrVYcaDOYrhbY6NjQC6sIJSaKQd1kBMWsPjvvfZqgmO3pmEQ79N0eJQ6B6mqK6AhBXDCyMlnaeU0B6mWVXxh4+5j4/gGqRY29lUbqVCsadjP/hZd2T63wT14xgpJPRk27Babo1bQ0WJiKUHSGzpv+EgxalTQ1mbZ9ZsKI87DyNxjU6Wkd6ix+DtgY1nF98RGevjll589o2UVP3jnDbyzlj+Rs+DICBAWOPZIb4Qk7FOupL62uDk4eqUqLAxB1RiQgwNptYAf7L1D6RrrDjgSUkyVMii6CiYx2ETKIEjUiIuvBYM7SXAsr/ie4Fhe8SPwih+R3lhe8YOZN6qmIOhL7rXl7Wfhji4SijEgDlJATATicwPnQydpESnvFhmR4IZ4HyHkgYAWfEFVAzh7qqLL0jieZPMNWV7xPcGxZuEewSzcIwLHmoV7IG4IRq6jWOxtPvHoW3ODnItdap+9xf2sTxhAQyYYrpdqEKMQyg43LFZOR6IC8zSL1yigUardneg1q4QWTjMvvsziK8FxjNxYCY5/cmyMwUlcHXmUVmMeW2MDk9RYQoUcewYUN5PdGMC7zNmri+UNLWJL984PMX3BMqfhGjaEsgNGzb2E4XLJJ4mN5RVfCY7jw8a7JjiOaDrVSnAcLMHRMd9YtPHmZ/vmRdRHHcn1kIFkEGCvDNo4QRqtUiqdLYrugEOzK7HN0SnfZu8N5yAPP6BKbyONkmI8TXAst/ie4Pj6y6+fPfq3Lz5d4Dg0OIjuIjhuxqn4JsGBfnHjH+dGR3W5+WI/4rGT4va1cBM1DqYw2nCAqROIJHs1DRqtFtdoN79RKrVUuINwcK87dnRKYM8VD9MrNfaTTIzn5Rbfkxur+cYR+P6OaEbVar5xqGK4qCWPEQd1R237GurqPGuoGUIaJh1cClASMqD2PjJGUtntLS5OvYwaAVFtVTSlUpoplTgKp9FDcnSSfvG8/OKLG3efG0eW31jcOISDow1PSD46A0fx2zdtGrVwwQgjFDcdHAU01gAp5WDSgauWXQeHDipYUCF0Z7Bh++6y6wKcMJGY8rAXJ8mN5Rdf41RHyY13Gqc6Nr2xxqkOwI2OkVskLyHYbbq93uiDuYcGOUsA7LOPxhgDJConGUwj7NYZQTfqCDMVEpEAoykPmTXVhQRrjIySTnIebl6G8cWNxY3FjePjhoWv1oeU1KhFC2ebN9+o5FIfBE1DNOngM8hsw+F8So2l+px36xpWlpi7DshlbtSIoJhIgup7yYye22lWUc/LMb7vONWqhvv+q+EeGTdWNdz93RsyfMCWJNufkTefTRWCqkQREJ2t/hgT5EQCI7balbCi7zvUMExEupEobW5kQRa0qEAJ3G1DJ3ysNdR//9P/BxoS7pMF0wEA",
		},
	}

	logs := parseCloudWatchLogs(cloudWatchEvent)

	time := time.Unix(0, 1618555235714*1000000)
	expectedLMEvent := ingest.Log{
		Message:    "{\"eventVersion\":\"1.08\",\"userIdentity\":{\"type\":\"AssumedRole\",\"principalId\":\"AROAS3ZZTSSJZC36CRUZ4:LMAssumeRoleSession\",\"arn\":\"arn:aws:sts::197152445587:assumed-role/LogicMonitor_119/LMAssumeRoleSession\",\"accountId\":\"197152445587\",\"accessKeyId\":\"ASIAS3ZZTSSJVBZZRIN7\",\"sessionContext\":{\"sessionIssuer\":{\"type\":\"Role\",\"principalId\":\"AROAS3ZZTSSJZC36CRUZ4\",\"arn\":\"arn:aws:iam::197152445587:role/LogicMonitor_119\",\"accountId\":\"197152445587\",\"userName\":\"LogicMonitor_119\"},\"webIdFederationData\":{},\"attributes\":{\"mfaAuthenticated\":\"false\",\"creationDate\":\"2021-04-16T06:03:12Z\"}}},\"eventTime\":\"2021-04-16T06:27:05Z\",\"eventSource\":\"ec2.amazonaws.com\",\"eventName\":\"DescribeInstanceStatus\",\"awsRegion\":\"ap-northeast-1\",\"sourceIPAddress\":\"34.221.10.3\",\"userAgent\":\"aws-sdk-java/1.11.918 Linux/4.14.193-149.317.amzn2.x86_64 OpenJDK_64-Bit_Server_VM/11.0.3+7-LTS java/11.0.3 vendor/Amazon.com_Inc.\",\"errorCode\":\"Client.RequestLimitExceeded\",\"errorMessage\":\"Request limit exceeded.\",\"requestParameters\":{\"instancesSet\":{\"items\":[{\"instanceId\":\"i-0d345eec77c8a08b1\"}]},\"filterSet\":{},\"includeAllInstances\":false},\"responseElements\":null,\"requestID\":\"23b2ea29-b7b1-449f-a584-035f485f05cf\",\"eventID\":\"703a3ad3-3d3d-4bd1-8637-3692f850e857\",\"readOnly\":true,\"eventType\":\"AwsApiCall\",\"managementEvent\":true,\"eventCategory\":\"Management\",\"recipientAccountId\":\"197152445587\"}",
		Timestamp:  time,
		ResourceID: map[string]string{"system.aws.accountid": "197152445587", "system.cloud.category": "AWS/LMAccount"},
	}

	assert.Equal(t, expectedLMEvent, logs[0])
}

func TestParseCloudtrailLogsS3(t *testing.T) {
	cloudWatchEvent := events.CloudwatchLogsEvent{
		AWSLogs: events.CloudwatchLogsRawData{
			Data: "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",
		},
	}

	logs := parseCloudWatchLogs(cloudWatchEvent)

	expectedLMEvent := ingest.Log{
		Message:    "{\"eventVersion\":\"1.08\",\"userIdentity\":{\"type\":\"AWSService\",\"invokedBy\":\"cloudtrail.amazonaws.com\"},\"eventTime\":\"2023-03-03T07:30:04Z\",\"eventSource\":\"s3.amazonaws.com\",\"eventName\":\"GetBucketAcl\",\"awsRegion\":\"us-east-1\",\"sourceIPAddress\":\"cloudtrail.amazonaws.com\",\"userAgent\":\"cloudtrail.amazonaws.com\",\"requestParameters\":{\"bucketName\":\"aws-cloudtrail-logs-700010466334-8d075b05\",\"Host\":\"aws-cloudtrail-logs-700010466334-8d075b05.s3.us-east-1.amazonaws.com\",\"acl\":\"\"},\"responseElements\":null,\"additionalEventData\":{\"SignatureVersion\":\"SigV4\",\"CipherSuite\":\"ECDHE-RSA-AES128-GCM-SHA256\",\"bytesTransferredIn\":0,\"AuthenticationMethod\":\"AuthHeader\",\"x-amz-id-2\":\"La8vQCxEf9pcqy/H8Y7Rs7aULfw0Qkc0EI+uKOFTyuMu8of/2a/yvPO6hKck3V5YaGneBCVwzkw=\",\"bytesTransferredOut\":568},\"requestID\":\"TAQNDGTZC32834P4\",\"eventID\":\"2514d9c5-5365-4b24-ac86-241dea825ad6\",\"readOnly\":true,\"resources\":[{\"accountId\":\"700010466334\",\"type\":\"AWS::S3::Bucket\",\"ARN\":\"arn:aws:s3:::aws-cloudtrail-logs-700010466334-8d075b05\"}],\"eventType\":\"AwsApiCall\",\"managementEvent\":true,\"recipientAccountId\":\"700010466334\",\"sharedEventID\":\"59523b9c-953d-4110-b4d5-b8209621af88\",\"eventCategory\":\"Management\"}",
		Timestamp:  time.Date(2023, time.March, 3, 7, 30, 27, 87000000, time.Local),
		ResourceID: map[string]string{"system.aws.arn": "arn:aws:s3:::aws-cloudtrail-logs-700010466334-8d075b05"},
	}

	assert.Equal(t, expectedLMEvent, logs[0])
}

func TestElbGzipLogs(t *testing.T) {
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
	}

	assert.Equal(t, expectedLMEvent, lmEvents[0])
}

//Test case for AWS kinesis logs from cloudtrail
func TestParseKinesisFirehoseLogs(t *testing.T) {
	cloudWatchEvent := events.CloudwatchLogsEvent{
		AWSLogs: events.CloudwatchLogsRawData{
			Data: "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",
		},
	}

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
	}

	assert.Equal(t, expectedLMEvent, logs[4])
}

func TestKinesisFirehoseErrorLog(t *testing.T) {
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
	}

	assert.Equal(t, expectedLMEvent, logs[0])
}

func TestKinesisDataStreamLog(t *testing.T) {
	cloudWatchEvent := events.CloudwatchLogsEvent{
		AWSLogs: events.CloudwatchLogsRawData{
			Data: "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",
		},
	}

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
	}

	assert.Equal(t, expectedLMEvent, logs[1])
}

func TestECSLog(t *testing.T) {
	cloudWatchEvent := events.CloudwatchLogsEvent{
		AWSLogs: events.CloudwatchLogsRawData{
			Data: "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",
		},
	}

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
	}

	assert.Equal(t, expectedLMEvent, logs[44])
}

func TestELBlowLogs(t *testing.T) {
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
	}

	assert.Equal(t, expectedLMEvent, logs[0])
}

func TestRDSFlowLogs(t *testing.T) {
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
	}

	assert.Equal(t, expectedLMEvent, logs[0])
}

func TestFargateLog(t *testing.T) {
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
	}

	assert.Equal(t, expectedLMEvent, logs[0])
}
