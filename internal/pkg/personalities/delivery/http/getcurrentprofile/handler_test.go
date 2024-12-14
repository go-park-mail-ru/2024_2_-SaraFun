package getcurrentprofile

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2024_2_SaraFun/internal/models"
	generatedAuth "github.com/go-park-mail-ru/2024_2_SaraFun/internal/pkg/auth/delivery/grpc/gen"
	authmocks "github.com/go-park-mail-ru/2024_2_SaraFun/internal/pkg/auth/delivery/grpc/gen/mocks"
	generatedPersonalities "github.com/go-park-mail-ru/2024_2_SaraFun/internal/pkg/personalities/delivery/grpc/gen"
	personalitiesmocks "github.com/go-park-mail-ru/2024_2_SaraFun/internal/pkg/personalities/delivery/grpc/gen/mocks"
	imageservicemocks "github.com/go-park-mail-ru/2024_2_SaraFun/internal/pkg/personalities/delivery/http/getcurrentprofile/mocks"
	"github.com/go-park-mail-ru/2024_2_SaraFun/internal/utils/consts"
	"github.com/golang/mock/gomock"
	"go.uber.org/zap"
)

// 1. Успешное получение профиля при валидной сессии.
// 2. Отсутствие cookie сессии, что приводит к ошибке авторизации.
// 3. Ошибка при получении изображений пользователя.
// 4. Ошибка при получении данных профиля пользователя.

func TestHandler(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	ctx = context.WithValue(ctx, consts.RequestIDKey, "sparkit")
	logger := zap.NewNop()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	sessionClient := authmocks.NewMockAuthClient(mockCtrl)
	personalitiesClient := personalitiesmocks.NewMockPersonalitiesClient(mockCtrl)
	imageService := imageservicemocks.NewMockImageService(mockCtrl)

	handler := NewHandler(imageService, personalitiesClient, sessionClient, logger)

	tests := []struct {
		name                     string
		cookieValue              string
		userID                   int32
		userIDError              error
		profileID                int32
		profileIDError           error
		images                   []models.Image
		imagesError              error
		profile                  *generatedPersonalities.Profile
		profileError             error
		expectedStatus           int
		expectedResponseContains string
	}{
		{
			name:                     "good test",
			cookieValue:              "valid_session",
			userID:                   10,
			userIDError:              nil,
			profileID:                100,
			profileIDError:           nil,
			images:                   []models.Image{{Link: "http://example.com/img1.jpg"}, {Link: "http://example.com/img2.jpg"}},
			imagesError:              nil,
			profile:                  &generatedPersonalities.Profile{ID: 100, FirstName: "John", LastName: "Doe", Age: 30, Gender: "male", Target: "friendship", About: "Hello there!"},
			profileError:             nil,
			expectedStatus:           http.StatusOK,
			expectedResponseContains: `"first_name":"John"`,
		},
		{
			name:                     "no cookie",
			cookieValue:              "",
			userIDError:              nil,
			expectedStatus:           http.StatusUnauthorized,
			expectedResponseContains: "session not found",
		},

		{
			name:                     "error getting images",
			cookieValue:              "valid_session",
			userID:                   10,
			userIDError:              nil,
			profileID:                100,
			imagesError:              errors.New("image error"),
			expectedStatus:           http.StatusInternalServerError,
			expectedResponseContains: "image error",
		},
		{
			name:                     "error getting profile",
			cookieValue:              "valid_session",
			userID:                   10,
			profileID:                100,
			userIDError:              nil,
			images:                   []models.Image{},
			profileError:             errors.New("profile error"),
			expectedStatus:           http.StatusInternalServerError,
			expectedResponseContains: "profile error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.cookieValue != "" {
				getUserReq := &generatedAuth.GetUserIDBySessionIDRequest{SessionID: tt.cookieValue}
				if tt.userIDError == nil {
					userResp := &generatedAuth.GetUserIDBYSessionIDResponse{UserId: tt.userID}
					sessionClient.EXPECT().GetUserIDBySessionID(gomock.Any(), getUserReq).
						Return(userResp, nil).Times(1)
				} else {
					sessionClient.EXPECT().GetUserIDBySessionID(gomock.Any(), getUserReq).
						Return(nil, tt.userIDError).Times(1)
				}

				if tt.userIDError == nil {
					getProfileIDReq := &generatedPersonalities.GetProfileIDByUserIDRequest{UserID: tt.userID}
					if tt.profileIDError == nil {
						profileIDResp := &generatedPersonalities.GetProfileIDByUserIDResponse{ProfileID: tt.profileID}
						personalitiesClient.EXPECT().GetProfileIDByUserID(gomock.Any(), getProfileIDReq).
							Return(profileIDResp, nil).Times(1)
					} else {
						personalitiesClient.EXPECT().GetProfileIDByUserID(gomock.Any(), getProfileIDReq).
							Return(nil, tt.profileIDError).Times(1)
					}

					if tt.profileIDError == nil {
						if tt.imagesError == nil {
							imageService.EXPECT().GetImageLinksByUserId(gomock.Any(), int(tt.userID)).
								Return(tt.images, nil).Times(1)
						} else {
							imageService.EXPECT().GetImageLinksByUserId(gomock.Any(), int(tt.userID)).
								Return(nil, tt.imagesError).Times(1)
						}
					}

					if tt.profileIDError == nil && tt.imagesError == nil {
						getProfileReq := &generatedPersonalities.GetProfileRequest{Id: tt.profileID}
						if tt.profileError == nil {
							resp := &generatedPersonalities.GetProfileResponse{
								Profile: tt.profile,
							}
							personalitiesClient.EXPECT().GetProfile(gomock.Any(), getProfileReq).
								Return(resp, nil).Times(1)
						} else {
							personalitiesClient.EXPECT().GetProfile(gomock.Any(), getProfileReq).
								Return(nil, tt.profileError).Times(1)
						}
					}
				}
			}

			req := httptest.NewRequest(http.MethodGet, "/profile", nil)
			req = req.WithContext(ctx)
			if tt.cookieValue != "" {
				cookie := &http.Cookie{
					Name:  consts.SessionCookie,
					Value: tt.cookieValue,
				}
				req.AddCookie(cookie)
			}
			w := httptest.NewRecorder()

			handler.Handle(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", w.Code, tt.expectedStatus)
			}
			if tt.expectedResponseContains != "" && !contains(w.Body.String(), tt.expectedResponseContains) {
				t.Errorf("handler returned unexpected body: got %v want substring %v", w.Body.String(), tt.expectedResponseContains)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > 0 && len(substr) > 0 && string(s[0:len(substr)]) == substr) || (len(s) > len(substr) && string(s[len(s)-len(substr):]) == substr) || (len(substr) > 0 && len(s) > len(substr) && findInString(s, substr)))
}

func findInString(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
