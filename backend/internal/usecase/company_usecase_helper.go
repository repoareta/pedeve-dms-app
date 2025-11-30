package usecase

import (
	"fmt"

	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/logger"
	"go.uber.org/zap"
)

// updateDescendantsLevel updates the level of all descendants recursively
// when a company's parent or level changes
// FIXED: Added max level limit (10) to prevent level from growing infinitely
func (uc *companyUseCase) updateDescendantsLevel(companyID string) error {
	zapLog := logger.GetLogger()
	
	// Use recursive approach: get all descendants and update their levels
	maxIterations := 10
	maxLevel := 10 // Maximum allowed level to prevent infinite growth
	
	for i := 0; i < maxIterations; i++ {
		descendants, err := uc.companyRepo.GetDescendants(companyID)
		if err != nil {
			zapLog.Error("Failed to get descendants", zap.Error(err))
			return fmt.Errorf("failed to get descendants: %w", err)
		}
		
		if len(descendants) == 0 {
			break // No descendants to update
		}
		
		// Update each descendant's level based on its parent
		updated := 0
		for _, desc := range descendants {
			// Get parent to determine correct level
			if desc.ParentID == nil {
				// If no parent, should be level 0 (holding)
				if desc.Level != 0 {
					desc.Level = 0
					if err := uc.companyRepo.Update(&desc); err != nil {
						zapLog.Warn("Failed to update descendant level",
							zap.String("descendant_id", desc.ID),
							zap.Error(err),
						)
						continue
					}
					updated++
				}
				continue
			}
			
			parent, err := uc.companyRepo.GetByID(*desc.ParentID)
			if err != nil {
				zapLog.Warn("Failed to get parent for descendant",
					zap.String("descendant_id", desc.ID),
					zap.String("parent_id", *desc.ParentID),
					zap.Error(err),
				)
				continue
			}
			
			// Calculate expected level: parent level + 1
			expectedLevel := parent.Level + 1
			
			// Safety check: prevent level from exceeding max level
			if expectedLevel > maxLevel {
				zapLog.Warn("Calculated level exceeds maximum, capping at max level",
					zap.String("descendant_id", desc.ID),
					zap.String("descendant_name", desc.Name),
					zap.Int("calculated_level", expectedLevel),
					zap.Int("max_level", maxLevel),
					zap.Int("parent_level", parent.Level),
				)
				expectedLevel = maxLevel
			}
			
			// Only update if level is incorrect
			if desc.Level != expectedLevel {
				oldLevel := desc.Level
				desc.Level = expectedLevel
				if err := uc.companyRepo.Update(&desc); err != nil {
					zapLog.Warn("Failed to update descendant level",
						zap.String("descendant_id", desc.ID),
						zap.Error(err),
					)
					continue
				}
				updated++
				zapLog.Info("Updated descendant level",
					zap.String("descendant_id", desc.ID),
					zap.String("descendant_name", desc.Name),
					zap.Int("old_level", oldLevel),
					zap.Int("new_level", expectedLevel),
					zap.Int("parent_level", parent.Level),
				)
			}
		}
		
		if updated == 0 {
			break // No more updates needed
		}
		
		zapLog.Debug("Updated descendant levels",
			zap.String("company_id", companyID),
			zap.Int("updated", updated),
			zap.Int("iteration", i+1),
		)
	}
	
	return nil
}

