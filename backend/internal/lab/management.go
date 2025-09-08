package lab

import (
	"context"
	"fmt"
	"time"

	"github.com/wcrum/labby/internal/interfaces"
	"github.com/wcrum/labby/internal/models"
)

// GetLab retrieves a lab by ID
func (s *Service) GetLab(labID string) (*models.Lab, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	lab, exists := s.labs[labID]
	if !exists {
		return nil, ErrLabNotFound
	}

	// Check if lab is expired
	if models.IsExpired(lab.EndsAt) {
		lab.Status = models.LabStatusExpired
	}

	return lab, nil
}

// GetLabsByOwner retrieves all labs for a specific owner
func (s *Service) GetLabsByOwner(ownerID string) ([]*models.Lab, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var labs []*models.Lab
	for _, lab := range s.labs {
		if lab.OwnerID == ownerID {
			// Check if lab is expired
			if models.IsExpired(lab.EndsAt) {
				lab.Status = models.LabStatusExpired
			}
			labs = append(labs, lab)
		}
	}

	return labs, nil
}

// GetAllLabs retrieves all labs (for admin purposes)
func (s *Service) GetAllLabs() []*models.Lab {
	s.mu.RLock()
	defer s.mu.RUnlock()

	labs := make([]*models.Lab, 0, len(s.labs))
	for _, lab := range s.labs {
		// Check if lab is expired
		if models.IsExpired(lab.EndsAt) {
			lab.Status = models.LabStatusExpired
		}
		labs = append(labs, lab)
	}

	return labs
}

// DeleteLab deletes a lab
func (s *Service) DeleteLab(labID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	lab, exists := s.labs[labID]
	if !exists {
		return ErrLabNotFound
	}

	// Create cleanup context
	cleanupCtx := &interfaces.CleanupContext{
		LabID:   labID,
		Context: context.Background(),
		Lab:     lab,
	}

	// Cleanup lab services
	if err := s.serviceManager.CleanupLabServices(cleanupCtx); err != nil {
		// Log error but continue with lab deletion
		fmt.Printf("Warning: Failed to cleanup lab services for lab %s: %v\n", labID, err)
	}

	// Cleanup progress tracking
	s.progressTracker.CleanupProgress(labID)

	// Remove the lab
	delete(s.labs, labID)

	return nil
}

// StopLab stops a lab (marks it as expired)
func (s *Service) StopLab(labID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	lab, exists := s.labs[labID]
	if !exists {
		return ErrLabNotFound
	}

	// Set lab status to expired
	lab.Status = models.LabStatusExpired
	lab.EndsAt = time.Now()
	lab.UpdatedAt = time.Now()

	// Perform cleanup of lab services
	cleanupCtx := &interfaces.CleanupContext{
		LabID:   labID,
		Context: context.Background(),
		Lab:     lab,
	}

	// Execute cleanup (don't fail the stop operation if cleanup fails)
	if err := s.serviceManager.CleanupLabServices(cleanupCtx); err != nil {
		// Log the cleanup error but don't fail the stop operation
		fmt.Printf("StopLab: Cleanup failed for lab %s: %v\n", labID, err)
	}

	return nil
}

// CleanupExpiredLabs removes expired labs (should be called periodically)
func (s *Service) CleanupExpiredLabs() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for labID, lab := range s.labs {
		if now.After(lab.EndsAt) {
			// Create cleanup context for expired lab
			cleanupCtx := &interfaces.CleanupContext{
				LabID:   labID,
				Context: context.Background(),
				Lab:     lab,
			}

			// Cleanup lab services before removing from memory
			if err := s.serviceManager.CleanupLabServices(cleanupCtx); err != nil {
				fmt.Printf("Warning: Failed to cleanup expired lab services for lab %s: %v\n", labID, err)
			}

			// Cleanup progress tracking
			s.progressTracker.CleanupProgress(labID)

			// Remove the lab from memory
			delete(s.labs, labID)
		}
	}
}

// CleanupFailedLabs removes labs that have been in error status for too long
func (s *Service) CleanupFailedLabs() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	// Clean up labs that have been in error status for more than 1 hour
	errorThreshold := time.Hour

	for labID, lab := range s.labs {
		if lab.Status == models.LabStatusError {
			timeSinceError := now.Sub(lab.UpdatedAt)
			if timeSinceError > errorThreshold {
				// Create cleanup context for failed lab
				cleanupCtx := &interfaces.CleanupContext{
					LabID:   labID,
					Context: context.Background(),
					Lab:     lab,
				}

				// Cleanup lab services before removing from memory
				if err := s.serviceManager.CleanupLabServices(cleanupCtx); err != nil {
					fmt.Printf("Warning: Failed to cleanup failed lab services for lab %s: %v\n", labID, err)
				}

				// Cleanup progress tracking
				s.progressTracker.CleanupProgress(labID)
				// Remove the lab
				delete(s.labs, labID)
			}
		}
	}
}

// StartCleanupScheduler starts a background task to clean up expired labs
func (s *Service) StartCleanupScheduler() {
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for range ticker.C {
			s.CleanupExpiredLabs()
			s.CleanupFailedLabs()
		}
	}()
}
