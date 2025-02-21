package pkg

import (
	"fmt"
	"os"

	"github.com/adityaw24/go-aws-garasi/configs"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/session"
)

func CreateSession() (*session.Session, error) {
	accessKey, secretKey, region := configs.GetAWSConfig()

	return session.NewSession(&aws.Config{
		Region: region,
		Credentials: credentials.NewStaticCredentialsProvider(
			accessKey,
			secretKey,
			"",
		),
	})
}

func UploadFileToS3(filePath, key, bucket, prefix string) error {
	sess, err := CreateSession()
	if err != nil {
		return err
	}

	svc := s3.New(sess)

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	objectKey := fmt.Sprintf("%s/%s", prefix, key)

	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket:        aws.String(bucket),
		Key:           aws.String(objectKey),
		Body:          file,
		ContentLength: aws.Int64(fileInfo.Size()),
	})
	if err != nil {
		return err
	}

	return nil
}
