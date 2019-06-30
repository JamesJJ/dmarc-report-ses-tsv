package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"strings"
)

func sqsDelete(conf config, deleteSqsChan chan *string) error {

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(*conf.sqsRegion)},
	)
	if err != nil {
		if err != nil {
			return fmt.Errorf("SQS Delete Session Error: %v", err)
		}
	}

	sqsClient := sqs.New(sess)

	sqsUrl, err := sqsClient.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: aws.String(*conf.sqsName),
	})
	if err != nil {
		if err != nil {
			return fmt.Errorf("SQS Delete URL Error: %v", err)
		}
	}

	for receiptHandle := range deleteSqsChan {
		_, err := sqsClient.DeleteMessage(&sqs.DeleteMessageInput{
			QueueUrl:      sqsUrl.QueueUrl,
			ReceiptHandle: receiptHandle,
		})
		if err != nil {
			Error.Printf("SQS Delete Error: %v", err)
		}
	}
	return nil
}

func PollSQS(conf config) ([]*S3EventMsg, error) {

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(*conf.sqsRegion)},
	)

	Debug.Printf("Polling SQS: %s, in %s", *conf.sqsName, *conf.sqsRegion)

	sqsClient := sqs.New(sess)

	sqsUrl, err := sqsClient.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: aws.String(*conf.sqsName),
	})
	if err != nil {
		if awserr, ok := err.(awserr.Error); ok && awserr.Code() == sqs.ErrCodeQueueDoesNotExist {
			return nil, fmt.Errorf("Unable to find queue %q.", *conf.sqsName)
		}
		return nil, fmt.Errorf("Unable to poll queue %q, %v.", *conf.sqsName, err)
	}

	result, err := sqsClient.ReceiveMessage(&sqs.ReceiveMessageInput{
		VisibilityTimeout: conf.sqsVisibilityTimeout,
		QueueUrl:          sqsUrl.QueueUrl,
		AttributeNames: aws.StringSlice([]string{
			"SentTimestamp",
		}),
		MaxNumberOfMessages: aws.Int64(*conf.sqsPollMaxMessages),
		MessageAttributeNames: aws.StringSlice([]string{
			"All",
		}),
		WaitTimeSeconds: aws.Int64(*conf.sqsPollTimeout),
	})
	if err != nil {
		return nil, fmt.Errorf("Unable to receive message from queue %q, %v.", *conf.sqsName, err)
	}

	Debug.Printf("SQS received %d messages.\n", len(result.Messages))
	return sqsDecodeMap(result.Messages, sqsDecode), nil
}

func sqsDecodeMap(rmsgs []*sqs.Message, f func(*sqs.Message) *S3EventMsg) []*S3EventMsg {
	m := make([]*S3EventMsg, len(rmsgs))
	for i, v := range rmsgs {
		m[i] = f(v)
	}
	return m
}

func sqsDecode(r *sqs.Message) *S3EventMsg {

	var s3msg S3EventMsg
	s3msg.ReceiptHandle = *r.ReceiptHandle

	recordSR := strings.NewReader(*r.Body)
	if err := json.NewDecoder(recordSR).Decode(&s3msg); err != nil {
		Error.Printf("SQS-S3 JSON Error: %#v", err)
	}
	Debug.Printf("UnMarshalled SQS JSON=%+v\n\n", s3msg)
	return &s3msg
}
