package storage

import (
	"context"
	"log"
	"testing"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func TestS3ListBuckets(t *testing.T){

    strg := NewClient()
	result, err := strg.Client.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
	if err != nil {
		t.Fatal(err)
	}

	for _, bucket := range result.Buckets {
		log.Printf("bucket=%s creation time=%s", aws.ToString(bucket.Name), bucket.CreationDate.Format("2006-01-02 15:04:05 Monday"))
	}
}

func TestS3FileUpload(t *testing.T) {
	// Create a temporary file to upload
	data := []byte("Hello, S3!")
    strg := NewClient()
	key := "example.txt"
    

    err := strg.UploadFile(data, key)

    if err != nil {
        t.Fatalf("Error uploading file: %s \n", err)
    }
}
