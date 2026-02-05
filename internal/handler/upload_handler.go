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

func NewUploadHandler(baseURL string, uploadDir string) *UploadHandler {
    return &UploadHandler{
        baseURL:   strings.TrimRight(baseURL, "/"),
        uploadDir: uploadDir,
    }
}

func (handler *UploadHandler) RegisterRoutes(public *gin.RouterGroup) {
    public.POST("/uploads", handler.upload)
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
        respondError(c, http.StatusInternalServerError, "gagal menyiapkan folder upload")
        return
    }

    ext := filepath.Ext(file.Filename)
    filename := util.NewUUID() + ext
    path := filepath.Join(handler.uploadDir, filename)
    if err := c.SaveUploadedFile(file, path); err != nil {
        respondError(c, http.StatusInternalServerError, "gagal menyimpan file")
        return
    }

    url := handler.baseURL + "/uploads/" + filename
    respondOK(c, gin.H{
        "url":  url,
        "name": file.Filename,
        "size": file.Size,
    })
}
