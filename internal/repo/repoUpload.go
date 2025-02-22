package repo

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/adityaw24/go-aws-garasi/internal/model"
	"github.com/adityaw24/go-aws-garasi/utils"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"github.com/gin-gonic/gin"
)

type RepoUpload interface {
	UploadFile(ctx *gin.Context, file io.Reader, objectKey string, attach utils.Upload, largeObject []byte) error
	ListObjects(ctx *gin.Context) ([]model.FileModel, error)
	PreviewFile(ctx *gin.Context, objectKey string) (string, error)
	CopyObject(ctx *gin.Context, objectRequest *model.CopyObjectRequest) error
	UpdateFile(ctx *gin.Context, oldKey string, file io.Reader, newKey string, attach utils.Upload, largeObject []byte) error
	DeleteFile(ctx *gin.Context, key string) error
}

type repoUpload struct {
	s3Client          *s3.Client
	s3PresignedClient *s3.PresignClient
	bucketName        string
	timeout           time.Duration
}

func NewRepoUpload(client *s3.Client, presignClient *s3.PresignClient, bucketName string, timeout time.Duration) *repoUpload {
	return &repoUpload{
		s3Client:          client,
		s3PresignedClient: presignClient,
		bucketName:        bucketName,
		timeout:           timeout,
	}
}

func (repo *repoUpload) UploadFile(ctx *gin.Context, file io.Reader, objectKey string, attach utils.Upload, largeObject []byte) error {
	largeBuffer := bytes.NewReader(largeObject)
	var partMiBs int64 = 10

	key := attach.Prefix + objectKey

	uploader := manager.NewUploader(repo.s3Client, func(u *manager.Uploader) {
		u.PartSize = partMiBs * 1024 * 1024
	})

	_, err := uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(repo.bucketName),
		Key:           aws.String(key),
		Body:          largeBuffer,
		ContentLength: aws.Int64(attach.Length),
		ContentType:   aws.String(attach.ContentType),
	})

	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) && apiErr.ErrorCode() == "EntityTooLarge" {
			errMsg := "error uploading. The object is too large.\n" +
				"The maximum size for a multipart upload is 5TB."
			utils.ErrorResp(ctx, http.StatusBadRequest, errMsg)
			return nil
		} else {
			log.Printf("Couldn't upload large object to %v:%v. Here's why: %v\n",
				repo.bucketName, objectKey, err)
			return err
		}
	}

	return nil
}

func (repo *repoUpload) UpdateFile(ctx *gin.Context, oldKey string, file io.Reader, newKey string, attach utils.Upload, largeObject []byte) error {
	_, err := repo.s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(repo.bucketName),
		Key:    aws.String(oldKey),
	})
	if err != nil {
		var notFound *types.NotFound
		if errors.As(err, &notFound) {
			// utils.ErrorResp(ctx, http.StatusNotFound, "file not found")
			return fmt.Errorf("file %s not found", oldKey)
			// return nil
		}
		return fmt.Errorf("error checking file existence: %v", err)
	}

	err = repo.DeleteFile(ctx, oldKey)
	if err != nil {
		return fmt.Errorf("error deleting old file: %v", err)
	}

	largeBuffer := bytes.NewReader(largeObject)
	var partMiBs int64 = 10

	key := attach.Prefix + newKey

	uploader := manager.NewUploader(repo.s3Client, func(u *manager.Uploader) {
		u.PartSize = partMiBs * 1024 * 1024
	})

	_, err = uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(repo.bucketName),
		Key:           aws.String(key),
		Body:          largeBuffer,
		ContentLength: aws.Int64(attach.Length),
		ContentType:   aws.String(attach.ContentType),
	})

	if err != nil {
		var smithyErr *smithy.GenericAPIError
		if errors.As(err, &smithyErr) {
			log.Printf("Upload error occurred: %v\n", smithyErr.Error())
			return smithyErr
		}
		log.Printf("Other error occurred: %v\n", err.Error())
		return err
	}

	return nil
}

func (repo *repoUpload) DeleteFile(ctx *gin.Context, key string) error {
	_, err := repo.s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(repo.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		var notFound *types.NotFound
		if errors.As(err, &notFound) {
			return fmt.Errorf("file %s not found", key)
		}
		return fmt.Errorf("error checking file existence: %v", err)
	}

	_, err = repo.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(repo.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("error deleting file: %v", err)
	}

	err = s3.NewObjectNotExistsWaiter(repo.s3Client).Wait(ctx,
		&s3.HeadObjectInput{
			Bucket: aws.String(repo.bucketName),
			Key:    aws.String(key),
		},
		repo.timeout)
	if err != nil {
		return fmt.Errorf("error waiting for file deletion: %v", err)
	}

	return nil
}

func (repo *repoUpload) ListObjects(ctx *gin.Context) ([]model.FileModel, error) {
	var err error
	var output *s3.ListObjectsV2Output
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(repo.bucketName),
	}

	var objects []model.FileModel
	objectPaginator := s3.NewListObjectsV2Paginator(repo.s3Client, input)
	for objectPaginator.HasMorePages() {
		output, err = objectPaginator.NextPage(ctx)
		if err != nil {
			var noBucket *types.NoSuchBucket
			if errors.As(err, &noBucket) {
				log.Printf("Bucket %s does not exist.\n", repo.bucketName)
				err = noBucket
			}
			break
		} else {
			for _, item := range output.Contents {
				url, _ := repo.PreviewFile(ctx, *item.Key)

				title := *item.Key
				if idx := strings.LastIndex(title, "_"); idx != -1 {
					title = title[:idx]
				}

				fileModel := model.FileModel{
					Key:   *item.Key,
					Title: title,
					Url:   url,
				}
				objects = append(objects, fileModel)
			}
		}
	}
	return objects, err
}

func (repo *repoUpload) PreviewFile(ctx *gin.Context, objectKey string) (string, error) {
	presignResult, err := repo.s3PresignedClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(repo.bucketName),
		Key:    aws.String(objectKey),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = repo.timeout
	})
	if err != nil {
		log.Printf("Couldn't get presigned URL for object %v:%v. Here's why: %v\n",
			repo.bucketName, objectKey, err)
		return "", err
	}

	return presignResult.URL, nil
}

func (repo *repoUpload) CopyObject(ctx *gin.Context, objectRequest *model.CopyObjectRequest) error {
	input := &s3.CopyObjectInput{
		Bucket:     aws.String(repo.bucketName),
		CopySource: aws.String(repo.bucketName + "/" + objectRequest.OldKey),
		Key:        aws.String(objectRequest.NewKey),
	}

	_, err := repo.s3Client.CopyObject(ctx, input)
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			log.Printf("error copying object, code: %s, message: %s", apiErr.ErrorCode(), apiErr.ErrorMessage())
			return fmt.Errorf("failed to copy object: %s", apiErr.ErrorMessage())
		}
		return fmt.Errorf("unexpected error copying object: %v", err)
	}

	return nil
}
