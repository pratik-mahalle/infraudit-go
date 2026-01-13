package iac

import "errors"

var (
	// IaCDefinition errors
	ErrMissingUserID     = errors.New("user ID is required")
	ErrMissingName       = errors.New("name is required")
	ErrMissingIaCType    = errors.New("IaC type is required")
	ErrMissingContent    = errors.New("content is required")
	ErrInvalidIaCType    = errors.New("invalid IaC type")
	ErrDefinitionNotFound = errors.New("IaC definition not found")

	// IaCResource errors
	ErrMissingDefinitionID  = errors.New("IaC definition ID is required")
	ErrMissingResourceType  = errors.New("resource type is required")
	ErrMissingResourceName  = errors.New("resource name is required")
	ErrMissingProvider      = errors.New("provider is required")
	ErrResourceNotFound     = errors.New("IaC resource not found")

	// IaCDriftResult errors
	ErrMissingDriftCategory  = errors.New("drift category is required")
	ErrInvalidDriftCategory  = errors.New("invalid drift category")
	ErrDriftNotFound         = errors.New("drift result not found")

	// Parsing errors
	ErrParsingFailed         = errors.New("failed to parse IaC file")
	ErrUnsupportedFormat     = errors.New("unsupported IaC format")
	ErrInvalidSyntax         = errors.New("invalid IaC syntax")

	// Comparison errors
	ErrComparisonFailed      = errors.New("failed to compare IaC with actual resources")
	ErrNoResourcesFound      = errors.New("no resources found for comparison")
)
