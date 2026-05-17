package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yuwanadev/mdm-backend/internal/models"
)

type GroupRepo struct {
	pool *pgxpool.Pool
}

func NewGroupRepo(pool *pgxpool.Pool) *GroupRepo {
	return &GroupRepo{pool: pool}
}

func (r *GroupRepo) Create(ctx context.Context, name string) (*models.Group, error) {
	var g models.Group
	err := r.pool.QueryRow(ctx,
		`INSERT INTO groups (name) VALUES ($1)
		 RETURNING id, name, created_at`,
		name,
	).Scan(&g.ID, &g.Name, &g.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *GroupRepo) GetAll(ctx context.Context) ([]models.Group, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, name, created_at FROM groups ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []models.Group
	for rows.Next() {
		var g models.Group
		if err := rows.Scan(&g.ID, &g.Name, &g.CreatedAt); err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, rows.Err()
}

func (r *GroupRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM groups WHERE id = $1`, id)
	return err
}
