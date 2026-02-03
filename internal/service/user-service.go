package service

import (
	"context"
	"errors"
	"fmt"

	"e-commerce.com/internal/config"
	"e-commerce.com/internal/db"
	"e-commerce.com/internal/models"
	"e-commerce.com/internal/repository"
	"e-commerce.com/internal/utils"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	//get
	GetUserByEmail(email string) (*models.User, error)

	VerifyUserToken(token string, ctx context.Context) (*models.User, error)

	GetUserSession(ctx context.Context, key string) (*models.User, error)

	DeleteUserSession(ctx context.Context, userId string) error

	//post
	Register(ctx context.Context, user *models.User) (string, error)
	Login(ctx context.Context, userLogin *models.UserLogin) (*models.User, error)

	// update

	// delete
}

type userService struct {
	repo repository.UserRepo
}

func NewUserService(repo repository.UserRepo) UserService {
	return &userService{repo: repo}
}

func (s *userService) Login(ctx context.Context, userLogin *models.UserLogin) (*models.User, error) {
	user, err := s.repo.LoginUser(ctx, userLogin)
	if err != nil {
		return nil, fmt.Errorf("failed to login: %w", err)
	}
	return user, nil
}

func (s *userService) DeleteUserSession(ctx context.Context, userId string) error {
	if err := s.repo.DeleteUserSessionFromRedis(ctx, userId); err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}

func (s *userService) GetUserSession(ctx context.Context, key string) (*models.User, error) {
	userData, err := s.repo.GetUserSessionFromRedis(ctx, key)

	if err != nil {
		return nil, err
	}
	return userData, nil
}

func (s *userService) VerifyUserToken(token string, ctx context.Context) (*models.User, error) {
	userData, err := s.repo.VerifyTokenFromRedis(token, ctx)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return userData, nil
}

func (s *userService) Register(ctx context.Context, user *models.User) (string, error) {
	redisClient, err := db.GetRedisClient()
	if err != nil {
		return "", fmt.Errorf("failed to connect database")
	}
	existingUser, err := s.repo.GetUserByEmail(user.Email)

	if err != nil {
		return "", fmt.Errorf("failed to check existing user : %w", err)
	}

	if existingUser != nil && existingUser.IsVerified {
		return "", errors.New("email already registered")
	}

	if existingUser != nil && !existingUser.IsVerified {
		//send email
		token, err := utils.SaveVerificationToken(ctx, redisClient, existingUser.Email)
		if err != nil {
			return "", fmt.Errorf("failed to create token : %v", err)
		}

		verificationLink := fmt.Sprintf("%s/verify-token?token=%s", config.AppConfig.FrontEndUrl, token)
		mailError := utils.SendVerificationEmail(existingUser.Email, verificationLink)

		if mailError != nil {
			return "", fmt.Errorf("failed to send verificaiton message  : %v", mailError)
		}

		return "User exists but not verified. Verification email resent.", nil
	}

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)

	if err != nil {
		return "", fmt.Errorf("failed to hash password : %w", err)
	}

	user.Password = string(hashPassword)
	user.IsVerified = false
	user.ID = uuid.New().String()

	fmt.Println("this is new user data : ", user)

	if err := s.repo.CreateUserDataInRedis(user, ctx); err != nil {
		return "", fmt.Errorf("failed to create user : %w", err)
	}

	token, err := utils.SaveVerificationToken(ctx, redisClient, user.Email)
	if err != nil {
		return "", fmt.Errorf("failed to create token : %v", err)
	}

	verificationLink := fmt.Sprintf("%s/verify-token?token=%s", config.AppConfig.FrontEndUrl, token)
	mailError := utils.SendVerificationEmail(user.Email, verificationLink)

	if mailError != nil {
		return "", fmt.Errorf("failed to send verificaiton message  : %v", mailError)
	}

	return "verify in  gmail", nil
}

func (s *userService) GetUserByEmail(email string) (*models.User, error) {
	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}
