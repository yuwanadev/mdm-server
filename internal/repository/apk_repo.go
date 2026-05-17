package repository

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yuwanadev/mdm-backend/internal/models"
)

type APKRepo struct {
	pool *pgxpool.Pool
}

func NewAPKRepo(pool *pgxpool.Pool) *APKRepo {
	return &APKRepo{pool: pool}
}

func (r *APKRepo) Create(ctx context.Context, apk *models.APK) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO apks (package_name, version_name, version_code, file_path, file_size)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, created_at`,
		apk.PackageName, apk.VersionName, apk.VersionCode, apk.FilePath, apk.FileSize,
	).Scan(&apk.ID, &apk.CreatedAt)
}

func (r *APKRepo) List(ctx context.Context) ([]models.APK, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, package_name, version_name, version_code, file_path, file_size, created_at FROM apks ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apks []models.APK
	for rows.Next() {
		var apk models.APK
		if err := rows.Scan(&apk.ID, &apk.PackageName, &apk.VersionName, &apk.VersionCode, &apk.FilePath, &apk.FileSize, &apk.CreatedAt); err != nil {
			return nil, err
		}
		apks = append(apks, apk)
	}
	return apks, nil
}

func (r *APKRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.APK, error) {
	var apk models.APK
	err := r.pool.QueryRow(ctx,
		`SELECT id, package_name, version_name, version_code, file_path, file_size, created_at FROM apks WHERE id = $1`, id,
	).Scan(&apk.ID, &apk.PackageName, &apk.VersionName, &apk.VersionCode, &apk.FilePath, &apk.FileSize, &apk.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &apk, nil
}
