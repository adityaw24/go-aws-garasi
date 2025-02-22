package usecase

import (
	"errors"
	"io"
	"net/http"
	"path/filepath"

	"github.com/adityaw24/go-aws-garasi/internal/model"
	"github.com/adityaw24/go-aws-garasi/internal/repo"
	"github.com/adityaw24/go-aws-garasi/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UsecaseUpload interface {
	UploadFile(ctx *gin.Context, fileRequest *model.FileRequest, validMimeTypes []string) (listObjects []model.FileModel, err error)
	PreviewFile(ctx *gin.Context, objectKey string) (string, error)
	ListObjects(ctx *gin.Context) ([]model.FileModel, error)
	UpdateFile(ctx *gin.Context, fileRequest *model.UpdateFileRequest, validMimeTypes []string) error
	DeleteFile(ctx *gin.Context, fileRequest *model.DeleteFileRequest) error
	UpdateObject(ctx *gin.Context, objectRequest *model.CopyObjectRequest) error
}

type usecaseUpload struct {
	repo repo.RepoUpload
}

func NewUsecaseUpload(repo repo.RepoUpload) UsecaseUpload {
	return &usecaseUpload{
		repo: repo,
	}
}

func (u *usecaseUpload) UploadFile(ctx *gin.Context, fileRequest *model.FileRequest, validMimeTypes []string) (listObjects []model.FileModel, err error) {

	file, err := fileRequest.File.Open()
	if err != nil {
		utils.ErrorLog("usecase", "UploadFile Open Fileheader", err)
		return nil, err
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		utils.ErrorLog("usecase", "UploadFile ReadAll", err)
		return nil, err
	}

	_, err = file.Seek(0, 0)
	if err != nil {
		utils.ErrorLog("usecase", "UploadFile Seek Fileheader", err)
		return nil, err
	}

	buff := make([]byte, 512)
	_, err = file.Read(buff)
	if err != nil {
		utils.ErrorLog("usecase", "UploadFile Read Fileheader", err)
		return nil, err
	}

	_, err = file.Seek(0, 0)
	if err != nil {
		utils.ErrorLog("usecase", "UploadFile Seek Fileheader", err)
		return nil, err
	}

	contentType := http.DetectContentType(buff)

	err = utils.ValidateContentType(contentType, validMimeTypes)
	if err != nil {
		utils.ErrorLog("usecase", "UploadFile ValidateContentType", err)
		return nil, err
	}

	fileUpload := utils.Upload{
		Length:      fileRequest.File.Size,
		ContentType: contentType,
		Prefix:      fileRequest.Title + "_",
		Ext:         filepath.Ext(fileRequest.File.Filename),
	}

	key := uuid.New().String() + fileUpload.Ext

	err = u.repo.UploadFile(ctx, file, key, fileUpload, fileBytes)
	if err != nil {
		utils.ErrorLog("usecase", "UploadFile Repository", err)
		return nil, err
	}

	objects, err := u.repo.ListObjects(ctx)
	if err != nil {
		utils.ErrorLog("usecase", "UploadFile Repository ListObjects", err)
		return nil, err
	}

	// var (
	// 	results []model.FileModel
	// )
	// for _, object := range objects {
	// 	url, _ := u.repo.PreviewFile(ctx, *object.Key)
	// 	result := model.FileModel{
	// 		Key:   *object.Key,
	// 		Title: fileRequest.Title,
	// 		Url:   url,
	// 	}

	// 	results = append(results, result)
	// }

	return objects, nil
}

func (u *usecaseUpload) PreviewFile(ctx *gin.Context, objectKey string) (string, error) {
	presignedURL, err := u.repo.PreviewFile(ctx, objectKey)
	if err != nil {
		utils.ErrorLog("usecase", "PreviewFile Repository", err)
		return "", err
	}

	return presignedURL, nil
}

func (u *usecaseUpload) ListObjects(ctx *gin.Context) ([]model.FileModel, error) {
	objects, err := u.repo.ListObjects(ctx)
	if err != nil {
		utils.ErrorLog("usecase", "ListObjects Repository", err)
		return nil, err
	}

	return objects, nil
}

func (u *usecaseUpload) UpdateFile(ctx *gin.Context, fileRequest *model.UpdateFileRequest, validMimeTypes []string) error {
	file, err := fileRequest.File.Open()
	if err != nil {
		utils.ErrorLog("usecase", "UpdateFile Open Fileheader", err)
		return err
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		utils.ErrorLog("usecase", "UpdateFile ReadAll", err)
		return err
	}

	_, err = file.Seek(0, 0)
	if err != nil {
		utils.ErrorLog("usecase", "UpdateFile Seek Fileheader", err)
		return err
	}

	buff := make([]byte, 512)
	_, err = file.Read(buff)
	if err != nil {
		utils.ErrorLog("usecase", "UpdateFile Read Fileheader", err)
		return err
	}

	contentType := http.DetectContentType(buff)
	err = utils.ValidateContentType(contentType, validMimeTypes)
	if err != nil {
		utils.ErrorLog("usecase", "UpdateFile ValidateContentType", err)
		return err
	}

	fileUpload := utils.Upload{
		Length:      fileRequest.File.Size,
		ContentType: contentType,
		Prefix:      fileRequest.Title + "_",
		Ext:         filepath.Ext(fileRequest.File.Filename),
	}

	newKey := uuid.New().String() + fileUpload.Ext

	err = u.repo.UpdateFile(ctx, fileRequest.Key, file, newKey, fileUpload, fileBytes)
	if err != nil {
		utils.ErrorLog("usecase", "UpdateFile Repository", err)
		return err
	}

	return nil
}

func (u *usecaseUpload) DeleteFile(ctx *gin.Context, fileRequest *model.DeleteFileRequest) error {
	if fileRequest.Key == "" {
		utils.ErrorLog("usecase", "DeleteFile", errors.New("key is required"))
		return errors.New("key is required")
	}

	err := u.repo.DeleteFile(ctx, fileRequest.Key)
	if err != nil {
		utils.ErrorLog("usecase", "DeleteFile Repository", err)
		return err
	}

	return nil
}

func (u *usecaseUpload) UpdateObject(ctx *gin.Context, objectRequest *model.CopyObjectRequest) error {
	err := u.repo.CopyObject(ctx, &model.CopyObjectRequest{
		OldKey: objectRequest.OldKey,
		NewKey: objectRequest.NewKey + filepath.Ext(objectRequest.OldKey),
	})
	if err != nil {
		utils.ErrorLog("usecase", "UpdateObject Repository", err)
		return err
	}

	err = u.repo.DeleteFile(ctx, objectRequest.OldKey)
	if err != nil {
		utils.ErrorLog("usecase", "UpdateObject DeleteFile", err)
		return err
	}

	return nil
}
