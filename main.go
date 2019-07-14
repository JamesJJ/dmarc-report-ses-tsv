package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/jamiealquiza/envy"
	"math/rand"
	"os"
	"sync"
	"time"
)

var (
	wg sync.WaitGroup
)

type config struct {
	sqsName                  *string
	sqsRegion                *string
	s3Name                   *string
	s3Region                 *string
	sqsPollTimeout           *int64
	sqsPollMaxMessages       *int64
	sqsVisibilityTimeout     *int64
	doneAfterCountEmptyPolls *int
	maxRecordsPerFile        *int
	moveFilesAfterProcessing *string
	logVerbose               *bool
	sqsDelete                *bool
	runDate                  *string
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {

	runDate := time.Now().UTC().Format("20060102")

	conf := config{
		flag.String("sqs", "", "Name of the SQS queue to poll [MANDATORY]"),
		flag.String("sqsregion", "", "AWS region of SQS queue [MANDATORY]"),
		flag.String("bucket", "", "Name of the S3 bucket to store TSV files [MANDATORY]"),
		flag.String("bucketregion", "", "AWS region of S3 bucket [MANDATORY]"),
		flag.Int64("polltimeout", 10, "SQS slow poll timeout, 1-20"),
		flag.Int64("pollmessages", 10, "SQS maximum messages per poll, 1-10"),
		flag.Int64("sqsprocessingtime", 3600, "SQS visibility timeout [DO NOT CHANGE]"),
		flag.Int("emptypolls", 3, "How many consecutive times to poll SQS and receive zero messages before exiting, 1+"),
		flag.Int("maxrecords", 32, "Maximum number * 1024 of records in a single S3 file, 1+, e.g 2 sets the limit to 2048"),
		flag.String("move", "", "Move email to this S3 prefix after processing. Date will be automatically added"),
		flag.Bool("verbose", false, "Show detailed information during run"),
		flag.Bool("deletesqs", true, "Delete messages from SQS after processing"),
		&runDate,
	}
	envy.Parse("DMARC")
	flag.Parse()

	if *conf.sqsName == "" ||
		*conf.sqsRegion == "" ||
		*conf.s3Name == "" ||
		*conf.s3Region == "" ||
		*conf.sqsPollTimeout < 1 ||
		*conf.sqsPollTimeout > 20 ||
		*conf.sqsPollMaxMessages > 20 ||
		*conf.sqsPollMaxMessages < 1 ||
		*conf.maxRecordsPerFile < 1 ||
		*conf.doneAfterCountEmptyPolls < 1 {
		flag.Usage()
		os.Exit(1)
	}

	logInit(conf)

	writeTSVChan := make(chan *[]string)
	uploadToS3Chan := make(chan *bytes.Buffer)
	deleteSqsChan := make(chan *string)
	moveS3FileChan := make(chan *S3EventRecord)

	if *conf.moveFilesAfterProcessing != "" {
		wg.Add(1)
		go func(conf config, wg *sync.WaitGroup, moveS3FileChan chan *S3EventRecord) {
			defer wg.Done()
			for {
				err := S3Move(conf, moveS3FileChan)
				if err == nil {
					break
				}
				time.Sleep(3 * time.Second)
			}
		}(conf, &wg, moveS3FileChan)

	}

	wg.Add(1)
	go func(conf config, wg *sync.WaitGroup, deleteSqsChan chan *string) {
		defer wg.Done()
		for {
			err := sqsDelete(conf, deleteSqsChan)
			if err == nil {
				break
			}
			time.Sleep(3 * time.Second)
		}
	}(conf, &wg, deleteSqsChan)

	wg.Add(1)
	go func(conf config, wg *sync.WaitGroup) {
		defer wg.Done()

		for file := range uploadToS3Chan { // TODO: Fix filename
			s3OutputPath := fmt.Sprintf("dmarc-data/go-upload-test-%s-%s.tsv.gz", time.Now().UTC().Format("20060102-150405"), RandStringBytes(6))
			S3Upload( // TODO: ERROR CHECKING
				conf.s3Name,
				&s3OutputPath,
				conf.s3Region,
				file,
			)
			returnToPool(file)
		}
	}(conf, &wg)

	wg.Add(1)
	go func(conf config, wg *sync.WaitGroup) {
		defer func() {
			close(uploadToS3Chan)
			wg.Done()
		}()
		WriteTSV(conf, writeTSVChan, uploadToS3Chan)
	}(conf, &wg)

	gracefulStop(func() {})

	pollCount := *conf.doneAfterCountEmptyPolls
	for pollCount > 0 {

		Debug.Printf("pollCount=%d", pollCount)

		s3records, err := PollSQS(conf)
		if err != nil {
			Error.Printf("Failed to poll SQS: %v", err)
			pollCount--
			continue
		}

		pollCount--
		for _, s3msgs := range s3records {
			pollCount = *conf.doneAfterCountEmptyPolls
			for _, msgRecord := range s3msgs.Records {
				bReader, s3Retry, errS3Download := S3Download(
					&msgRecord.S3.Bucket.Name,
					&msgRecord.S3.Object.Key,
					&msgRecord.AwsRegion,
				)
				if errS3Download == nil {
					ReadMail(
						bReader,
						writeTSVChan,
					)
					if *conf.moveFilesAfterProcessing != "" {
						moveS3FileChan <- &msgRecord
					}
				} else {
					Error.Printf(
						"Failed to download from S3: s3://%s/%s (retry_later=%v)",
						msgRecord.S3.Bucket.Name,
						msgRecord.S3.Object.Key,
						s3Retry,
					)
				}
				if *conf.sqsDelete && !s3Retry {
					deleteSqsChan <- &s3msgs.ReceiptHandle
				}

			}
		}
	}

	close(deleteSqsChan)
	close(moveS3FileChan)
	close(writeTSVChan)
	wg.Wait()

}
