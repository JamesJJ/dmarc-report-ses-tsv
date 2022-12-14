# Deploy to AWS Serverless

This directory includes a [template](./template.yml) and [config](./samconfig.toml) to deploy this app using the [AWS Serverless Application Model](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/what-is-sam.html).

## Prerequisites

You'll need:

 - `go` installed
 - [AWS SAM CLI](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/serverless-sam-cli-install.html) installed
 - AWS CLI credentials set-up

## Usage

 1. Execute `./deploy.sh -g`
 2. When asked `Parameter BUCKET:`, provide the name of the S3 bucket that you have set-up to collect DMARC emails received by AWS SES. *Ensure SES puts the emails in a sub-directory and not the root of the bucket: specify a prefix when you configure the S3 receipt rule in SQS.*
 3. The defaults are usually suitable for the other prompts.
 4. When the stack has finished deploying, it will show the `SqsQueueArn` in the outputs. You should set the S3 bucket to send `s3:ObjectCreated:Put` events to this SQS ARN. S3 events to SQS determines which raw DMARC email files are processed.

## What will it do?

 - **Once a day** the SQS queue will be read, and corresponding raw DMARC emails in S3 will be processed (look for `cron` in the [template](./template.yml) to [understand the schedule](https://docs.aws.amazon.com/AmazonCloudWatch/latest/events/ScheduledEvents.html#CronExpressions))
 - Raw DMARC emails that have been processed will be moved to `archive-dmarc/...` in the S3 bucket
 - Generated TSV data files will be placed in `data-dmarc/...` in the S3 bucket.

**Note:** Just re-highlight, processing is intentionally not triggered by the arrival of raw DMARC emails. Instead, all new raw DMARC email files are batch processed once a day, in-order to create a smaller number of larger output files.
