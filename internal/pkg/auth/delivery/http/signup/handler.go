package signup

import (
	"context"
	"fmt"
	"github.com/go-park-mail-ru/2024_2_SaraFun/internal/models"
	generatedAuth "github.com/go-park-mail-ru/2024_2_SaraFun/internal/pkg/auth/delivery/grpc/gen"
	generatedPayments "github.com/go-park-mail-ru/2024_2_SaraFun/internal/pkg/payments/delivery/grpc/gen"
	generatedPersonalities "github.com/go-park-mail-ru/2024_2_SaraFun/internal/pkg/personalities/delivery/grpc/gen"
	"github.com/go-park-mail-ru/2024_2_SaraFun/internal/utils/consts"
	"github.com/mailru/easyjson"
	"go.uber.org/zap"
	"net/http"
	"time"
)

//go:generate mockgen -destination=./mocks/mock_UserService.go -package=sign_up_mocks . UserService

//go:generate mockgen -destination=./mocks/mock_SessionService.go -package=sign_up_mocks . SessionService

//go:generate mockgen -destination=./mocks/mock_ProfileService.go -package=sign_up_mocks . ProfileService

//go:generate easyjson --all handler.go

type SessionClient interface {
	CreateSession(ctx context.Context, in *generatedAuth.CreateSessionRequest) (*generatedAuth.CreateSessionResponse, error)
}

type PersonalitiesClient interface {
	RegisterUser(ctx context.Context,
		in *generatedPersonalities.RegisterUserRequest) (*generatedPersonalities.RegisterUserResponse, error)
	CheckUsernameExists(ctx context.Context,
		in *generatedPersonalities.CheckUsernameExistsRequest) (*generatedPersonalities.CheckUsernameExistsResponse, error)
	CreateProfile(ctx context.Context,
		in *generatedPersonalities.CreateProfileRequest) (*generatedPersonalities.CreateProfileResponse, error)
}

//easyjson:skip
type Handler struct {
	personalitiesClient generatedPersonalities.PersonalitiesClient
	sessionClient       generatedAuth.AuthClient
	paymentsClient      generatedPayments.PaymentClient
	logger              *zap.Logger
}

type Request struct {
	User      models.User
	Profile   models.Profile
	Username  string `json:"username"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Age       int    `json:"age"`
	BirthDate string `json:"birth_date"`
	Gender    string `json:"gender"`
}

func NewHandler(personalitiesClient generatedPersonalities.PersonalitiesClient,
	sessionsClient generatedAuth.AuthClient, paymentsClient generatedPayments.PaymentClient,
	logger *zap.Logger) *Handler {
	return &Handler{
		personalitiesClient: personalitiesClient,
		sessionClient:       sessionsClient,
		paymentsClient:      paymentsClient,
		logger:              logger,
	}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	h.logger.Info("ctx", zap.Any("ctx", ctx))
	req_id := ctx.Value(consts.RequestIDKey).(string)
	h.logger.Info("Handling request", zap.String("request_id", req_id))
	if r.Method != http.MethodPost {
		h.logger.Error("bad method", zap.String("method", r.Method))
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	//request := Request{}
	//if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
	//	h.logger.Error("failed to decode request", zap.Error(err))
	//	http.Error(w, err.Error(), http.StatusBadRequest)
	//	return
	//}
	request := &Request{}
	err := easyjson.UnmarshalFromReader(r.Body, request)
	if err != nil {
		h.logger.Error("failed to parse request", zap.Error(err))
		http.Error(w, "Неверный формат данных", http.StatusBadRequest)
		return
	}
	h.logger.Info("request", zap.Any("request", request))
	user := models.User{
		Username: request.Username,
		Password: request.Password,
	}
	profile := models.Profile{
		FirstName:    request.FirstName,
		LastName:     request.LastName,
		Age:          request.Age,
		Gender:       request.Gender,
		BirthdayDate: request.BirthDate,
	}
	user.Sanitize()
	profile.Sanitize()
	h.logger.Info("user", zap.Any("user", user))
	h.logger.Info("profile", zap.Any("profile", profile))
	//personalitiesGRPC
	checkUsernameRequest := &generatedPersonalities.CheckUsernameExistsRequest{Username: user.Username}
	exists, err := h.personalitiesClient.CheckUsernameExists(ctx, checkUsernameRequest)
	if err != nil {
		h.logger.Error("failed to check username exists", zap.Error(err))
		http.Error(w, "Неудачная проверка на никнейм", http.StatusInternalServerError)
		return
	}
	if exists.Exists {
		h.logger.Error("user already exists", zap.String("username", user.Username))
		http.Error(w, "Пользователь с таким никнеймом уже существует", http.StatusBadRequest)
		return
	}
	h.logger.Info("exists", zap.Any("exists", exists))
	//personalitiesGRPC
	genProfile := &generatedPersonalities.Profile{ID: int32(request.Profile.ID),
		FirstName: profile.FirstName,
		LastName:  profile.LastName,
		Age:       int32(profile.Age),
		Gender:    profile.Gender,
		Target:    profile.Target,
		About:     profile.About,
		BirthDate: profile.BirthdayDate,
	}
	createProfileRequest := &generatedPersonalities.CreateProfileRequest{Profile: genProfile}
	profileId, err := h.personalitiesClient.CreateProfile(ctx, createProfileRequest)
	if err != nil {
		h.logger.Error("failed to create Profile", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//h.logger.Info("hash start")
	user.Profile = int(profileId.ProfileId)
	//hashedPass, err := hashing.HashPassword(user.Password)
	//if err != nil {
	//	h.logger.Error("failed to hash password", zap.Error(err))
	//	http.Error(w, "bad password", http.StatusBadRequest)
	//	return
	//}
	//user.Password = hashedPass
	// personalities grpc
	genUser := &generatedPersonalities.User{
		ID:       int32(user.ID),
		Username: user.Username,
		Password: user.Password,
		Email:    user.Email,
		Profile:  int32(user.Profile),
	}
	h.logger.Info("ctx ctx ctx", zap.Any("ctx", ctx))
	registerUserRequest := &generatedPersonalities.RegisterUserRequest{User: genUser}
	id, err := h.personalitiesClient.RegisterUser(ctx, registerUserRequest)
	if err != nil {
		h.logger.Error("failed to register User", zap.Error(err))
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	user.ID = int(id.UserId)

	createActivityReq := &generatedPayments.CreateActivityRequest{UserID: id.UserId}
	_, err = h.paymentsClient.CreateActivity(ctx, createActivityReq)
	if err != nil {
		h.logger.Error("failed to create Activity", zap.Error(err))
		http.Error(w, "что-то пошло не так :(", http.StatusInternalServerError)
		return
	}

	createBalancesRequest := &generatedPayments.CreateBalancesRequest{
		UserID:          id.UserId,
		MoneyAmount:     0,
		DailyAmount:     consts.DailyLikeLimit,
		PurchasedAmount: 5,
	}
	_, err = h.paymentsClient.CreateBalances(ctx, createBalancesRequest)
	if err != nil {
		h.logger.Error("failed to create balances", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//auth grpc
	sessUser := &generatedAuth.User{
		ID:       int32(user.ID),
		Username: user.Username,
		Password: user.Password,
		Email:    user.Email,
		Profile:  int32(user.Profile),
	}
	createSessionRequest := &generatedAuth.CreateSessionRequest{
		User: sessUser,
	}
	if session, err := h.sessionClient.CreateSession(ctx, createSessionRequest); err != nil {
		h.logger.Error("failed to create session", zap.Error(err))
		http.Error(w, "Не удалось создать сессию", http.StatusInternalServerError)
		return
	} else {
		http.SetCookie(w, &http.Cookie{
			Name:     consts.SessionCookie,
			Value:    session.Session.SessionID,
			Expires:  time.Now().Add(time.Hour * 24),
			Secure:   false,
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
			Path:     "/",
		})
	}
	h.logger.Info("good signup", zap.String("username", user.Username))
	fmt.Fprintf(w, "Вы успешно зарегистрировались")
}
