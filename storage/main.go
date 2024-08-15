package storage

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/rs/zerolog"
)

var logger = zerolog.New(os.Stdout)

type Storage struct {
    Client *s3.Client
}

func NewClient() Storage {
    cfg := aws.NewConfig()
    cfg.Region = os.Getenv("AWS_REGION")
    cfg.EndpointResolverWithOptions = aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		if service == s3.ServiceID && region == "ru-central1" {
			return aws.Endpoint{
				PartitionID:   "yc",
				URL:           "https://storage.yandexcloud.net",
				SigningRegion: "ru-central1",
			}, nil
		}
		return aws.Endpoint{}, fmt.Errorf("unknown endpoint requested")
	})
    cfg.Credentials = aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
        return aws.Credentials{
            AccessKeyID: os.Getenv("AWS_ACCESS_KEY_ID"),
            SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
        }, nil
    })	

    // Создаем клиента для доступа к хранилищу S3
    return Storage{Client: s3.NewFromConfig(*cfg)}
}

func (storage Storage) UploadFile(data []byte, fileName string) error {
    file, err := os.CreateTemp("", fileName)
	if err != nil {
		logger.Err(err)
        return err
	}
	defer os.Remove(file.Name())
    
	// Write some data to the file
	if _, err := file.Write(data); err != nil {
		logger.Err(err)
        return err
	}
    path := file.Name()
    file.Close() 
    file, err = os.Open(path)

    if err != nil {
        logger.Err(err)
        return err
    }
	// Create a new S3 bucket and object key
	bucket := "userfiles" // Replace with your S3 bucket name

	// Upload the file to S3
	_, err = storage.Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(fileName),
		Body:   file,
	})

	if err != nil {
		logger.Err(err)
        return err
	}

	// Validate that the file was successfully uploaded
	resp, err := storage.Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(fileName),
	})
	if err != nil {
		logger.Err(err)
        return err
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		logger.Err(err)
        return err
	}

	if !bytes.Equal(buf.Bytes(), data) {
        err = fmt.Errorf("Written data does not match original data")       
        logger.Err(err)
        return err
	}
    return nil
}

func (storage Storage) GetFile(key string) (string, error) {
    resp, err := storage.Client.GetObject(context.TODO(), &s3.GetObjectInput{
       Key: aws.String(key), 
       Bucket: aws.String("userfiles"),
    })
    if err != nil {
        logger.Err(err)
        return "", err
    }
    defer resp.Body.Close()

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		logger.Err(err)
        return "", err
	}

    file, err := os.CreateTemp("", key)
    if err != nil {
        logger.Err(err)
        return "", err
    }
    defer file.Close()
    _, err = file.Write(buf.Bytes())
    if err != nil {
        logger.Err(err)
        return "", err
    }
    return file.Name(), nil
}

func (storage Storage) DeleteFile(key string) error {
    _, err := storage.Client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
        Bucket: aws.String("userfiles"),
        Key: aws.String(key),
    })
    if err != nil {
        logger.Err(err)
        return err
    }
    return nil
}

