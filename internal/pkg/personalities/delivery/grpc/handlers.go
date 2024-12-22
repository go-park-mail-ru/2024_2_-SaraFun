package personalitiesgrpc

import (
	"context"
	"fmt"
	"github.com/go-park-mail-ru/2024_2_SaraFun/internal/models"
	generatedPersonalities "github.com/go-park-mail-ru/2024_2_SaraFun/internal/pkg/personalities/delivery/grpc/gen"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//go:generate mockgen -destination=./mocks/mock_userUsecase.go -package=mocks . UserUsecase
type UserUsecase interface {
	GetFeedList(ctx context.Context, userId int, receivers []int) ([]models.User, error)
	RegisterUser(ctx context.Context, user models.User) (int, error)
	CheckPassword(ctx context.Context, username string, password string) (models.User, error)
	GetProfileIdByUserId(ctx context.Context, userId int) (int, error)
	GetUsernameByUserId(ctx context.Context, userId int) (string, error)
	GetUserIdByUsername(ctx context.Context, username string) (int, error)
	CheckUsernameExists(ctx context.Context, username string) (bool, error)
	ChangePassword(ctx context.Context, userId int, password string) error
}

//go:generate mockgen -destination=./mocks/mock_profileUsecase.go -package=mocks . ProfileUsecase
type ProfileUsecase interface {
	CreateProfile(ctx context.Context, profile models.Profile) (int, error)
	UpdateProfile(ctx context.Context, id int, profile models.Profile) error
	GetProfile(ctx context.Context, id int) (models.Profile, error)
	DeleteProfile(ctx context.Context, id int) error
}

type GrpcPersonalitiesHandler struct {
	generatedPersonalities.PersonalitiesServer
	userUC    UserUsecase
	profileUC ProfileUsecase
	logger    *zap.Logger
}

func NewGrpcPersonalitiesHandler(userUC UserUsecase, profileUC ProfileUsecase, logger *zap.Logger) *GrpcPersonalitiesHandler {
	return &GrpcPersonalitiesHandler{
		userUC:    userUC,
		profileUC: profileUC,
		logger:    logger,
	}
}

func (h *GrpcPersonalitiesHandler) GetFeedList(ctx context.Context,
	in *generatedPersonalities.GetFeedListRequest) (*generatedPersonalities.GetFeedListResponse, error) {
	userId := int(in.UserID)
	rec := in.Receivers
	receivers := make([]int, len(rec))
	for i, v := range rec {
		receivers[i] = int(v)
	}
	users, err := h.userUC.GetFeedList(ctx, userId, receivers)
	if err != nil {
		return nil, fmt.Errorf("grpc get feed list error: %w", err)
	}
	resUsers := make([]*generatedPersonalities.User, len(users))
	for i, u := range users {
		resUsers[i] = &generatedPersonalities.User{ID: int32(u.ID), Username: u.Username,
			Email: u.Email, Password: u.Password, Profile: int32(u.Profile)}
	}
	return &generatedPersonalities.GetFeedListResponse{Users: resUsers}, nil
}

func (h *GrpcPersonalitiesHandler) RegisterUser(ctx context.Context,
	in *generatedPersonalities.RegisterUserRequest) (*generatedPersonalities.RegisterUserResponse, error) {
	user := &models.User{
		ID:       int(in.User.ID),
		Username: in.User.Username,
		Email:    in.User.Email,
		Password: in.User.Password,
		Profile:  int(in.User.Profile),
	}
	userId, err := h.userUC.RegisterUser(ctx, *user)
	if err != nil {
		return nil, fmt.Errorf("grpc register user error: %w", err)
	}
	return &generatedPersonalities.RegisterUserResponse{UserId: int32(userId)}, nil
}

func (h *GrpcPersonalitiesHandler) CheckPassword(ctx context.Context,
	in *generatedPersonalities.CheckPasswordRequest) (*generatedPersonalities.CheckPasswordResponse, error) {
	username := in.GetUsername()
	password := in.GetPassword()

	user, err := h.userUC.CheckPassword(ctx, username, password)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid password")
	}
	res := &generatedPersonalities.User{
		ID:       int32(user.ID),
		Username: user.Username,
		Email:    user.Email,
		Password: user.Password,
		Profile:  int32(user.Profile),
	}

	return &generatedPersonalities.CheckPasswordResponse{User: res}, nil
}

func (h *GrpcPersonalitiesHandler) GetProfileIDByUserID(ctx context.Context,
	in *generatedPersonalities.GetProfileIDByUserIDRequest) (*generatedPersonalities.GetProfileIDByUserIDResponse, error) {
	userId := int(in.UserID)
	//req_id := ctx.Value(consts.RequestIDKey).(string)
	//fmt.Println(req_id)
	profileId, err := h.userUC.GetProfileIdByUserId(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("grpc get profile id by user id error: %w", err)
	}
	res := &generatedPersonalities.GetProfileIDByUserIDResponse{ProfileID: int32(profileId)}
	return res, nil
}

func (h *GrpcPersonalitiesHandler) GetUsernameByUserID(ctx context.Context,
	in *generatedPersonalities.GetUsernameByUserIDRequest) (*generatedPersonalities.GetUsernameByUserIDResponse, error) {
	h.logger.Info("test")
	userId := int(in.UserID)
	username, err := h.userUC.GetUsernameByUserId(ctx, userId)
	h.logger.Info("test2")
	if err != nil {
		return nil, fmt.Errorf("grpc get username by user id error: %w", err)
	}
	res := &generatedPersonalities.GetUsernameByUserIDResponse{Username: username}

	return res, nil
}

func (h *GrpcPersonalitiesHandler) GetUserIDByUsername(ctx context.Context,
	in *generatedPersonalities.GetUserIDByUsernameRequest) (*generatedPersonalities.GetUserIDByUsernameResponse, error) {
	username := in.GetUsername()
	userId, err := h.userUC.GetUserIdByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("grpc get user id by username error: %w", err)
	}
	res := &generatedPersonalities.GetUserIDByUsernameResponse{UserID: int32(userId)}
	return res, nil
}

func (h *GrpcPersonalitiesHandler) CheckUsernameExists(ctx context.Context,
	in *generatedPersonalities.CheckUsernameExistsRequest) (*generatedPersonalities.CheckUsernameExistsResponse, error) {
	username := in.GetUsername()
	//req_id := ctx.Value(consts.RequestIDKey)
	exists, err := h.userUC.CheckUsernameExists(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("grpc check username exists error: %w", err)
	}
	res := &generatedPersonalities.CheckUsernameExistsResponse{Exists: exists}
	return res, nil
}

func (h *GrpcPersonalitiesHandler) CreateProfile(ctx context.Context,
	in *generatedPersonalities.CreateProfileRequest) (*generatedPersonalities.CreateProfileResponse, error) {
	profile := models.Profile{
		ID:           int(in.Profile.ID),
		FirstName:    in.Profile.FirstName,
		LastName:     in.Profile.LastName,
		Age:          int(in.Profile.Age),
		Gender:       in.Profile.Gender,
		Target:       in.Profile.Target,
		About:        in.Profile.About,
		BirthdayDate: in.Profile.BirthDate,
	}

	profileId, err := h.profileUC.CreateProfile(ctx, profile)
	if err != nil {
		return nil, fmt.Errorf("grpc create profile error: %w", err)
	}
	res := &generatedPersonalities.CreateProfileResponse{ProfileId: int32(profileId)}
	return res, nil
}

func (h *GrpcPersonalitiesHandler) UpdateProfile(ctx context.Context,
	in *generatedPersonalities.UpdateProfileRequest) (*generatedPersonalities.UpdateProfileResponse, error) {
	id := int(in.Profile.ID)
	profile := models.Profile{
		ID:           int(in.Profile.ID),
		FirstName:    in.Profile.FirstName,
		LastName:     in.Profile.LastName,
		Age:          int(in.Profile.Age),
		Gender:       in.Profile.Gender,
		Target:       in.Profile.Target,
		About:        in.Profile.About,
		BirthdayDate: in.Profile.BirthDate,
	}
	h.logger.Info("in", zap.Any("profile", profile))
	h.logger.Info("profile", zap.Any("profile", profile))

	err := h.profileUC.UpdateProfile(ctx, id, profile)
	h.logger.Info("update profile error", zap.Error(err))
	if err != nil {
		return nil, fmt.Errorf("grpc update profile error: %w", err)
	}
	res := &generatedPersonalities.UpdateProfileResponse{}
	return res, nil
}

func (h *GrpcPersonalitiesHandler) GetProfile(ctx context.Context,
	in *generatedPersonalities.GetProfileRequest) (*generatedPersonalities.GetProfileResponse, error) {
	Id := int(in.Id)
	profile, err := h.profileUC.GetProfile(ctx, Id)
	if err != nil {
		return nil, fmt.Errorf("grpc get profile error: %w", err)
	}
	resProfile := &generatedPersonalities.Profile{ID: int32(profile.ID),
		FirstName: profile.FirstName,
		LastName:  profile.LastName,
		Age:       int32(profile.Age),
		Gender:    profile.Gender,
		Target:    profile.Target,
		About:     profile.About,
		BirthDate: profile.BirthdayDate,
	}
	res := &generatedPersonalities.GetProfileResponse{Profile: resProfile}
	return res, nil
}

func (h *GrpcPersonalitiesHandler) DeleteProfile(ctx context.Context,
	in *generatedPersonalities.DeleteProfileRequest) (*generatedPersonalities.DeleteProfileResponse, error) {
	Id := int(in.Id)
	err := h.profileUC.DeleteProfile(ctx, Id)
	if err != nil {
		return nil, fmt.Errorf("grpc delete profile error: %w", err)
	}
	res := &generatedPersonalities.DeleteProfileResponse{}
	return res, nil
}

func (h *GrpcPersonalitiesHandler) ChangePassword(ctx context.Context,
	in *generatedPersonalities.ChangePasswordRequest) (*generatedPersonalities.ChangePasswordResponse, error) {
	err := h.userUC.ChangePassword(ctx, int(in.UserID), in.Password)
	if err != nil {
		return nil, fmt.Errorf("Grpc change password error : %w", err)
	}
	return &generatedPersonalities.ChangePasswordResponse{}, nil
}
