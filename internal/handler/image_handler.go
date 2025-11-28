package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"watermark/internal/service"
	"watermark/pkg/logger"

	"github.com/gorilla/mux"
)

type ImageHandler struct {
	service *service.ImageService
	logger  *logger.Logger
}

func NewImageHandler(service *service.ImageService, logger *logger.Logger) *ImageHandler {
	return &ImageHandler{
		service: service,
		logger:  logger,
	}
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func (h *ImageHandler) GetImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	imageID := vars["id"]

	weight, err := strconv.ParseFloat(r.URL.Query().Get("weight"), 64)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid weight parameter")
		return
	}

	dimensions := r.URL.Query().Get("dimensions")
	if dimensions == "" {
		h.respondError(w, http.StatusBadRequest, "Missing dimensions parameter")
		return
	}

	imageData, err := h.service.ProcessImage(r.Context(), service.ProcessRequest{
		ImageID:    imageID,
		Weight:     weight,
		Dimensions: dimensions,
	})
	if err != nil {
		h.logger.Error("Failed to process image",
			"imageID", imageID,
			"error", err,
		)
		h.respondError(w, http.StatusInternalServerError, "Failed to process image")
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Length", strconv.Itoa(len(imageData)))
	w.Header().Set("Cache-Control", "public, max-age=604800")
	w.WriteHeader(http.StatusOK)
	w.Write(imageData)
}

func (h *ImageHandler) respondError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:   http.StatusText(code),
		Message: message,
	})
}
