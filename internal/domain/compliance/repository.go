package compliance

import (
	"context"
)

// Repository defines the compliance repository interface
type Repository interface {
	// Frameworks
	CreateFramework(ctx context.Context, framework *Framework) error
	GetFramework(ctx context.Context, id string) (*Framework, error)
	GetFrameworkByName(ctx context.Context, name string) (*Framework, error)
	ListFrameworks(ctx context.Context) ([]*Framework, error)
	UpdateFramework(ctx context.Context, framework *Framework) error

	// Controls
	CreateControl(ctx context.Context, control *Control) error
	GetControl(ctx context.Context, id string) (*Control, error)
	GetControlByFrameworkAndID(ctx context.Context, frameworkID, controlID string) (*Control, error)
	ListControls(ctx context.Context, frameworkID string, category string) ([]*Control, error)
	CountControls(ctx context.Context, frameworkID string) (int, error)

	// Control Mappings
	CreateMapping(ctx context.Context, mapping *ControlMapping) error
	GetMappingsForControl(ctx context.Context, controlID string) ([]*ControlMapping, error)
	GetMappingsForSecurityRule(ctx context.Context, ruleType string, resourceType string) ([]*ControlMapping, error)

	// Assessments
	CreateAssessment(ctx context.Context, assessment *Assessment) error
	GetAssessment(ctx context.Context, id string) (*Assessment, error)
	UpdateAssessment(ctx context.Context, assessment *Assessment) error
	ListAssessments(ctx context.Context, userID int64, frameworkID string, limit, offset int) ([]*Assessment, int64, error)
	GetLatestAssessment(ctx context.Context, userID int64, frameworkID string) (*Assessment, error)
}
