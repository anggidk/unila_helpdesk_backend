package handler

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

const maxUploadSize = 5 << 20 // 5MB

type UploadHandler struct{}

func NewUploadHandler(_ string) *UploadHandler {
	return &UploadHandler{}
}

func (handler *UploadHandler) RegisterRoutes(public *gin.RouterGroup) {
	public.POST("/uploads", handler.upload)
}

// upload membaca file dari multipart form, mengenkode sebagai base64 data URI,
// dan mengembalikan data URI langsung — tidak ada yang disimpan ke filesystem.
// Klien menyimpan nilai "url" ini (data URI) ke kolom lamp1/lamp2 tiket.
func (handler *UploadHandler) upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		respondError(c, http.StatusBadRequest, "file wajib diisi")
		return
	}
	if file.Size > maxUploadSize {
		respondError(c, http.StatusBadRequest, "ukuran file maksimal 5MB")
		return
	}

	src, err := file.Open()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "gagal membuka file")
		return
	}
	defer src.Close()

	data, err := io.ReadAll(src)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "gagal membaca file")
		return
	}

	// DetectContentType inspects the first 512 bytes
	mimeType := http.DetectContentType(data)
	dataURI := fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(data))

	respondOK(c, gin.H{
		"url":  dataURI,
		"name": file.Filename,
		"size": file.Size,
	})
}
