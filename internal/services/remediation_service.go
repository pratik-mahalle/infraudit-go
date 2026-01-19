package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/pratik-mahalle/infraudit/internal/domain/drift"
	"github.com/pratik-mahalle/infraudit/internal/domain/remediation"
	"github.com/pratik-mahalle/infraudit/internal/domain/vulnerability"
	"github.com/pratik-mahalle/infraudit/internal/pkg/logger"
)

// RemediationService implements remediation.Service
type RemediationService struct {
	repo         remediation.Repository
	driftService drift.Service
	vulnService  vulnerability.Service
	logger       *logger.Logger
}

// NewRemediationService creates a new remediation service
func NewRemediationService(
	repo remediation.Repository,
	driftService drift.Service,
	vulnService vulnerability.Service,
	log *logger.Logger,
) remediation.Service {
	return &RemediationService{
		repo:         repo,
		driftService: driftService,
		vulnService:  vulnService,
		logger:       log,
	}
}

// SuggestForDrift generates remediation suggestions for a drift
func (s *RemediationService) SuggestForDrift(ctx context.Context, driftID string) ([]*remediation.Suggestion, error) {
	// Parse drift ID as int64
	id, err := strconv.ParseInt(driftID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid drift ID: %w", err)
	}

	// Get drift (user ID 0 means admin/system lookup - in real implementation, get userID from context)
	d, err := s.driftService.GetByID(ctx, 0, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get drift: %w", err)
	}

	suggestions := s.generateDriftSuggestions(d)

	s.logger.WithFields(map[string]interface{}{
		"drift_id":    driftID,
		"suggestions": len(suggestions),
	}).Info("Generated remediation suggestions for drift")

	return suggestions, nil
}

// generateDriftSuggestions creates suggestions based on drift type
func (s *RemediationService) generateDriftSuggestions(d *drift.Drift) []*remediation.Suggestion {
	var suggestions []*remediation.Suggestion

	// Generate IaC PR suggestion for configuration drifts
	if d.DriftType == drift.TypeEncryption || d.DriftType == drift.TypeSecurityGroup || d.DriftType == drift.TypeConfigurationChange {
		suggestions = append(suggestions, &remediation.Suggestion{
			ID:              uuid.New().String(),
			IssueType:       "drift",
			IssueID:         fmt.Sprintf("%d", d.ID),
			Title:           fmt.Sprintf("Update IaC to match deployed configuration for resource %s", d.ResourceID),
			Description:     "Generate a pull request to update the Infrastructure as Code definition to match the current deployed state.",
			Severity:        d.Severity,
			RemediationType: remediation.RemediationTypeIaCPR,
			Strategy: &remediation.Strategy{
				Type:        remediation.RemediationTypeIaCPR,
				Description: "Create a pull request with the updated configuration",
				Steps: []remediation.RemediationStep{
					{Order: 1, Name: "Analyze drift", Description: "Compare IaC definition with deployed state"},
					{Order: 2, Name: "Generate patch", Description: "Create patch to align IaC with deployed state"},
					{Order: 3, Name: "Create branch", Description: "Create a new branch for the changes"},
					{Order: 4, Name: "Commit changes", Description: "Commit the patched file"},
					{Order: 5, Name: "Create PR", Description: "Create a pull request for review"},
				},
			},
			Risk:          "low",
			Impact:        "Updates IaC to match current state, requires review before applying",
			EstimatedTime: "5 minutes",
		})
	}

	// Generate Cloud API suggestion for immediate fix
	if d.Severity == drift.SeverityCritical || d.Severity == drift.SeverityHigh {
		suggestions = append(suggestions, &remediation.Suggestion{
			ID:              uuid.New().String(),
			IssueType:       "drift",
			IssueID:         fmt.Sprintf("%d", d.ID),
			Title:           fmt.Sprintf("Directly remediate resource %s via Cloud API", d.ResourceID),
			Description:     "Apply the fix directly through the cloud provider API to immediately resolve the security issue.",
			Severity:        d.Severity,
			RemediationType: remediation.RemediationTypeCloudAPI,
			Strategy: &remediation.Strategy{
				Type:        remediation.RemediationTypeCloudAPI,
				Description: "Apply fix via cloud provider API",
				Steps: []remediation.RemediationStep{
					{Order: 1, Name: "Validate access", Description: "Verify cloud credentials and permissions"},
					{Order: 2, Name: "Backup state", Description: "Record current state for rollback"},
					{Order: 3, Name: "Apply fix", Description: "Execute API call to apply remediation"},
					{Order: 4, Name: "Verify", Description: "Confirm the fix was applied successfully"},
				},
			},
			Risk:          "medium",
			Impact:        "Immediate fix applied to cloud resource, can be rolled back",
			EstimatedTime: "2 minutes",
		})
	}

	// Generate manual remediation suggestion
	suggestions = append(suggestions, &remediation.Suggestion{
		ID:              uuid.New().String(),
		IssueType:       "drift",
		IssueID:         fmt.Sprintf("%d", d.ID),
		Title:           "Manual remediation steps",
		Description:     "Follow the manual steps to remediate this issue through the cloud console or CLI.",
		Severity:        d.Severity,
		RemediationType: remediation.RemediationTypeManual,
		Strategy: &remediation.Strategy{
			Type:        remediation.RemediationTypeManual,
			Description: "Manual remediation guide",
			Steps:       s.generateManualSteps(d),
		},
		Risk:          "low",
		Impact:        "Manual intervention required, full control over changes",
		EstimatedTime: "10-30 minutes",
	})

	return suggestions
}

// generateManualSteps creates manual remediation steps based on drift type
func (s *RemediationService) generateManualSteps(d *drift.Drift) []remediation.RemediationStep {
	steps := []remediation.RemediationStep{
		{Order: 1, Name: "Identify resource", Description: fmt.Sprintf("Locate resource %s in the cloud console", d.ResourceID)},
	}

	switch d.DriftType {
	case drift.TypeEncryption:
		steps = append(steps,
			remediation.RemediationStep{Order: 2, Name: "Enable encryption", Description: "Navigate to the encryption settings and enable encryption at rest"},
			remediation.RemediationStep{Order: 3, Name: "Select key", Description: "Choose a KMS key for encryption"},
			remediation.RemediationStep{Order: 4, Name: "Apply changes", Description: "Save the changes and wait for encryption to complete"},
		)
	case drift.TypeSecurityGroup:
		steps = append(steps,
			remediation.RemediationStep{Order: 2, Name: "Review rules", Description: "Review the current inbound and outbound rules"},
			remediation.RemediationStep{Order: 3, Name: "Remove overly permissive rules", Description: "Remove any rules allowing 0.0.0.0/0 access"},
			remediation.RemediationStep{Order: 4, Name: "Add specific rules", Description: "Add rules with specific IP ranges or security groups"},
		)
	default:
		steps = append(steps,
			remediation.RemediationStep{Order: 2, Name: "Review configuration", Description: "Compare current configuration with expected state"},
			remediation.RemediationStep{Order: 3, Name: "Apply fix", Description: "Update the configuration to match the expected state"},
		)
	}

	steps = append(steps,
		remediation.RemediationStep{Order: len(steps) + 1, Name: "Verify fix", Description: "Run drift detection again to confirm the issue is resolved"},
	)

	return steps
}

// SuggestForVulnerability generates remediation suggestions for a vulnerability
func (s *RemediationService) SuggestForVulnerability(ctx context.Context, vulnerabilityID string) ([]*remediation.Suggestion, error) {
	// Parse vulnerability ID as int64
	id, err := strconv.ParseInt(vulnerabilityID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid vulnerability ID: %w", err)
	}

	v, err := s.vulnService.GetByID(ctx, 0, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get vulnerability: %w", err)
	}

	suggestions := s.generateVulnerabilitySuggestions(v)

	s.logger.WithFields(map[string]interface{}{
		"vulnerability_id": vulnerabilityID,
		"suggestions":      len(suggestions),
	}).Info("Generated remediation suggestions for vulnerability")

	return suggestions, nil
}

// generateVulnerabilitySuggestions creates suggestions for vulnerabilities
func (s *RemediationService) generateVulnerabilitySuggestions(v *vulnerability.Vulnerability) []*remediation.Suggestion {
	var suggestions []*remediation.Suggestion

	vulnIDStr := fmt.Sprintf("%d", v.ID)

	// IaC update suggestion
	suggestions = append(suggestions, &remediation.Suggestion{
		ID:              uuid.New().String(),
		IssueType:       "vulnerability",
		IssueID:         vulnIDStr,
		Title:           fmt.Sprintf("Fix %s by updating IaC configuration", v.Title),
		Description:     v.Description,
		Severity:        v.Severity,
		RemediationType: remediation.RemediationTypeIaCPR,
		Strategy: &remediation.Strategy{
			Type:        remediation.RemediationTypeIaCPR,
			Description: v.Remediation,
			Steps: []remediation.RemediationStep{
				{Order: 1, Name: "Identify affected code", Description: "Locate the IaC code defining this resource"},
				{Order: 2, Name: "Apply fix", Description: v.Remediation},
				{Order: 3, Name: "Create PR", Description: "Submit changes for review"},
			},
		},
		Risk:          "low",
		Impact:        "Fixes the security vulnerability in IaC",
		EstimatedTime: "10 minutes",
	})

	// Manual fix
	suggestions = append(suggestions, &remediation.Suggestion{
		ID:              uuid.New().String(),
		IssueType:       "vulnerability",
		IssueID:         vulnIDStr,
		Title:           "Manual remediation",
		Description:     "Follow the remediation steps manually",
		Severity:        v.Severity,
		RemediationType: remediation.RemediationTypeManual,
		Strategy: &remediation.Strategy{
			Type:        remediation.RemediationTypeManual,
			Description: v.Remediation,
			Steps: []remediation.RemediationStep{
				{Order: 1, Name: "Review vulnerability", Description: v.Description},
				{Order: 2, Name: "Apply remediation", Description: v.Remediation},
				{Order: 3, Name: "Rescan", Description: "Run vulnerability scan to verify fix"},
			},
		},
		Risk:          "low",
		Impact:        "Manual intervention required",
		EstimatedTime: "15-30 minutes",
	})

	return suggestions
}

// Create creates a remediation action from a suggestion
func (s *RemediationService) Create(ctx context.Context, userID int64, suggestion *remediation.Suggestion) (*remediation.Action, error) {
	action := &remediation.Action{
		ID:               uuid.New().String(),
		UserID:           userID,
		RemediationType:  suggestion.RemediationType,
		Status:           remediation.ActionStatusPending,
		Strategy:         suggestion.Strategy,
		ApprovalRequired: suggestion.RemediationType.RequiresApproval(),
	}

	// Set the related ID
	if suggestion.IssueType == "drift" {
		action.DriftID = &suggestion.IssueID
	} else if suggestion.IssueType == "vulnerability" {
		action.VulnerabilityID = &suggestion.IssueID
	}

	if err := s.repo.Create(ctx, action); err != nil {
		return nil, fmt.Errorf("failed to create remediation action: %w", err)
	}

	s.logger.WithFields(map[string]interface{}{
		"action_id":        action.ID,
		"user_id":          userID,
		"remediation_type": action.RemediationType,
	}).Info("Remediation action created")

	return action, nil
}

// Execute executes a remediation action
func (s *RemediationService) Execute(ctx context.Context, actionID string) error {
	action, err := s.repo.GetByID(ctx, actionID)
	if err != nil {
		return err
	}

	// Check if approval is required but not given
	if action.ApprovalRequired && action.Status != remediation.ActionStatusApproved {
		return fmt.Errorf("action requires approval before execution")
	}

	// Update status to in progress
	action.Status = remediation.ActionStatusInProgress
	now := time.Now()
	action.StartedAt = &now

	if err := s.repo.Update(ctx, action); err != nil {
		return err
	}

	// Execute based on type
	go func() {
		execCtx := context.Background()
		var result *remediation.Result
		var execErr error

		switch action.RemediationType {
		case remediation.RemediationTypeIaCPR:
			result, execErr = s.executeIaCPR(execCtx, action)
		case remediation.RemediationTypeCloudAPI:
			result, execErr = s.executeCloudAPI(execCtx, action)
		case remediation.RemediationTypePolicy:
			result, execErr = s.executePolicy(execCtx, action)
		case remediation.RemediationTypeManual:
			// Manual actions just need to be marked as completed by user
			result = &remediation.Result{
				Success: true,
				Message: "Manual remediation steps provided",
			}
		}

		completedAt := time.Now()
		action.CompletedAt = &completedAt

		if execErr != nil {
			action.Status = remediation.ActionStatusFailed
			action.ErrorMessage = execErr.Error()
		} else {
			action.Status = remediation.ActionStatusCompleted
			if result != nil {
				resultJSON, _ := json.Marshal(result)
				action.Result = resultJSON
			}
		}

		s.repo.Update(execCtx, action)

		s.logger.WithFields(map[string]interface{}{
			"action_id": actionID,
			"status":    action.Status,
		}).Info("Remediation action execution completed")
	}()

	return nil
}

// executeIaCPR executes IaC PR remediation
func (s *RemediationService) executeIaCPR(ctx context.Context, action *remediation.Action) (*remediation.Result, error) {
	// TODO: Integrate with GitHub/GitLab API
	// For now, return a placeholder result
	return &remediation.Result{
		Success:          true,
		Message:          "IaC PR generation is configured but requires GitHub/GitLab integration",
		RollbackPossible: false,
		Details: map[string]interface{}{
			"note": "Configure GITHUB_TOKEN or GITLAB_TOKEN to enable PR creation",
		},
	}, nil
}

// executeCloudAPI executes Cloud API remediation
func (s *RemediationService) executeCloudAPI(ctx context.Context, action *remediation.Action) (*remediation.Result, error) {
	// TODO: Integrate with cloud provider SDK
	// For now, return a placeholder result
	return &remediation.Result{
		Success:          true,
		Message:          "Cloud API remediation requires cloud provider credentials",
		RollbackPossible: true,
		Details: map[string]interface{}{
			"note": "Configure cloud provider credentials to enable API remediation",
		},
	}, nil
}

// executePolicy executes policy remediation
func (s *RemediationService) executePolicy(ctx context.Context, action *remediation.Action) (*remediation.Result, error) {
	// TODO: Integrate with policy management
	return &remediation.Result{
		Success: true,
		Message: "Policy enforcement configured",
	}, nil
}

// Approve approves a remediation action
func (s *RemediationService) Approve(ctx context.Context, actionID string, approverID int64) error {
	action, err := s.repo.GetByID(ctx, actionID)
	if err != nil {
		return err
	}

	if action.Status != remediation.ActionStatusPending {
		return fmt.Errorf("action is not pending approval")
	}

	action.Status = remediation.ActionStatusApproved
	action.ApprovedBy = &approverID
	now := time.Now()
	action.ApprovedAt = &now

	if err := s.repo.Update(ctx, action); err != nil {
		return err
	}

	s.logger.WithFields(map[string]interface{}{
		"action_id":   actionID,
		"approved_by": approverID,
	}).Info("Remediation action approved")

	return nil
}

// Reject rejects a remediation action
func (s *RemediationService) Reject(ctx context.Context, actionID string, reason string) error {
	action, err := s.repo.GetByID(ctx, actionID)
	if err != nil {
		return err
	}

	if action.Status != remediation.ActionStatusPending {
		return fmt.Errorf("action is not pending")
	}

	action.Status = remediation.ActionStatusRejected
	action.ErrorMessage = reason
	now := time.Now()
	action.CompletedAt = &now

	if err := s.repo.Update(ctx, action); err != nil {
		return err
	}

	s.logger.WithFields(map[string]interface{}{
		"action_id": actionID,
		"reason":    reason,
	}).Info("Remediation action rejected")

	return nil
}

// Rollback rolls back a completed remediation action
func (s *RemediationService) Rollback(ctx context.Context, actionID string) error {
	action, err := s.repo.GetByID(ctx, actionID)
	if err != nil {
		return err
	}

	if action.Status != remediation.ActionStatusCompleted {
		return fmt.Errorf("can only rollback completed actions")
	}

	if len(action.RollbackData) == 0 {
		return fmt.Errorf("no rollback data available")
	}

	// TODO: Execute rollback based on rollback data

	action.Status = remediation.ActionStatusRolledBack
	if err := s.repo.Update(ctx, action); err != nil {
		return err
	}

	s.logger.WithFields(map[string]interface{}{
		"action_id": actionID,
	}).Info("Remediation action rolled back")

	return nil
}

// GetAction retrieves a remediation action
func (s *RemediationService) GetAction(ctx context.Context, id string) (*remediation.Action, error) {
	return s.repo.GetByID(ctx, id)
}

// ListActions lists remediation actions
func (s *RemediationService) ListActions(ctx context.Context, filter remediation.Filter, limit, offset int) ([]*remediation.Action, int64, error) {
	return s.repo.List(ctx, filter, limit, offset)
}

// GetPendingApprovals retrieves actions pending approval
func (s *RemediationService) GetPendingApprovals(ctx context.Context, userID int64) ([]*remediation.Action, error) {
	return s.repo.GetPendingApprovals(ctx, userID)
}

// GetSummary returns a summary of remediation actions by status
func (s *RemediationService) GetSummary(ctx context.Context, userID int64) (map[remediation.ActionStatus]int, error) {
	return s.repo.CountByStatus(ctx, userID)
}
