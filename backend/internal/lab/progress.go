package lab

import (
	"sync"
	"time"
)

// Constants for progress tracking
const (
	MinLabDurationMinutes = 15
	MaxLabDurationMinutes = 480
	MaxLogEntries         = 50
	ProgressComplete      = 100
)

// ProgressStep represents a step within a service
type ProgressStep struct {
	Name        string    `json:"name"`
	Status      string    `json:"status"` // "pending", "running", "completed", "failed"
	Message     string    `json:"message"`
	StartedAt   time.Time `json:"started_at,omitempty"`
	CompletedAt time.Time `json:"completed_at,omitempty"`
	Error       string    `json:"error,omitempty"`
}

// ServiceProgress represents the progress of a single service
type ServiceProgress struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Status      string         `json:"status"`   // "pending", "running", "completed", "failed"
	Progress    int            `json:"progress"` // 0-100
	Steps       []ProgressStep `json:"steps"`
	StartedAt   time.Time      `json:"started_at,omitempty"`
	CompletedAt time.Time      `json:"completed_at,omitempty"`
	Error       string         `json:"error,omitempty"`
}

// LabProgress tracks the progress of lab creation
type LabProgress struct {
	LabID       string            `json:"lab_id"`
	Overall     int               `json:"overall"` // 0-100
	CurrentStep string            `json:"current_step"`
	Services    []ServiceProgress `json:"services"`
	Logs        []string          `json:"logs"`
	UpdatedAt   time.Time         `json:"updated_at"`
	mu          sync.RWMutex
}

// ProgressTracker manages progress for all labs
type ProgressTracker struct {
	progress map[string]*LabProgress
	mu       sync.RWMutex
}

// NewProgressTracker creates a new progress tracker
func NewProgressTracker() *ProgressTracker {
	return &ProgressTracker{
		progress: make(map[string]*LabProgress),
	}
}

// InitializeProgress creates a new progress tracker for a lab
func (pt *ProgressTracker) InitializeProgress(labID string) *LabProgress {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	progress := &LabProgress{
		LabID:       labID,
		Overall:     0,
		CurrentStep: "Initializing",
		Services:    []ServiceProgress{},
		Logs:        []string{},
		UpdatedAt:   time.Now(),
	}

	pt.progress[labID] = progress
	return progress
}

// AddService adds a service to the progress tracker
func (pt *ProgressTracker) AddService(labID, serviceName, description string, steps []string) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	progress, exists := pt.progress[labID]
	if !exists {
		return
	}

	progress.mu.Lock()
	defer progress.mu.Unlock()

	// Create steps for the service
	progressSteps := make([]ProgressStep, len(steps))
	for i, stepName := range steps {
		progressSteps[i] = ProgressStep{
			Name:   stepName,
			Status: "pending",
		}
	}

	service := ServiceProgress{
		Name:        serviceName,
		Description: description,
		Status:      "pending",
		Progress:    0,
		Steps:       progressSteps,
	}

	progress.Services = append(progress.Services, service)
	progress.UpdatedAt = time.Now()
}

// GetProgress returns the progress for a lab
func (pt *ProgressTracker) GetProgress(labID string) *LabProgress {
	pt.mu.RLock()
	defer pt.mu.RUnlock()
	return pt.progress[labID]
}

// UpdateServiceStep updates a specific step within a service
func (pt *ProgressTracker) UpdateServiceStep(labID, serviceName, stepName, status, message string) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	progress, exists := pt.progress[labID]
	if !exists {
		return
	}

	progress.mu.Lock()
	defer progress.mu.Unlock()

	// Find the service and update the step
	for i, service := range progress.Services {
		if service.Name == serviceName {
			// Find and update the step
			for j, step := range service.Steps {
				if step.Name == stepName {
					progress.Services[i].Steps[j].Status = status
					progress.Services[i].Steps[j].Message = message

					if status == "running" && step.StartedAt.IsZero() {
						progress.Services[i].Steps[j].StartedAt = time.Now()
					} else if status == "completed" || status == "failed" {
						progress.Services[i].Steps[j].CompletedAt = time.Now()
					}

					break
				}
			}

			// Update service status
			if status == "running" && service.Status == "pending" {
				progress.Services[i].Status = "running"
				progress.Services[i].StartedAt = time.Now()
			} else if status == "completed" {
				progress.Services[i].Status = "completed"
				progress.Services[i].CompletedAt = time.Now()
			} else if status == "failed" {
				progress.Services[i].Status = "failed"
				progress.Services[i].Error = message
			}

			// Calculate service progress
			completed := 0
			for _, step := range service.Steps {
				if step.Status == "completed" {
					completed++
				}
			}
			progress.Services[i].Progress = (completed * 100) / len(service.Steps)
			break
		}
	}

	// Update current step
	progress.CurrentStep = stepName
	progress.UpdatedAt = time.Now()

	// Calculate overall progress
	totalCompleted := 0
	totalSteps := 0
	for _, service := range progress.Services {
		for _, step := range service.Steps {
			totalSteps++
			if step.Status == "completed" {
				totalCompleted++
			}
		}
	}
	if totalSteps > 0 {
		progress.Overall = (totalCompleted * 100) / totalSteps
	}
}

// AddLog adds a log message to the progress
func (pt *ProgressTracker) AddLog(labID, message string) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	progress, exists := pt.progress[labID]
	if !exists {
		return
	}

	progress.mu.Lock()
	defer progress.mu.Unlock()

	timestamp := time.Now().Format("15:04:05")
	logEntry := "[" + timestamp + "] " + message
	progress.Logs = append(progress.Logs, logEntry)

	// Keep only last 50 logs
	if len(progress.Logs) > MaxLogEntries {
		progress.Logs = progress.Logs[len(progress.Logs)-MaxLogEntries:]
	}

	progress.UpdatedAt = time.Now()
}

// CompleteProgress marks the progress as complete
func (pt *ProgressTracker) CompleteProgress(labID string) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	progress, exists := pt.progress[labID]
	if !exists {
		return
	}

	progress.mu.Lock()
	defer progress.mu.Unlock()

	// Mark all pending services as completed
	for i, service := range progress.Services {
		if service.Status == "pending" || service.Status == "running" {
			progress.Services[i].Status = "completed"
			progress.Services[i].CompletedAt = time.Now()

			// Mark all pending steps as completed
			for j, step := range service.Steps {
				if step.Status == "pending" || step.Status == "running" {
					progress.Services[i].Steps[j].Status = "completed"
					progress.Services[i].Steps[j].Message = "Lab setup completed successfully"
					progress.Services[i].Steps[j].CompletedAt = time.Now()
				}
			}

			// Update service progress to 100%
			progress.Services[i].Progress = ProgressComplete
		}
	}

	// Update overall progress to 100%
	progress.Overall = ProgressComplete
	progress.CurrentStep = "Lab setup completed successfully"
	progress.UpdatedAt = time.Now()
}

// FailProgress marks the progress as failed
func (pt *ProgressTracker) FailProgress(labID, error string) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	progress, exists := pt.progress[labID]
	if !exists {
		return
	}

	progress.mu.Lock()
	defer progress.mu.Unlock()

	// Mark all pending services as failed
	for i, service := range progress.Services {
		if service.Status == "pending" || service.Status == "running" {
			progress.Services[i].Status = "failed"
			progress.Services[i].Error = error

			// Mark all pending steps as failed
			for j, step := range service.Steps {
				if step.Status == "pending" || step.Status == "running" {
					progress.Services[i].Steps[j].Status = "failed"
					progress.Services[i].Steps[j].Message = "Lab setup failed: " + error
					progress.Services[i].Steps[j].CompletedAt = time.Now()
				}
			}
		}
	}

	progress.CurrentStep = "Lab setup failed: " + error
	progress.UpdatedAt = time.Now()
}

// CleanupProgress removes progress data for a lab
func (pt *ProgressTracker) CleanupProgress(labID string) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	delete(pt.progress, labID)
}
