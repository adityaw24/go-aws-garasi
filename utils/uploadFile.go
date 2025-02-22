package utils

import (
	"errors"
)

type Upload struct {
	Length      int64
	ContentType string
	Prefix      string
	Ext         string
}

func ValidateContentType(contentType string, validMimeTypes []string) (err error) {
	var isValidType bool
	for _, validMimeType := range validMimeTypes {
		if contentType == validMimeType {
			isValidType = true
			break
		}
	}
	if !isValidType {
		err = errors.New("not valid mime-type")
		return err
	}

	return nil
}
