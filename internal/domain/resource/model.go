package resource

import "time"

// Resource represents a cloud resource
type Resource struct {
	ID            string    `json:"id"`
	UserID        int64     `json:"user_id"`
	Provider      string    `json:"provider"`
	ResourceID    string    `json:"resource_id"`
	Name          string    `json:"name"`
	Type          string    `json:"type"`
	Region        string    `json:"region"`
	Status        string    `json:"status"`
	Configuration string    `json:"configuration,omitempty"` // JSON string of full resource config
	LastScanned   time.Time `json:"last_scanned,omitempty"`
	CreatedAt     time.Time `json:"created_at,omitempty"`
	UpdatedAt     time.Time `json:"updated_at,omitempty"`
}

// Resource types
const (
	TypeEC2Instance      = "ec2-instance"
	TypeS3Bucket         = "s3-bucket"
	TypeRDSInstance      = "rds-instance"
	TypeLambdaFunction   = "lambda-function"
	TypeGCEInstance      = "gce-instance"
	TypeGCSBucket        = "gcs-bucket"
	TypeAzureVM          = "azure-vm"
	TypeAzureStorage     = "azure-storage"
)

// Resource status
const (
	StatusRunning   = "running"
	StatusStopped   = "stopped"
	StatusTerminated = "terminated"
	StatusActive    = "active"
	StatusInactive  = "inactive"
	StatusUnknown   = "unknown"
)

// Filter contains resource filtering options
type Filter struct {
	Provider string
	Type     string
	Region   string
	Status   string
}
