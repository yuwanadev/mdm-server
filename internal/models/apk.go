package models

import (
	"time"
	"github.com/google/uuid"
)

type APK struct {
	ID          uuid.UUID `json:"id"`
	PackageName string    `json:"package_name"`
	VersionName string    `json:"version_name"`
	VersionCode int       `json:"version_code"`
	FilePath    string    `json:"file_path"`
	FileSize    int64     `json:"file_size"`
	CreatedAt   time.Time `json:"created_at"`
}
