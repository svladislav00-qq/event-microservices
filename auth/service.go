package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/segmentio/ksuid"
	"github.com/svladislav00-qq/event-microservices/pkg/logger/sl"
	"github.com/svladislav00-qq/event-microservices/pkg/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Auth struct {
	log         *slog.Logger
	usrSaver    UserSaver
	usrProvider UserProvider
	tokenTTL    time.Duration
}

type Claims struct {
	UserId     string `json:"user_id"`
	Username   string `json:"username"`
	Role       string `json:"role"`
	Department string `json:"department,omitempty"`
	jwt.RegisteredClaims
}

type UserSaver interface {
	PostAccount(ctx context.Context, a Account) error
}

type UserProvider interface {
	GetAccountByEmail(ctx context.Context, email string) (*Account, error)
	GetAccountById(ctx context.Context, id string) (*Account, error)
	PromoteToModerator(ctx context.Context, userID string, departmentId string) error
	GetAccounts(ctx context.Context, skip uint64, take uint64) ([]Account, error)
	GetUsersByIDs(ctx context.Context, ids []string) ([]models.Account, error)
}

type Account struct {
	ID           string         `json:"id" gorm:"primaryKey"`
	Email        string         `json:"email" gorm:"uniqueIndex;not null"`
	PasswordHash string         `json:"-"`
	Username     string         `json:"username"`
	Role         string         `json:"role"`
	Department   string         `json:"department,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at"`
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

func New(log *slog.Logger, userSaver UserSaver, userProvider UserProvider, tokenTTL time.Duration) *Auth {
	return &Auth{
		usrSaver:    userSaver,
		usrProvider: userProvider,
		log:         log,
		tokenTTL:    tokenTTL,
	}
}

func (a *Auth) Register(ctx context.Context, email, password, username string) (*Account, error) {
	const op = "auth.service.Register"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	log.Info("registering user")

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate password hash", sl.Err(err))

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	role := "user"

	res := &Account{
		ID:           ksuid.New().String(),
		Email:        email,
		PasswordHash: string(passwordHash),
		Username:     username,
		Role:         role,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := a.usrSaver.PostAccount(ctx, *res); err != nil {
		if errors.Is(err, ErrUserExists) {
			log.Warn("user already exists", sl.Err(err))
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		log.Error("failed to save user", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return res, nil
}

func (a *Auth) Login(ctx context.Context, email, password string) (string, error) {
	const op = "auth.service.Login"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	log.Info("attempting to login user")

	user, err := a.usrProvider.GetAccountByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			log.Warn("user not found", sl.Err(err))
			return "", fmt.Errorf("%s: %w", op, ErrUserNotFound)
		}
		log.Error("failed to get user", sl.Err(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		log.Info("invalid credentials", sl.Err(err))
		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	log.Info("user logged successfully")

	claims := Claims{
		UserId:     user.ID,
		Username:   user.Username,
		Role:       user.Role,
		Department: user.Department,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(a.tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		log.Error("failed to sign token", slog.String("op", op), sl.Err(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return signedToken, nil
}

func (a *Auth) PromoteToModerator(ctx context.Context, userID string, departmentId string) (*Account, error) {
	const op = "auth.service.PromoteToModerator"

	log := a.log.With(
		slog.String("op", op),
		slog.String("userId", userID),
		slog.String("departmentId", departmentId),
	)

	log.Info("promoting user to moderator with department")

	err := a.usrProvider.PromoteToModerator(ctx, userID, departmentId)
	if err != nil {
		log.Error("can not promote user to moderator", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return a.usrProvider.GetAccountById(ctx, userID)
}
func (a *Auth) GetAccounts(ctx context.Context, skip uint64, take uint64) ([]Account, error) {
	const op = "auth.service.GetAccounts"

	log := a.log.With(
		slog.String("op", op),
		slog.Uint64("skip", skip),
		slog.Uint64("take", take),
	)

	log.Info("starting showing all accounts with pagination")

	res, err := a.usrProvider.GetAccounts(ctx, skip, take)
	if err != nil {
		log.Error("can not get accounts", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return res, nil
}

func (a *Auth) GetAccountById(ctx context.Context, userID string) (*Account, error) {
	const op = "auth.service.GetAccountById"

	log := a.log.With("op", op, "user_id", userID)

	acc, err := a.usrProvider.GetAccountById(ctx, userID)
	if err != nil {
		log.Error("failed to get account", "error", err)
		return nil, err
	}

	return acc, nil
}

func (a *Auth) GetUsersByIDs(ctx context.Context, ids []string) ([]models.Account, error) {
	const op = "auth.service.GetUsersByIDs"

	log := a.log.With(
		slog.String("op", op),
	)

	log.Info("getting accounts by ids")

	accounts, err := a.usrProvider.GetUsersByIDs(ctx, ids)
	if err != nil {
		log.Error("failed to get accounts by ids", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return accounts, nil
}
