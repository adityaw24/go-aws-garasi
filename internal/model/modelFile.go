package model

import "mime/multipart"

type FileRequest struct {
	Title string                `json:"title" binding:"required"`
	File  *multipart.FileHeader `json:"file" binding:"required"`
}

type UpdateFileRequest struct {
	Key string `json:"key" binding:"required"`
	FileRequest
}

type DeleteFileRequest struct {
	Key string `json:"key" binding:"required"`
}

type CopyObjectRequest struct {
	OldKey string `json:"oldKey" binding:"required"`
	NewKey string `json:"newKey" binding:"required"`
}

type FileModel struct {
	Key   string `json:"key"`
	Title string `json:"title"`
	Url   string `json:"url"`
}
