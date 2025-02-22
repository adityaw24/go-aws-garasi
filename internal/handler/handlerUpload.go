package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/adityaw24/go-aws-garasi/internal/model"
	"github.com/adityaw24/go-aws-garasi/internal/usecase"
	"github.com/adityaw24/go-aws-garasi/utils"
	"github.com/gin-gonic/gin"
)

type HandlerUpload interface {
	UploadFile(ctx *gin.Context)
	PreviewFile(ctx *gin.Context)
	ListObjects(ctx *gin.Context)
	UpdateFile(ctx *gin.Context)
	DeleteFile(ctx *gin.Context)
	UpdateObject(ctx *gin.Context)
}

type handlerUpload struct {
	usecases usecase.UsecaseUpload
}

func NewHandlerUpload(usecases usecase.UsecaseUpload) HandlerUpload {
	return &handlerUpload{
		usecases: usecases,
	}
}

func (h *handlerUpload) UploadFile(ctx *gin.Context) {
	fileHeader, err := ctx.FormFile("file")
	title := ctx.PostForm("title")

	if err != nil {
		if errors.Is(err, http.ErrNotMultipart) || errors.Is(err, http.ErrMissingBoundary) {
			utils.ErrorLog("handler", "UploadFile", err)
			utils.ErrorResp(ctx, http.StatusBadRequest, "content-Type header is not valid")
			return
		}
		if errors.Is(err, http.ErrMissingFile) {
			utils.ErrorLog("handler", "UploadFile", err)
			utils.ErrorResp(ctx, http.StatusBadRequest, "request did not contain a file")
			return
		}
		// if errors.Is(err, multipart.ErrMessageTooLarge) {
		// 	utils.ErrorResp(ctx, http.StatusRequestEntityTooLarge, "max byte to upload is 8mB")
		// 	return ""
		// }

		utils.ErrorLog("handler", "UploadFile", err)
		utils.ErrorResp(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	fileMRequest := model.FileRequest{
		Title: title,
		File:  fileHeader,
	}

	if fileMRequest.Title == "" {
		utils.ErrorLog("handler", "UploadFile", errors.New("title is required"))
		utils.ErrorResp(ctx, http.StatusBadRequest, "title is required")
		return
	}

	if fileMRequest.File == nil {
		utils.ErrorLog("handler", "UploadFile", errors.New("file is required"))
		utils.ErrorResp(ctx, http.StatusBadRequest, "file is required")
		return
	}

	listObjects, err := h.usecases.UploadFile(ctx, &fileMRequest, utils.ValidImageTypes)
	if err != nil {
		utils.ErrorLog("handler", "UploadFile", err)
		utils.ErrorResp(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResp(ctx, http.StatusOK, "success upload file", listObjects)
}

func (h *handlerUpload) PreviewFile(ctx *gin.Context) {
	objectKey := ctx.Param("key")
	if objectKey == "" {
		utils.ErrorLog("handler", "PreviewFile", errors.New("key parameter is required"))
		utils.ErrorResp(ctx, http.StatusBadRequest, "key parameter is required")
		return
	}

	presignedURL, err := h.usecases.PreviewFile(ctx, objectKey)
	if err != nil {
		utils.ErrorLog("handler", "PreviewFile", err)
		utils.ErrorResp(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResp(ctx, http.StatusOK, "success get preview url", gin.H{
		"url": presignedURL,
	})
}

func (h *handlerUpload) ListObjects(ctx *gin.Context) {
	objects, err := h.usecases.ListObjects(ctx)
	if err != nil {
		utils.ErrorLog("handler", "ListObjects", err)
		utils.ErrorResp(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResp(ctx, http.StatusOK, "success get list files", objects)
}

func (h *handlerUpload) UpdateFile(ctx *gin.Context) {
	var err error
	fileHeader, err := ctx.FormFile("file")
	title := ctx.PostForm("title")
	key := ctx.PostForm("key")

	if err != nil {
		if errors.Is(err, http.ErrNotMultipart) || errors.Is(err, http.ErrMissingBoundary) {
			utils.ErrorLog("handler", "UpdateFile", err)
			utils.ErrorResp(ctx, http.StatusBadRequest, "multipart form data is required")
			return
		}
		utils.ErrorLog("handler", "UpdateFile", err)
		utils.ErrorResp(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	if fileHeader == nil {
		utils.ErrorLog("handler", "UpdateFile", errors.New("file is required"))
		utils.ErrorResp(ctx, http.StatusBadRequest, "file is required")
		return
	}

	fileMRequest := model.UpdateFileRequest{
		Key: key,
		FileRequest: model.FileRequest{
			Title: title,
			File:  fileHeader,
		},
	}

	if fileMRequest.Key == "" {
		utils.ErrorLog("handler", "UpdateFile", errors.New("key is required"))
		utils.ErrorResp(ctx, http.StatusBadRequest, "key is required")
		return
	}

	if fileMRequest.Title == "" {
		utils.ErrorLog("handler", "UpdateFile", errors.New("title is required"))
		utils.ErrorResp(ctx, http.StatusBadRequest, "title is required")
		return
	}

	err = h.usecases.UpdateFile(ctx, &fileMRequest, utils.ValidImageTypes)
	if err != nil {
		utils.ErrorLog("handler", "UpdateFile", err)
		utils.ErrorResp(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResp(ctx, http.StatusOK, "success update file", nil)
}

func (h *handlerUpload) DeleteFile(ctx *gin.Context) {
	key := ctx.Param("key")

	if key == "" {
		utils.ErrorLog("handler", "DeleteFile", errors.New("key parameter is required"))
		utils.ErrorResp(ctx, http.StatusBadRequest, "key parameter is required")
		return
	}

	fileRequest := model.DeleteFileRequest{
		Key: key,
	}

	err := h.usecases.DeleteFile(ctx, &fileRequest)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.ErrorResp(ctx, http.StatusNotFound, err.Error())
			return
		}
		utils.ErrorLog("handler", "DeleteFile", err)
		utils.ErrorResp(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResp(ctx, http.StatusOK, "success delete file", nil)
}

func (h *handlerUpload) UpdateObject(ctx *gin.Context) {
	oldKey := ctx.PostForm("oldKey")
	newKey := ctx.PostForm("newKey")

	if oldKey == "" {
		utils.ErrorLog("handler", "UpdateObject", errors.New("oldKey is required"))
		utils.ErrorResp(ctx, http.StatusBadRequest, "oldKey is required")
		return
	}

	if newKey == "" {
		utils.ErrorLog("handler", "UpdateObject", errors.New("newKey is required"))
		utils.ErrorResp(ctx, http.StatusBadRequest, "newKey is required")
		return
	}

	err := h.usecases.UpdateObject(ctx, &model.CopyObjectRequest{
		OldKey: oldKey,
		NewKey: newKey,
	})
	if err != nil {
		utils.ErrorLog("handler", "UpdateObject", err)
		utils.ErrorResp(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResp(ctx, http.StatusOK, "success update object", nil)
}
