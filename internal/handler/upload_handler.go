package handler

import (
    "io"
    "net/http"
    "strings"

    "unila_helpdesk_backend/internal/domain"
    "unila_helpdesk_backend/internal/repository"
    "unila_helpdesk_backend/internal/util"

    "github.com/gin-gonic/gin"
)

const maxUploadSize = 5 << 20 // 5MB

type UploadHandler struct {
    baseURL   string
    attachments *repository.AttachmentRepository
}

func NewUploadHandler(baseURL string, attachments *repository.AttachmentRepository) *UploadHandler {
    return &UploadHandler{
        baseURL:     strings.TrimRight(baseURL, "/"),
        attachments: attachments,
    }
}

func (handler *UploadHandler) RegisterRoutes(public *gin.RouterGroup) {
    public.POST("/uploads", handler.upload)
    public.GET("/uploads/:id", handler.download)
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
    opened, err := file.Open()
    if err != nil {
        respondError(c, http.StatusInternalServerError, "gagal membaca file")
        return
    }
    defer opened.Close()

    data, err := io.ReadAll(opened)
    if err != nil {
        respondError(c, http.StatusInternalServerError, "gagal membaca file")
        return
    }
    if int64(len(data)) > maxUploadSize {
        respondError(c, http.StatusBadRequest, "ukuran file maksimal 5MB")
        return
    }

    contentType := file.Header.Get("Content-Type")
    if contentType == "" {
        contentType = http.DetectContentType(data)
    }

    attachment := &domain.Attachment{
        ID:          util.NewUUID(),
        Filename:    file.Filename,
        ContentType: contentType,
        Size:        int64(len(data)),
        Data:        data,
    }
    if err := handler.attachments.Create(attachment); err != nil {
        respondError(c, http.StatusInternalServerError, "gagal menyimpan file")
        return
    }

    url := handler.baseURL + "/uploads/" + attachment.ID
    respondOK(c, gin.H{
        "id":   attachment.ID,
        "url":  url,
        "name": file.Filename,
        "size": attachment.Size,
    })
}

func (handler *UploadHandler) download(c *gin.Context) {
    id := strings.TrimSpace(c.Param("id"))
    if id == "" {
        respondError(c, http.StatusBadRequest, "id wajib diisi")
        return
    }
    attachment, err := handler.attachments.FindByID(id)
    if err != nil {
        respondError(c, http.StatusNotFound, "file tidak ditemukan")
        return
    }
    c.Header("Content-Disposition", "inline; filename=\""+attachment.Filename+"\"")
    c.Data(http.StatusOK, attachment.ContentType, attachment.Data)
}
