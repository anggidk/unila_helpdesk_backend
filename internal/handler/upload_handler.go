package handler

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"unila_helpdesk_backend/internal/util"

	"github.com/gin-gonic/gin"
)

const maxUploadSize = 5 << 20 // 5MB

type UploadHandler struct {
	baseURL   string
	uploadDir string
}

func NewUploadHandler(baseURL string) *UploadHandler {
	return &UploadHandler{
		baseURL:   strings.TrimRight(baseURL, "/"),
		uploadDir: "uploads",
	}
}

func (handler *UploadHandler) RegisterRoutes(public *gin.RouterGroup) {
	public.POST("/uploads", handler.upload)
	public.GET("/uploads/:name", handler.download)
}

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
	if err := os.MkdirAll(handler.uploadDir, 0o755); err != nil {
		respondError(c, http.StatusInternalServerError, "gagal menyiapkan direktori upload")
		return
	}

	extension := strings.ToLower(filepath.Ext(file.Filename))
	storedName := util.NewID(24) + extension
	targetPath := filepath.Join(handler.uploadDir, storedName)
	if err := c.SaveUploadedFile(file, targetPath); err != nil {
		respondError(c, http.StatusInternalServerError, "gagal menyimpan file")
		return
	}

	respondOK(c, gin.H{
		"id":   storedName,
		"path": storedName,
		"url":  handler.baseURL + "/uploads/" + storedName,
		"name": file.Filename,
		"size": file.Size,
	})
}

func (handler *UploadHandler) download(c *gin.Context) {
	filename := filepath.Base(strings.TrimSpace(c.Param("name")))
	if filename == "" || filename == "." {
		respondError(c, http.StatusBadRequest, "nama file wajib diisi")
		return
	}

	targetPath := filepath.Join(handler.uploadDir, filename)
	info, err := os.Stat(targetPath)
	if err != nil || info.IsDir() {
		respondError(c, http.StatusNotFound, "file tidak ditemukan")
		return
	}

	c.File(targetPath)
}
