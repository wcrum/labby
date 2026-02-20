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
	lab, err := s.repo.GetLabByID(labID)
	if err != nil {
		return nil, ErrLabNotFound
	}

	// Check if lab is expired
	if models.IsExpired(lab.EndsAt) {
		lab.Status = models.LabStatusExpired
		// Update the lab status in the database
		lab.UpdatedAt = time.Now()
		s.repo.UpdateLab(lab)
	}

	return lab, nil
}

// GetLabsByOwner retrieves all labs for a specific owner
func (s *Service) GetLabsByOwner(ownerID string) ([]*models.Lab, error) {
	labs, err := s.repo.GetLabsByOwnerID(ownerID)
	if err != nil {
		return nil, err
	}

	// Check if any labs are expired and update them
	for _, lab := range labs {
		if models.IsExpired(lab.EndsAt) && lab.Status != models.LabStatusExpired {
			lab.Status = models.LabStatusExpired
			lab.UpdatedAt = time.Now()
			s.repo.UpdateLab(lab)
		}
	}

	return labs, nil
}

// GetAllLabs retrieves all labs (for admin purposes)
func (s *Service) GetAllLabs() ([]*models.Lab, error) {
	labs, err := s.repo.GetAllLabs()
	if err != nil {
		return nil, err
	}

	// Check if any labs are expired and update them
	for _, lab := range labs {
		if models.IsExpired(lab.EndsAt) && lab.Status != models.LabStatusExpired {
			lab.Status = models.LabStatusExpired
			lab.UpdatedAt = time.Now()
			s.repo.UpdateLab(lab)
		}
	}

	return labs, nil
}

// DeleteLab deletes a lab
func (s *Service) DeleteLab(labID string) error {
	lab, err := s.repo.GetLabByID(labID)
	if err != nil {
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

	// Delete the lab from database
	return s.repo.DeleteLab(labID)
}

// StopLab stops a lab (marks it as expired)
func (s *Service) StopLab(labID string) error {
	lab, err := s.repo.GetLabByID(labID)
	if err != nil {
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

	// Update the lab in the database
	return s.repo.UpdateLab(lab)
}

// CleanupExpiredLabs removes expired labs (should be called periodically)
func (s *Service) CleanupExpiredLabs() {
	// Get expired labs from database
	expiredLabs, err := s.repo.GetExpiredLabs()
	if err != nil {
		fmt.Printf("CleanupExpiredLabs: Failed to get expired labs: %v\n", err)
		return
	}

	for _, lab := range expiredLabs {
		// Create cleanup context for expired lab
		cleanupCtx := &interfaces.CleanupContext{
			LabID:   lab.ID,
			Context: context.Background(),
			Lab:     lab,
		}

		// Cleanup lab services
		if err := s.serviceManager.CleanupLabServices(cleanupCtx); err != nil {
			fmt.Printf("CleanupExpiredLabs: Failed to cleanup lab %s: %v\n", lab.ID, err)
		}

		// Cleanup progress tracking
		s.progressTracker.CleanupProgress(lab.ID)

		// Update lab status to expired
		lab.Status = models.LabStatusExpired
		lab.UpdatedAt = time.Now()
		if err := s.repo.UpdateLab(lab); err != nil {
			fmt.Printf("CleanupExpiredLabs: Failed to update lab %s status: %v\n", lab.ID, err)
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

	// Get all labs from database
	labs, err := s.repo.GetAllLabs()
	if err != nil {
		fmt.Printf("CleanupFailedLabs: Failed to get labs: %v\n", err)
		return
	}

	for _, lab := range labs {
		if lab.Status == models.LabStatusError {
			timeSinceError := now.Sub(lab.UpdatedAt)
			if timeSinceError > errorThreshold {
				// Create cleanup context for failed lab
				cleanupCtx := &interfaces.CleanupContext{
					LabID:   lab.ID,
					Context: context.Background(),
					Lab:     lab,
				}

				// Cleanup lab services
				if err := s.serviceManager.CleanupLabServices(cleanupCtx); err != nil {
					fmt.Printf("Warning: Failed to cleanup failed lab services for lab %s: %v\n", lab.ID, err)
				}

				// Cleanup progress tracking
				s.progressTracker.CleanupProgress(lab.ID)

				// Delete the lab from database
				if err := s.repo.DeleteLab(lab.ID); err != nil {
					fmt.Printf("Warning: Failed to delete failed lab %s: %v\n", lab.ID, err)
				}
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
