package main

import (
	"bytes"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	//	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func S3Upload(bucket *string, item *string, region *string, bodyBuf *bytes.Buffer) error {

	Info.Printf("Uploading: s3://%s/%s (%s, %d)", *bucket, *item, *region, bodyBuf.Len())

	// The session the S3 Uploader will use
	session, err := session.NewSession(&aws.Config{
		Region: aws.String(*region),
	})
	if err != nil {
		return fmt.Errorf("S3 Upload Session Error: %v", err)
	}

	// Create an uploader with the session and custom options
	uploader := s3manager.NewUploader(session, func(u *s3manager.Uploader) {
		u.PartSize = 5 * 1024 * 1024 // Must be at least 5MB
	})

	// Upload input parameters
	upParams := &s3manager.UploadInput{
		Bucket: bucket,
		Key:    item,
		Body:   bodyBuf,
	}

	// Perform an upload.
	result, err := uploader.Upload(upParams)
	if err != nil {
		return fmt.Errorf("Unable to upload item to s3://%s/%s: %v", *bucket, *item, err)
	}
	Debug.Printf("Upload Result=%v", result)

	return nil
}
