package main

import (
	"bytes"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func S3Download(bucket *string, item *string, region *string) (*bytes.Reader, bool, error) {

	Debug.Printf("Downloading s3://%s/%s (%s)\n\n", *bucket, *item, *region)

	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String(*region)},
	)

	downloader := s3manager.NewDownloader(sess)

	buff := &aws.WriteAtBuffer{}

	numBytes, err := downloader.Download(buff,
		&s3.GetObjectInput{
			Bucket: aws.String(*bucket),
			Key:    aws.String(*item),
		})
	if err != nil {

		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchBucket:
				return nil, false, fmt.Errorf("S3 Download Error: Bucket not exist: %s (%v)", *bucket, err)
			case s3.ErrCodeNoSuchKey:
				return nil, false, fmt.Errorf("S3 Download Error: File not exist: %s (%v)", *item, err)
			}
		}
		return nil, true, fmt.Errorf("S3 Download Error: s3://%s/%s (%v)", *bucket, *item, err)
	}

	Debug.Printf("Downloaded %d bytes", numBytes)

	return bytes.NewReader(buff.Bytes()), false, nil
}
