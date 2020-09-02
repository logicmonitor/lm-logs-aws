package main

import (
	"github.com/logicmonitor/lm-logs-sdk-go/ingest"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
)

func TestParseELBlogs(t *testing.T) {
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
	lmEvents := parseELBlogs(s3Event, getContentsFromS3BucketMock)

	//Assertion
	expectedLMEvent := ingest.Log{
		Message:    message,
		Timestamp:  time,
		ResourceID: map[string]string{"system.aws.arn": "arn:aws:elasticloadbalancing:us-west-1:123123123123:loadbalancer/test"},
	}

	assert.Equal(t, expectedLMEvent, lmEvents[0])
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

	time, _ := time.Parse(time.RFC3339, "2020-04-08T13:08:34+00:00")
	expectedLMEvent := ingest.Log{
		Message:    "Apr  8 13:08:34 ip-172-40-0-227 dhclient[2221]: XMT: Solicit on eth0, interval 71330ms.",
		Timestamp:  time,
		ResourceID: map[string]string{"system.aws.arn": "arn:aws:ec2::664833354492:instance/i-01fb3c5139e4b27bb"},
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

	time, _ := time.Parse(time.RFC3339, "2020-08-04T23:46:04+00:00")

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

	time, _ := time.Parse(time.RFC3339, "2020-08-05T23:55:21+00:00")

	expectedLMEvent := ingest.Log{
		Message:    "{\"engine\":\"MYSQL\",\"instanceID\":\"database-2\",\"instanceResourceID\":\"db-3AA6AU62HV6KOH2W6IU7IN6I4Q\",\"timestamp\":\"2020-08-05T23:55:21Z\",\"version\":1,\"uptime\":\"00:20:17\",\"numVCPUs\":1,\"cpuUtilization\":{\"guest\":0.0,\"irq\":0.0,\"system\":0.8,\"wait\":0.2,\"idle\":97.3,\"user\":1.6,\"total\":2.7,\"steal\":0.1,\"nice\":0.0},\"loadAverageMinute\":{\"one\":0.0,\"five\":0.0,\"fifteen\":0.0},\"memory\":{\"writeback\":0,\"hugePagesFree\":0,\"hugePagesRsvd\":0,\"hugePagesSurp\":0,\"cached\":435728,\"hugePagesSize\":2048,\"free\":100836,\"hugePagesTotal\":0,\"inactive\":294920,\"pageTables\":3476,\"dirty\":280,\"mapped\":61940,\"active\":524112,\"total\":1019328,\"slab\":42776,\"buffers\":25824},\"tasks\":{\"sleeping\":96,\"zombie\":0,\"running\":0,\"stopped\":0,\"total\":96,\"blocked\":0},\"swap\":{\"cached\":0,\"total\":4095996,\"free\":4095996,\"in\":0.0,\"out\":0.0},\"network\":[{\"interface\":\"eth0\",\"rx\":654.28,\"tx\":2893.35}],\"diskIO\":[{\"writeKbPS\":5.13,\"readIOsPS\":0.17,\"await\":0.71,\"readKbPS\":0.67,\"rrqmPS\":0.0,\"util\":0.04,\"avgQueueLen\":0.0,\"tps\":1.4,\"readKb\":40,\"device\":\"rdsdev\",\"writeKb\":308,\"avgReqSz\":8.29,\"wrqmPS\":0.0,\"writeIOsPS\":1.23},{\"writeKbPS\":27.4,\"readIOsPS\":0.17,\"await\":0.32,\"readKbPS\":0.67,\"rrqmPS\":0.0,\"util\":0.08,\"avgQueueLen\":0.0,\"tps\":2.53,\"readKb\":40,\"device\":\"filesystem\",\"writeKb\":1644,\"avgReqSz\":22.16,\"wrqmPS\":2.27,\"writeIOsPS\":2.37}],\"physicalDeviceIO\":[{\"writeKbPS\":5.13,\"readIOsPS\":1.17,\"await\":0.48,\"readKbPS\":4.67,\"rrqmPS\":0.0,\"util\":0.08,\"avgQueueLen\":0.0,\"tps\":1.67,\"readKb\":280,\"device\":\"xvdg\",\"writeKb\":308,\"avgReqSz\":11.76,\"wrqmPS\":0.68,\"writeIOsPS\":0.5}],\"fileSys\":[{\"used\":379496,\"name\":\"\",\"usedFiles\":210,\"usedFilePercent\":0.02,\"maxFiles\":1310720,\"mountPoint\":\"/rdsdbdata\",\"total\":20496340,\"usedPercent\":1.85},{\"used\":2172928,\"name\":\"\",\"usedFiles\":75334,\"usedFilePercent\":11.5,\"maxFiles\":655360,\"mountPoint\":\"/\",\"total\":10190104,\"usedPercent\":21.32}],\"processList\":[{\"vss\":760392,\"name\":\"OS processes\",\"tgid\":0,\"parentID\":0,\"memoryUsedPc\":3.67,\"cpuUsedPc\":0.02,\"id\":0,\"rss\":37452,\"vmlimit\":0},{\"vss\":2148212,\"name\":\"RDS processes\",\"tgid\":0,\"parentID\":0,\"memoryUsedPc\":26.49,\"cpuUsedPc\":1.47,\"id\":0,\"rss\":270036,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4745,\"rss\":154532,\"vmlimit\":\"unlimited\"},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.02,\"id\":4748,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4749,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4750,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4751,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4752,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.02,\"id\":4753,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4754,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4755,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4756,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4757,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4758,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.15,\"id\":4759,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.02,\"id\":4760,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.02,\"id\":4761,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4762,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.02,\"id\":4763,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.02,\"id\":4764,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4765,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4766,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4767,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4780,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4782,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4784,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4785,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4786,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4788,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4789,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4790,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4791,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.02,\"id\":4795,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4796,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4797,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4798,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4799,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4814,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.0,\"id\":4867,\"rss\":154532,\"vmlimit\":0},{\"vss\":720284,\"name\":\"mysqld\",\"tgid\":4745,\"parentID\":4741,\"memoryUsedPc\":15.16,\"cpuUsedPc\":0.05,\"id\":5100,\"rss\":154532,\"vmlimit\":0}]}",
		Timestamp:  time,
		ResourceID:  map[string]string{"system.aws.arn": "arn:aws:rds::664833354492:db:database-2"},
	}

	assert.Equal(t, expectedLMEvent, logs[0])
}

func TestLambdaLogs(t *testing.T) {
	cloudWatchEvent := events.CloudwatchLogsEvent{
		AWSLogs: events.CloudwatchLogsRawData{
			Data: "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",
		},
	}

	logs := parseCloudWatchLogs(cloudWatchEvent)

	time, _ := time.Parse(time.RFC3339, "2020-08-27T08:41:49.043+00:00")

	expectedLMEvent := ingest.Log{
		Message:    "START RequestId: b8cf7efa-a997-4a31-a1ff-881feee2273e Version: $LATEST\n",
		Timestamp:  time,
		ResourceID:  map[string]string{"system.aws.arn": "arn:aws:lambda::197152445587:function:observatory-worker"},
	}

	assert.Equal(t, expectedLMEvent, logs[0])
}

