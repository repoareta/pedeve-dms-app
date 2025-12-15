package repository

import (
	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/database"
	"gorm.io/gorm"
)

// DirectorPositionRepository interface untuk director position operations
type DirectorPositionRepository interface {
	Create(directorPosition *domain.DirectorPositionModel) error
	GetByID(id string) (*domain.DirectorPositionModel, error)
	GetByName(name string) (*domain.DirectorPositionModel, error)
	GetAll(includeInactive bool) ([]domain.DirectorPositionModel, error)
	GetActive() ([]domain.DirectorPositionModel, error)
	Update(directorPosition *domain.DirectorPositionModel) error
	SoftDelete(id string) error // Soft delete: set is_active = false
	CountUsage(id string) (int64, error) // Count directors using this position
	IncrementUsage(id string) error // Increment usage count
	DecrementUsage(id string) error // Decrement usage count
}

type directorPositionRepository struct {
	db *gorm.DB
}

// NewDirectorPositionRepositoryWithDB creates a new director position repository with injected DB (for testing)
func NewDirectorPositionRepositoryWithDB(db *gorm.DB) DirectorPositionRepository {
	return &directorPositionRepository{
		db: db,
	}
}

// NewDirectorPositionRepository creates a new director position repository with default DB
func NewDirectorPositionRepository() DirectorPositionRepository {
	return NewDirectorPositionRepositoryWithDB(database.GetDB())
}

func (r *directorPositionRepository) Create(directorPosition *domain.DirectorPositionModel) error {
	return r.db.Create(directorPosition).Error
}

func (r *directorPositionRepository) GetByID(id string) (*domain.DirectorPositionModel, error) {
	var directorPosition domain.DirectorPositionModel
	err := r.db.Where("id = ?", id).First(&directorPosition).Error
	if err != nil {
		return nil, err
	}
	return &directorPosition, nil
}

func (r *directorPositionRepository) GetByName(name string) (*domain.DirectorPositionModel, error) {
	var directorPosition domain.DirectorPositionModel
	err := r.db.Where("LOWER(name) = LOWER(?)", name).First(&directorPosition).Error
	if err != nil {
		return nil, err
	}
	return &directorPosition, nil
}

func (r *directorPositionRepository) GetAll(includeInactive bool) ([]domain.DirectorPositionModel, error) {
	var directorPositions []domain.DirectorPositionModel
	query := r.db
	if !includeInactive {
		query = query.Where("is_active = ?", true)
	}
	err := query.Order("name ASC").Find(&directorPositions).Error
	return directorPositions, err
}

func (r *directorPositionRepository) GetActive() ([]domain.DirectorPositionModel, error) {
	var directorPositions []domain.DirectorPositionModel
	err := r.db.Where("is_active = ?", true).Order("name ASC").Find(&directorPositions).Error
	return directorPositions, err
}

func (r *directorPositionRepository) Update(directorPosition *domain.DirectorPositionModel) error {
	return r.db.Save(directorPosition).Error
}

func (r *directorPositionRepository) SoftDelete(id string) error {
	return r.db.Model(&domain.DirectorPositionModel{}).Where("id = ?", id).Update("is_active", false).Error
}

func (r *directorPositionRepository) CountUsage(id string) (int64, error) {
	// First get the director position name
	directorPosition, err := r.GetByID(id)
	if err != nil {
		return 0, err
	}
	
	var count int64
	// Count directors that use this position (check if position string contains the position name)
	// Position is stored as comma-separated string in DirectorModel.Position
	err = r.db.Model(&domain.DirectorModel{}).
		Where("position LIKE ?", "%"+directorPosition.Name+"%"). // Check if position string contains the name
		Count(&count).Error
	return count, err
}

func (r *directorPositionRepository) IncrementUsage(id string) error {
	return r.db.Model(&domain.DirectorPositionModel{}).
		Where("id = ?", id).
		UpdateColumn("usage_count", gorm.Expr("usage_count + 1")).Error
}

func (r *directorPositionRepository) DecrementUsage(id string) error {
	return r.db.Model(&domain.DirectorPositionModel{}).
		Where("id = ?", id).
		UpdateColumn("usage_count", gorm.Expr("GREATEST(usage_count - 1, 0)")).Error
}

