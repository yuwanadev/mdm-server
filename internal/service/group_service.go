package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/yuwanadev/mdm-backend/internal/models"
	"github.com/yuwanadev/mdm-backend/internal/repository"
)

type GroupService struct {
	groupRepo *repository.GroupRepo
}

func NewGroupService(groupRepo *repository.GroupRepo) *GroupService {
	return &GroupService{groupRepo: groupRepo}
}

func (s *GroupService) CreateGroup(ctx context.Context, name string) (*models.Group, error) {
	return s.groupRepo.Create(ctx, name)
}

func (s *GroupService) GetAllGroups(ctx context.Context) ([]models.Group, error) {
	return s.groupRepo.GetAll(ctx)
}

func (s *GroupService) DeleteGroup(ctx context.Context, id uuid.UUID) error {
	return s.groupRepo.Delete(ctx, id)
}
