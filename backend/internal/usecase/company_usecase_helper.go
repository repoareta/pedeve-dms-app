package usecase

import (
	"fmt"

	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/logger"
	"go.uber.org/zap"
)

// updateDescendantsLevel updates the level of all descendants recursively
// when a company's parent or level changes
func (uc *companyUseCase) updateDescendantsLevel(companyID string) error {
	zapLog := logger.GetLogger()
	
	// Use recursive approach: get all descendants and update their levels
	maxIterations := 10
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
			
			expectedLevel := parent.Level + 1
			if desc.Level != expectedLevel {
				desc.Level = expectedLevel
				if err := uc.companyRepo.Update(&desc); err != nil {
					zapLog.Warn("Failed to update descendant level",
						zap.String("descendant_id", desc.ID),
						zap.Error(err),
					)
					continue
				}
				updated++
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

