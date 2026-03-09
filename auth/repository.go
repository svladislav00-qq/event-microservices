package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/svladislav00-qq/event-microservices/pkg/logger/sl"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Repository interface {
	Close()
	PostAccount(ctx context.Context, a Account) error
	GetAccountById(ctx context.Context, id string) (*Account, error)
	GetAccountByEmail(ctx context.Context, email string) (*Account, error)
	GetAccounts(ctx context.Context, skip uint64, take uint64) ([]Account, error)
	PromoteToModerator(ctx context.Context, userID string, departmentId string) error
}

type postgresRepository struct {
	db  *gorm.DB
	log *slog.Logger
}

var (
	ErrUserExists   = errors.New("user already exists")
	ErrUserNotFound = errors.New("user not found")
)

func NewPostgresRepository(log *slog.Logger, databaseURL string) (Repository, error) {
	const op = "auth.repository.NewPostgresRepository"

	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = db.AutoMigrate(&Account{})
	if err != nil {
		return nil, fmt.Errorf("%s: failed to migrate: %w", op, err)
	}

	return &postgresRepository{
		db:  db,
		log: log,
	}, nil
}

func (r *postgresRepository) Close() {
	sqlDB, err := r.db.DB()
	if err != nil {
		return
	}

	sqlDB.Close()
}

func (r *postgresRepository) PostAccount(ctx context.Context, a Account) error {
	const op = "auth.repository.PostAccount"

	err := r.db.WithContext(ctx).Create(a).Error
	if err != nil {
		r.log.Error("failed to post account",
			slog.String("op", op),
			sl.Err(err),
		)
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *postgresRepository) GetAccountById(ctx context.Context, id string) (*Account, error) {
	const op = "auth.repository.GetAccountById"

	var account Account
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&account).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.log.Warn("user not found", slog.String("op", op), slog.String("id", id), sl.Err(err))
			return nil, fmt.Errorf("%s: %w", op, ErrUserNotFound)
		}
		r.log.Error("failed to get account by ID",
			slog.String("op", op),
			slog.String("id", id),
			sl.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &account, nil
}

func (r *postgresRepository) GetAccountByEmail(ctx context.Context, email string) (*Account, error) {
	const op = "auth.repository.GetAccountByEmail"

	var account Account
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&account).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.log.Warn("user not found",
				slog.String("op", op),
				slog.String("email", email),
				sl.Err(err))
			return nil, fmt.Errorf("%s: %w", op, ErrUserNotFound)
		}
		r.log.Error("failed to get account by ID",
			slog.String("op", op),
			slog.String("email", email),
			sl.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &account, nil
}

func (r *postgresRepository) GetAccounts(ctx context.Context, skip uint64, take uint64) ([]Account, error) {
	const op = "auth.repository.GetAccounts"
	var accounts []Account

	if take == 0 {
		take = 10
	}

	result := r.db.WithContext(ctx).Limit(int(take)).Offset(int(skip)).Find(&accounts)
	if result.Error != nil {
		r.log.Error("failed to get accounts",
			slog.String("op", op),
			sl.Err(result.Error))
		return nil, fmt.Errorf("%s: %w", op, result.Error)
	}

	return accounts, nil
}

func (r *postgresRepository) PromoteToModerator(ctx context.Context, userID string, departmentId string) error {
	const op = "auth.repository.PromoteToModerator"

	err := r.db.WithContext(ctx).Model(&Account{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"role":          "moderator",
		"department_id": departmentId,
	}).Error

	if err != nil {
		r.log.Error("failed to promote user to moderator",
			slog.String("op", op),
			slog.String("user_id", userID),
			slog.String("department_id", departmentId),
			sl.Err(err),
		)
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
