package main

import (
	"errors"
	"fmt"

	"github.com/mitchellh/mapstructure"
)

type s3event struct {
	Records []s3Record
}

type s3Record struct {
	EventTime string
	S3        s3Struct
	AWSRegion string
}

type s3Struct struct {
	Object    s3Object
	Bucket    s3Bucket
	AWSRegion string
	EventName string
}

type s3Bucket struct {
	ARN  string
	Name string
}

type s3Object struct {
	Key       string
	Size      int64
	ETag      string
	Sequencer string
}

/*
var s3Event = `{
  "Records": [
    {
      "eventVersion": "2.0",
      "eventTime": "1970-01-01T00:00:00.000Z",
      "requestParameters": {
        "sourceIPAddress": "127.0.0.1"
      },
      "s3": {
        "configurationId": "testConfigRule",
        "object": {
          "eTag": "0123456789abcdef0123456789abcdef",
          "sequencer": "0A1B2C3D4E5F678901",
          "key": "{{.Key}}",
          "size": 1024
        },
        "bucket": {
          "arn": "arn:aws:s3:::{{.Bucket}}",
          "name": "{{.Bucket}}",
          "ownerIdentity": {
            "principalId": "EXAMPLE"
          }
        },
        "s3SchemaVersion": "1.0"
      },
      "responseElements": {
        "x-amz-id-2": "EXAMPLE123/5678abcdefghijklambdaisawesome/mnopqrstuvwxyzABCDEFGH",
        "x-amz-request-id": "EXAMPLE123456789"
      },
      "awsRegion": "{{.Region}}",
      "eventName": "ObjectCreated:Put",
      "userIdentity": {
        "principalId": "EXAMPLE"
      },
      "eventSource": "aws:s3"
    }
  ]
}`
*/

func getBucketAndKey(input map[string]interface{}) (string, string, string, error) {
	var event s3event
	if err := mapstructure.Decode(input, &event); err != nil {
		return "", "", "", err
	}

	fmt.Printf("Decoded event (partial): %#v\n", event)

	if len(event.Records) == 0 {
		return "", "", "", errors.New("no s3 records in event")
	}

	s3 := event.Records[0].S3
	return s3.Bucket.Name, s3.Object.Key, event.Records[0].AWSRegion, nil
}
