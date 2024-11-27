package uploadimage

import (
	"context"
	"encoding/json"
	generatedAuth "github.com/go-park-mail-ru/2024_2_SaraFun/internal/pkg/auth/delivery/grpc/gen"
	"github.com/go-park-mail-ru/2024_2_SaraFun/internal/utils/consts"
	"go.uber.org/zap"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
)

//go:generate mockgen -destination=./mocks/mock_ImageService.go -package=uploadimage_mocks . ImageService
type ImageService interface {
	SaveImage(ctx context.Context, file multipart.File, fileExt string, userId int, ordNumber int) (int, error)
}

//go:generate mockgen -destination=./mocks/mock_SessionService.go -package=uploadimage_mocks . SessionService
type SessionService interface {
	GetUserIDBySessionID(ctx context.Context, sessionID string) (int, error)
}

type Response struct {
	ImageId int
}

type Handler struct {
	imageService   ImageService
	sessionService generatedAuth.AuthClient
	logger         *zap.Logger
}

func NewHandler(imageService ImageService, sessionService generatedAuth.AuthClient, logger *zap.Logger) *Handler {
	return &Handler{imageService: imageService, sessionService: sessionService, logger: logger}
}
func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req_id := ctx.Value(consts.RequestIDKey).(string)
	h.logger.Info("Handling request", zap.String("request_id", req_id))

	limitedReader := http.MaxBytesReader(w, r.Body, 32<<20)
	defer r.Body.Close()

	//bodyContent, err := io.ReadAll(limitedReader)
	//if err != nil && !errors.Is(err, io.EOF) {
	//	h.logger.Error("Error reading limited body", zap.Error(err))
	//	if errors.As(err, new(*http.MaxBytesError)) {
	//		http.Error(w, "request entity too large", http.StatusRequestEntityTooLarge)
	//		return
	//	}
	//}

	//err := json.NewDecoder(r.Body).Decode(&number)
	//if err != nil {
	//	h.logger.Error("Error parsing request body", zap.Error(err))
	//	http.Error(w, "Bad Request", http.StatusBadRequest)
	//	return
	//}
	r.Body = limitedReader
	//err := json.NewDecoder(limitedReader).Decode(&number)
	//if err != nil {
	//	h.logger.Error("Error parsing request body", zap.Error(err))
	//	http.Error(w, "Bad Request", http.StatusBadRequest)
	//	return
	//}
	//h.logger.Info("body content", zap.Binary("body content", bodyContent))
	//fileFormat := http.DetectContentType(bodyContent)
	//h.logger.Info("File format", zap.String("file_format", fileFormat))
	//if fileFormat != "image/png" && fileFormat != "image/jpeg" && fileFormat != "image/jpg" {
	//	h.logger.Error("Invalid image format", zap.String("request_id", req_id))
	//	http.Error(w, "invalid image format", http.StatusBadRequest)
	//	return
	//}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		h.logger.Error("parse multipart form", zap.Error(err))
		http.Error(w, "bad image", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		h.logger.Error("failed to parse multipart form", zap.Error(err))
		http.Error(w, "bad image file", http.StatusBadRequest)
		return
	}
	if header == nil {
		h.logger.Error("failed to parse multipart form")
		http.Error(w, "bad image file", http.StatusBadRequest)
		return
	}
	number := r.FormValue("number")
	//if err != nil {
	//	h.logger.Error("failed to parse multipart form", zap.Error(err))
	//	http.Error(w, "bad image number", http.StatusBadRequest)
	//	return
	//}

	fileHeader := make([]byte, 512)

	if _, err := file.Read(fileHeader); err != nil {
		http.Error(w, "bad image header", http.StatusBadRequest)
		return
	}

	if _, err := file.Seek(0, 0); err != nil {
		http.Error(w, "bad image header", http.StatusBadRequest)
		return
	}
	fileFormat := http.DetectContentType(fileHeader)
	h.logger.Info("File format", zap.String("file_format", fileFormat))
	if fileFormat != "image/png" && fileFormat != "image/jpeg" && fileFormat != "image/jpg" {
		h.logger.Error("Invalid image format", zap.String("request_id", req_id))
		http.Error(w, "invalid image format", http.StatusBadRequest)
		return
	}
	defer file.Close()

	fileExt := filepath.Ext(header.Filename)
	log.Print("good image file")

	cookie, err := r.Cookie(consts.SessionCookie)
	if err != nil {
		h.logger.Error("failed to get session cookie", zap.Error(err))
		http.Error(w, "session not found", http.StatusUnauthorized)
		return
	}
	log.Print("good session cookie")
	getUserRequest := &generatedAuth.GetUserIDBySessionIDRequest{SessionID: cookie.Value}
	userId, err := h.sessionService.GetUserIDBySessionID(ctx, getUserRequest)
	if err != nil {
		h.logger.Error("failed to get user id", zap.Error(err))
		http.Error(w, "user session err", http.StatusInternalServerError)
		return
	}
	num, err := strconv.Atoi(number)
	if err != nil {
		h.logger.Error("failed to convert number", zap.Error(err))
		http.Error(w, "invalid number", http.StatusBadRequest)
	}
	id, err := h.imageService.SaveImage(ctx, file, fileExt, int(userId.UserId), num)
	if err != nil {
		h.logger.Error("failed to save image", zap.Error(err))
		http.Error(w, "save image err", http.StatusInternalServerError)
		return
	}

	response := Response{ImageId: id}
	jsonData, err := json.Marshal(response)
	if err != nil {
		h.logger.Error("failed to marshal image", zap.Error(err))
		http.Error(w, "save image err", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(jsonData)
	if err != nil {
		h.logger.Error("failed to write response", zap.Error(err))
		http.Error(w, "save image err", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	h.logger.Info("image saved successfully")
}
