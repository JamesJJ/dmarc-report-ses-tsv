package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"path/filepath"
)

func S3RenameFile(svc *s3.S3, bucket *string, source *string, destination *string) error {

	inputCopy := &s3.CopyObjectInput{
		Bucket:     aws.String(*bucket),
		CopySource: aws.String(fmt.Sprintf("%s/%s", *bucket, *source)),
		Key:        aws.String(*destination),
	}

	_, errCopy := svc.CopyObject(inputCopy)
	if errCopy != nil {
		return fmt.Errorf("S3 Rename Copy Error: %v", errCopy)
	}

	inputDelete := &s3.DeleteObjectInput{
		Bucket: aws.String(*bucket),
		Key:    aws.String(*source),
	}

	_, errDelete := svc.DeleteObject(inputDelete)
	if errDelete != nil {
		return fmt.Errorf("S3 Rename Delete Error: %v", errDelete)
	}
	return nil
}

func S3Move(conf config, moveS3FileChan chan *S3EventRecord) error {

	// This session is using conf for region, so is in report output bucket region
	// Files being moved are in email input bucket region
	// If the buckets are in different regions, then .... !???

	session, err := session.NewSession(&aws.Config{
		Region: aws.String(*conf.s3Region),
	})
	if err != nil {
		return fmt.Errorf("S3 Move Session Error: %v", err)
	}

	svc := s3.New(session)

	for msgRecord := range moveS3FileChan {

		newKey := filepath.Join(*conf.moveFilesAfterProcessing, *conf.runDate, filepath.Base(msgRecord.S3.Object.Key))

		Debug.Printf(
			"Moving: s3://%s/%s to s3://%s/%s (%s)",
			msgRecord.S3.Bucket.Name,
			msgRecord.S3.Object.Key,
			msgRecord.S3.Bucket.Name,
			newKey,
			msgRecord.AwsRegion,
		)
		err := S3RenameFile(
			svc,
			&msgRecord.S3.Bucket.Name,
			&msgRecord.S3.Object.Key,
			&newKey,
		)
		if err != nil {
			Error.Printf("S3 Move Error: %v", err)
		}

	}
	return nil
}
