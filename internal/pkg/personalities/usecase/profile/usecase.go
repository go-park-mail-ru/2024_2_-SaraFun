package profile

import (
	"context"
	"fmt"
	sparkiterrors "github.com/go-park-mail-ru/2024_2_SaraFun/internal/errors"
	"github.com/go-park-mail-ru/2024_2_SaraFun/internal/models"
	"go.uber.org/zap"
	"time"
)

//go:generate mockgen -destination=./mocks/mock_repository.go -package=mocks . Repository
type Repository interface {
	CreateProfile(ctx context.Context, profile models.Profile) (int, error)
	UpdateProfile(ctx context.Context, id int, profile models.Profile) error
	GetProfile(ctx context.Context, id int) (models.Profile, error)
	DeleteProfile(ctx context.Context, id int) error
}

type UseCase struct {
	repo   Repository
	logger *zap.Logger
}

func New(repo Repository, logger *zap.Logger) *UseCase {
	return &UseCase{repo: repo, logger: logger}
}

func (u *UseCase) CreateProfile(ctx context.Context, profile models.Profile) (int, error) {
	//req_id := ctx.Value(consts.RequestIDKey).(string)
	//u.logger.Info("usecase request-id", zap.String("request_id", req_id))
	age, err := GetAge(profile.BirthdayDate)
	if err != nil {
		return -1, fmt.Errorf("get age error: %w", err)
	}
	profile.Age = age
	err = checkAge(profile.Age)
	if err != nil {
		return -1, err
	}
	id, err := u.repo.CreateProfile(ctx, profile)
	if err != nil {
		u.logger.Error("create profile err", zap.Error(err))
		return 0, fmt.Errorf("create profile err: %w", err)
	}
	return id, nil
}

func (u *UseCase) UpdateProfile(ctx context.Context, id int, profile models.Profile) error {
	//req_id := ctx.Value(consts.RequestIDKey).(string)
	//u.logger.Info("usecase request-id", zap.String("request_id", req_id))
	u.logger.Info("update profile", zap.Any("profile", profile))
	age, err := GetAge(profile.BirthdayDate)
	if err != nil {
		return fmt.Errorf("get age error: %w", err)
	}
	u.logger.Info("update profile", zap.Any("age", age))
	profile.Age = age
	err = checkAge(profile.Age)
	if err != nil {
		return err
	}
	if err := u.repo.UpdateProfile(ctx, id, profile); err != nil {
		u.logger.Error("update profile err", zap.Error(err))
		return fmt.Errorf("update profile err: %w", err)
	}
	return nil
}

func (u *UseCase) GetProfile(ctx context.Context, id int) (models.Profile, error) {
	//req_id := ctx.Value(consts.RequestIDKey).(string)
	//u.logger.Info("usecase request-id", zap.String("request_id", req_id))
	res, err := u.repo.GetProfile(ctx, id)
	if err != nil {
		u.logger.Error("get profile err", zap.Error(err))
		return models.Profile{}, fmt.Errorf("get profile err: %w", err)
	}
	u.logger.Info("get profile", zap.Any("res", res))
	res.Age, err = GetAge(res.BirthdayDate)
	if err != nil {
		u.logger.Error("get profile err", zap.Error(err))
		return models.Profile{}, fmt.Errorf("get profile err: %w", err)
	}
	return res, nil
}

func (u *UseCase) DeleteProfile(ctx context.Context, id int) error {
	//req_id := ctx.Value(consts.RequestIDKey).(string)
	//u.logger.Info("usecase request-id", zap.String("request_id", req_id))
	if err := u.repo.DeleteProfile(ctx, id); err != nil {
		u.logger.Error("delete profile err", zap.Error(err))
		return fmt.Errorf("delete profile err: %w", err)
	}
	return nil
}

func GetAge(birthdayDate string) (int, error) {
	birthDate, err := time.Parse("2006-01-02", birthdayDate)
	if err != nil {
		return -1, fmt.Errorf("birth date format error: %w", err)
	}
	currentDate := time.Now()
	age := currentDate.Year() - birthDate.Year()
	if currentDate.YearDay() < birthDate.YearDay() {
		age--
	}
	return age, nil
}

func checkAge(age int) error {
	if age < 18 {
		return sparkiterrors.ErrSmallAge
	}
	if age > 120 {
		return sparkiterrors.ErrBigAge
	}
	return nil
}
