package repository

import (
	"github.com/Fajarriswandi/dms-app/backend/internal/domain"
	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/database"
	"gorm.io/gorm"
)

// CompanyRepository interface untuk company operations
type CompanyRepository interface {
	Create(company *domain.CompanyModel) error
	GetByID(id string) (*domain.CompanyModel, error)
	GetByCode(code string) (*domain.CompanyModel, error)
	GetAll() ([]domain.CompanyModel, error)
	GetByParentID(parentID string) ([]domain.CompanyModel, error)
	GetChildren(companyID string) ([]domain.CompanyModel, error)
	GetDescendants(companyID string) ([]domain.CompanyModel, error) // Get all descendants (children, grandchildren, etc)
	GetAncestors(companyID string) ([]domain.CompanyModel, error)  // Get all ancestors (parent, grandparent, etc)
	Update(company *domain.CompanyModel) error
	Delete(id string) error
	IsDescendantOf(childID, parentID string) (bool, error) // Check if childID is descendant of parentID
}

type companyRepository struct {
	db *gorm.DB
}

// NewCompanyRepository creates a new company repository
func NewCompanyRepository() CompanyRepository {
	return &companyRepository{
		db: database.GetDB(),
	}
}

func (r *companyRepository) Create(company *domain.CompanyModel) error {
	return r.db.Create(company).Error
}

func (r *companyRepository) GetByID(id string) (*domain.CompanyModel, error) {
	var company domain.CompanyModel
	err := r.db.Where("id = ?", id).First(&company).Error
	if err != nil {
		return nil, err
	}
	return &company, nil
}

func (r *companyRepository) GetByCode(code string) (*domain.CompanyModel, error) {
	var company domain.CompanyModel
	err := r.db.Where("code = ?", code).First(&company).Error
	if err != nil {
		return nil, err
	}
	return &company, nil
}

func (r *companyRepository) GetAll() ([]domain.CompanyModel, error) {
	var companies []domain.CompanyModel
	err := r.db.Where("is_active = ?", true).Find(&companies).Error
	return companies, err
}

func (r *companyRepository) GetByParentID(parentID string) ([]domain.CompanyModel, error) {
	var companies []domain.CompanyModel
	err := r.db.Where("parent_id = ?", parentID).Find(&companies).Error
	return companies, err
}

func (r *companyRepository) GetChildren(companyID string) ([]domain.CompanyModel, error) {
	return r.GetByParentID(companyID)
}

// GetDescendants menggunakan recursive CTE untuk mendapatkan semua descendants
func (r *companyRepository) GetDescendants(companyID string) ([]domain.CompanyModel, error) {
	var descendants []domain.CompanyModel
	
	// PostgreSQL recursive CTE
	query := `
		WITH RECURSIVE descendants AS (
			-- Base case: direct children
			SELECT * FROM companies WHERE parent_id = ? AND is_active = true
			UNION ALL
			-- Recursive case: children of children
			SELECT c.* FROM companies c
			INNER JOIN descendants d ON c.parent_id = d.id
			WHERE c.is_active = true
		)
		SELECT * FROM descendants
	`
	
	err := r.db.Raw(query, companyID).Scan(&descendants).Error
	return descendants, err
}

// GetAncestors menggunakan recursive CTE untuk mendapatkan semua ancestors
func (r *companyRepository) GetAncestors(companyID string) ([]domain.CompanyModel, error) {
	var ancestors []domain.CompanyModel
	
	// PostgreSQL recursive CTE
	query := `
		WITH RECURSIVE ancestors AS (
			-- Base case: company itself
			SELECT * FROM companies WHERE id = ?
			UNION ALL
			-- Recursive case: parent of parent
			SELECT c.* FROM companies c
			INNER JOIN ancestors a ON c.id = a.parent_id
		)
		SELECT * FROM ancestors WHERE id != ?
	`
	
	err := r.db.Raw(query, companyID, companyID).Scan(&ancestors).Error
	return ancestors, err
}

func (r *companyRepository) Update(company *domain.CompanyModel) error {
	return r.db.Save(company).Error
}

func (r *companyRepository) Delete(id string) error {
	// Soft delete: set is_active = false
	return r.db.Model(&domain.CompanyModel{}).Where("id = ?", id).Update("is_active", false).Error
}

// IsDescendantOf checks if childID is a descendant of parentID
func (r *companyRepository) IsDescendantOf(childID, parentID string) (bool, error) {
	descendants, err := r.GetDescendants(parentID)
	if err != nil {
		return false, err
	}
	
	for _, desc := range descendants {
		if desc.ID == childID {
			return true, nil
		}
	}
	
	return false, nil
}

