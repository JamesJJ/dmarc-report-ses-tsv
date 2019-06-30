package main

type S3EventMsg struct {
	Records       []S3EventRecord `json:"Records"`
	ReceiptHandle string          `json:"ReceiptHandle"`
}

type S3Data struct {
	Bucket S3BucketData `json:"bucket"`
	Object S3ObjectData `json:"object"`
}

type S3BucketData struct {
	Name string `json:"name"`
}

type S3ObjectData struct {
	Key  string `json:"key"`
	Size int64  `json:"size"`
}

type S3EventRecord struct {
	EventSource string `json:"eventSource"`
	AwsRegion   string `json:"awsRegion"`
	EventName   string `json:"eventName"`
	S3          S3Data `json:"s3"`
}
