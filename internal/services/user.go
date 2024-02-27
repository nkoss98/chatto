package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"scratch/api"
	"scratch/internal/authorization/session"
	db "scratch/internal/storage/database"
	"strconv"

	"golang.org/x/crypto/bcrypt"
)

var (
	UserNotFoundErr      = errors.New("user not found")
	UserExistErr         = errors.New("user with that email already exist")
	IncorrectPasswordErr = errors.New("incorrect credentials")
)

type AccountManager interface {
	CreateUser(ctx context.Context, model api.RegisterUserRequest) (int, error)
	Login(ctx context.Context, model api.LoginUserRequest) (api.LoginUserResponse, error)
	GetUser(ctx context.Context, id int) (api.GetUserResponse, error)
	CleanUserTable(ctx context.Context) error
	MigrationMessage(ctx context.Context) (string, error)
}

// AccountService is responsible for handling account creation, login and sessions.
type AccountService struct {
	db         db.Querier
	tokenMaker session.IdentityGenerator
	logger     slog.Logger
}

func NewAccountService(db db.Querier, tokenGenerator session.IdentityGenerator, logger slog.Logger) *AccountService {
	return &AccountService{db: db, tokenMaker: tokenGenerator, logger: logger}
}

func (a *AccountService) CreateUser(ctx context.Context, model api.RegisterUserRequest) (int, error) {
	isExist, err := a.isUserExist(ctx, model.Email)
	if err != nil {
		return 0, fmt.Errorf("find user by email: %w", err)
	}
	if isExist {
		return 0, UserExistErr
	}

	pwd, err := a.hashAndSalt([]byte(model.Password))
	if err != nil {
		return 0, fmt.Errorf("problem to hash password: %w", err)
	}

	user, err := a.db.CreateUser(ctx, db.CreateUserParams{
		Name:     model.Name,
		Email:    model.Email,
		Password: pwd,
	})
	if err != nil {
		return 0, fmt.Errorf("create user: %w", err)
	}
	return int(user.ID), nil
}

func (a *AccountService) Login(ctx context.Context, model api.LoginUserRequest) (api.LoginUserResponse, error) {
	user, err := a.db.GetUserByEmail(ctx, model.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return api.LoginUserResponse{}, fmt.Errorf("login user: %w", UserNotFoundErr)
		}
		return api.LoginUserResponse{}, fmt.Errorf("login: %w", err)
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(model.Password))
	if err != nil {
		return api.LoginUserResponse{}, IncorrectPasswordErr
	}

	session, err := a.tokenMaker.GenerateTokens(strconv.Itoa(int(user.ID)))
	if err != nil {
		return api.LoginUserResponse{}, fmt.Errorf("generate token: %w", err)
	}

	return api.LoginUserResponse{
		Token:        session.Token,
		RefreshToken: session.RefreshToken,
	}, nil
}

func (a *AccountService) GetUser(ctx context.Context, id int) (api.GetUserResponse, error) {
	// add midelware here
	return api.GetUserResponse{}, nil
}

func (a *AccountService) CleanUserTable(ctx context.Context) error {
	return a.db.CleanUserTable(ctx)
}

func (a *AccountService) MigrationMessage(ctx context.Context) (string, error) {
	return a.db.MigrationMessage(ctx)

}

func (a *AccountService) hashAndSalt(pwd []byte) (string, error) {
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func (a *AccountService) isUserExist(ctx context.Context, email string) (bool, error) {
	_, err := a.db.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("can not get user by email: %w", err)
	}

	return true, nil
}
