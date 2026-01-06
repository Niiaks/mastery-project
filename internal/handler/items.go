package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"mastery-project/internal/config"
	"mastery-project/internal/model"
	"mastery-project/internal/service"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
)

type ItemHandler struct {
	Handler
	ItemService *service.ItemService
}

func NewItemHandler(cfg *config.Config, itemService *service.ItemService) *ItemHandler {
	return &ItemHandler{
		Handler:     NewHandler(cfg.ENV),
		ItemService: itemService}
}

func (h *ItemHandler) Create(w http.ResponseWriter, r *http.Request) {
	const maxUploadSize = 5 << 20 // 5MB
	const uploadDir = "uploads"

	err := r.ParseMultipartForm(maxUploadSize)
	if err != nil {
		h.JSON(w, http.StatusBadRequest, "file too large")
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		h.JSON(w, http.StatusBadRequest, "file is required")
		return
	}
	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			h.JSON(w, http.StatusInternalServerError, "failed to close file")
			return
		}
	}(file)

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !isAllowedExtension(ext) {
		h.JSON(w, http.StatusForbidden, "file type not allowed")
		return
	}

	//Validate MIME type (sniffing)
	buffer := make([]byte, 512)
	if _, err := file.Read(buffer); err != nil && err != io.EOF {
		h.JSON(w, http.StatusInternalServerError, "failed to read file")
		return
	}

	mimeType := http.DetectContentType(buffer)
	if mimeType != "image/jpeg" && mimeType != "image/png" {
		h.JSON(w, http.StatusForbidden, "invalid file content")
		return
	}

	// reset reader
	if _, err := file.Seek(0, 0); err != nil {
		h.JSON(w, http.StatusInternalServerError, "failed to reset file")
		return
	}

	// Ensure uploads directory exists
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		h.JSON(w, http.StatusInternalServerError, "failed to create upload dir")
		return
	}

	// Generate secure filename
	uniqueName := generateSecureFilename(ext)
	dstPath := filepath.Join(uploadDir, uniqueName)

	dst, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		h.JSON(w, http.StatusInternalServerError, "failed to save file")
		return
	}
	defer func(dst *os.File) {
		err := dst.Close()
		if err != nil {
			h.JSON(w, http.StatusInternalServerError, "failed to save file")
			return
		}
	}(dst)

	// Copy with size enforcement
	if _, err := io.Copy(dst, io.LimitReader(file, maxUploadSize)); err != nil {
		h.JSON(w, http.StatusInternalServerError, "upload failed")
		return
	}

	user := r.Context().Value("user").(*model.User)
	//Read form fields
	item := model.Item{
		UserID:      user.ID,
		Title:       r.FormValue("title"),
		Description: r.FormValue("description"),
		FilePath:    uniqueName, // store ONLY filename in DB
	}

	if err := validate.Struct(item); err != nil {
		h.JSON(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.ItemService.Save(r.Context(), item); err != nil {
		h.JSON(w, http.StatusInternalServerError, "failed to create item")
		return
	}

	h.JSON(w, http.StatusCreated, item)
}
func (h *ItemHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	items, err := h.ItemService.GetAll(r.Context())
	if err != nil {
		h.JSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.JSON(w, http.StatusOK, items)
}
func (h *ItemHandler) GetOne(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	item, err := h.ItemService.GetOne(r.Context(), id)
	if err != nil {
		h.JSON(w, http.StatusNotFound, err.Error())
		return
	}

	h.JSON(w, http.StatusOK, item)
}
func (h *ItemHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	item, err := h.ItemService.GetOne(r.Context(), id)
	if err != nil {
		h.JSON(w, http.StatusNotFound, err.Error())
		return
	}

	// delete db record
	if err := h.ItemService.Delete(r.Context(), id); err != nil {
		h.JSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	// delete file
	_ = os.Remove(filepath.Join("uploads", item.FilePath))

	h.JSON(w, http.StatusOK, map[string]string{"message": "item deleted"})
}
func (h *ItemHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var item model.UpdateItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		h.JSON(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := validate.Struct(item); err != nil {
		h.JSON(w, http.StatusBadRequest, err.Error())
		return
	}

	err := h.ItemService.Update(r.Context(), id, item)
	if err != nil {
		h.JSON(w, http.StatusNotFound, err.Error())
		return
	}

	h.JSON(w, http.StatusOK, map[string]string{"message": "item updated"})
}
func (h *ItemHandler) ViewImage(w http.ResponseWriter, r *http.Request) {
	filename := chi.URLParam(r, "filename")

	// basic hardening
	if strings.Contains(filename, "..") {
		http.Error(w, "invalid filename", http.StatusBadRequest)
		return
	}

	path := filepath.Join("uploads", filename)

	file, err := os.Open(path)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer file.Close()

	// sniff content-type
	buffer := make([]byte, 512)
	n, _ := file.Read(buffer)
	contentType := http.DetectContentType(buffer[:n])

	// only images
	if contentType != "image/jpeg" && contentType != "image/png" {
		http.Error(w, "unsupported file type", http.StatusForbidden)
		return
	}

	file.Seek(0, 0)
	w.Header().Set("Content-Type", contentType)
	io.Copy(w, file)
}

func isAllowedExtension(fileName string) bool {
	allowedExtensions := []string{".jpg", ".jpeg", ".png"}
	ext := strings.ToLower(filepath.Ext(fileName))
	for _, allowed := range allowedExtensions {
		if ext == strings.ToLower(allowed) {
			return true
		}
	}
	return false
}

func generateSecureFilename(ext string) string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}
	return hex.EncodeToString(b) + ext
}
