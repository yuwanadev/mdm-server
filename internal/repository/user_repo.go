package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yuwanadev/mdm-backend/internal/models"
)

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

// GetByUsername finds a user by username.
func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	var u models.User
	err := r.pool.QueryRow(ctx,
		`SELECT id, username, password_hash, created_at
		 FROM users WHERE username = $1`, username,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// Create inserts a new user.
func (r *UserRepo) Create(ctx context.Context, username, passwordHash string) (*models.User, error) {
	var u models.User
	err := r.pool.QueryRow(ctx,
		`INSERT INTO users (username, password_hash)
		 VALUES ($1, $2)
		 RETURNING id, username, password_hash, created_at`,
		username, passwordHash,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// Exists checks if any user exists.
func (r *UserRepo) Exists(ctx context.Context) (bool, error) {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetByID finds a user by ID.
func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var u models.User
	err := r.pool.QueryRow(ctx,
		`SELECT id, username, password_hash, created_at
		 FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
