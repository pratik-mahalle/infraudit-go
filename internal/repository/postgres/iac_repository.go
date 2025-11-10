package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/pratik-mahalle/infraudit/internal/domain/iac"
	"github.com/pratik-mahalle/infraudit/internal/pkg/errors"
)

// IaCRepository handles database operations for IaC definitions
type IaCRepository struct {
	db *sql.DB
}

// NewIaCRepository creates a new IaC repository
func NewIaCRepository(db *sql.DB) *IaCRepository {
	return &IaCRepository{db: db}
}

// CreateDefinition creates a new IaC definition
func (r *IaCRepository) CreateDefinition(ctx context.Context, def *iac.IaCDefinition) error {
	if def.ID == "" {
		def.ID = uuid.New().String()
	}

	now := time.Now()
	def.CreatedAt = now
	def.UpdatedAt = now

	// Serialize parsed resources to JSON
	var parsedResourcesJSON sql.NullString
	if def.ParsedResources != nil {
		data, err := json.Marshal(def.ParsedResources)
		if err != nil {
			return errors.DatabaseError("Failed to marshal parsed resources", err)
		}
		parsedResourcesJSON = sql.NullString{String: string(data), Valid: true}
	}

	query := `
		INSERT INTO iac_definitions
		(id, user_id, name, iac_type, file_path, content, parsed_resources, last_parsed, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		def.ID,
		def.UserID,
		def.Name,
		def.IaCType,
		def.FilePath,
		def.Content,
		parsedResourcesJSON,
		def.LastParsed,
		def.CreatedAt,
		def.UpdatedAt,
	)

	if err != nil {
		return errors.DatabaseError("Failed to create IaC definition", err)
	}

	return nil
}

// GetDefinitionByID retrieves an IaC definition by ID
func (r *IaCRepository) GetDefinitionByID(ctx context.Context, userID, definitionID string) (*iac.IaCDefinition, error) {
	query := `
		SELECT id, user_id, name, iac_type, file_path, content, parsed_resources, last_parsed, created_at, updated_at
		FROM iac_definitions
		WHERE id = ? AND user_id = ?
	`

	var def iac.IaCDefinition
	var parsedResourcesJSON sql.NullString
	var lastParsed sql.NullTime

	err := r.db.QueryRowContext(ctx, query, definitionID, userID).Scan(
		&def.ID,
		&def.UserID,
		&def.Name,
		&def.IaCType,
		&def.FilePath,
		&def.Content,
		&parsedResourcesJSON,
		&lastParsed,
		&def.CreatedAt,
		&def.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, iac.ErrDefinitionNotFound
	}
	if err != nil {
		return nil, errors.DatabaseError("Failed to get IaC definition", err)
	}

	// Deserialize parsed resources
	if parsedResourcesJSON.Valid {
		if err := json.Unmarshal([]byte(parsedResourcesJSON.String), &def.ParsedResources); err != nil {
			return nil, errors.DatabaseError("Failed to unmarshal parsed resources", err)
		}
	}

	if lastParsed.Valid {
		def.LastParsed = &lastParsed.Time
	}

	return &def, nil
}

// ListDefinitions lists all IaC definitions for a user
func (r *IaCRepository) ListDefinitions(ctx context.Context, userID string, iacType *iac.IaCType) ([]*iac.IaCDefinition, error) {
	query := `
		SELECT id, user_id, name, iac_type, file_path, content, parsed_resources, last_parsed, created_at, updated_at
		FROM iac_definitions
		WHERE user_id = ?
	`

	args := []interface{}{userID}

	if iacType != nil {
		query += " AND iac_type = ?"
		args = append(args, *iacType)
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.DatabaseError("Failed to list IaC definitions", err)
	}
	defer rows.Close()

	definitions := make([]*iac.IaCDefinition, 0)

	for rows.Next() {
		var def iac.IaCDefinition
		var parsedResourcesJSON sql.NullString
		var lastParsed sql.NullTime

		err := rows.Scan(
			&def.ID,
			&def.UserID,
			&def.Name,
			&def.IaCType,
			&def.FilePath,
			&def.Content,
			&parsedResourcesJSON,
			&lastParsed,
			&def.CreatedAt,
			&def.UpdatedAt,
		)
		if err != nil {
			return nil, errors.DatabaseError("Failed to scan IaC definition", err)
		}

		// Deserialize parsed resources
		if parsedResourcesJSON.Valid {
			if err := json.Unmarshal([]byte(parsedResourcesJSON.String), &def.ParsedResources); err != nil {
				return nil, errors.DatabaseError("Failed to unmarshal parsed resources", err)
			}
		}

		if lastParsed.Valid {
			def.LastParsed = &lastParsed.Time
		}

		definitions = append(definitions, &def)
	}

	return definitions, nil
}

// UpdateDefinition updates an IaC definition
func (r *IaCRepository) UpdateDefinition(ctx context.Context, def *iac.IaCDefinition) error {
	def.UpdatedAt = time.Now()

	// Serialize parsed resources
	var parsedResourcesJSON sql.NullString
	if def.ParsedResources != nil {
		data, err := json.Marshal(def.ParsedResources)
		if err != nil {
			return errors.DatabaseError("Failed to marshal parsed resources", err)
		}
		parsedResourcesJSON = sql.NullString{String: string(data), Valid: true}
	}

	query := `
		UPDATE iac_definitions
		SET name = ?, content = ?, parsed_resources = ?, last_parsed = ?, updated_at = ?
		WHERE id = ? AND user_id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		def.Name,
		def.Content,
		parsedResourcesJSON,
		def.LastParsed,
		def.UpdatedAt,
		def.ID,
		def.UserID,
	)

	if err != nil {
		return errors.DatabaseError("Failed to update IaC definition", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.DatabaseError("Failed to get rows affected", err)
	}

	if rowsAffected == 0 {
		return iac.ErrDefinitionNotFound
	}

	return nil
}

// DeleteDefinition deletes an IaC definition
func (r *IaCRepository) DeleteDefinition(ctx context.Context, userID, definitionID string) error {
	query := `DELETE FROM iac_definitions WHERE id = ? AND user_id = ?`

	result, err := r.db.ExecContext(ctx, query, definitionID, userID)
	if err != nil {
		return errors.DatabaseError("Failed to delete IaC definition", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.DatabaseError("Failed to get rows affected", err)
	}

	if rowsAffected == 0 {
		return iac.ErrDefinitionNotFound
	}

	return nil
}

// CreateResource creates a new IaC resource
func (r *IaCRepository) CreateResource(ctx context.Context, res *iac.IaCResource) error {
	if res.ID == "" {
		res.ID = uuid.New().String()
	}

	now := time.Now()
	res.CreatedAt = now

	// Serialize configuration
	configJSON, err := json.Marshal(res.Configuration)
	if err != nil {
		return errors.DatabaseError("Failed to marshal configuration", err)
	}

	query := `
		INSERT INTO iac_resources
		(id, iac_definition_id, user_id, resource_type, resource_name, resource_address, provider, configuration, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = r.db.ExecContext(ctx, query,
		res.ID,
		res.IaCDefinitionID,
		res.UserID,
		res.ResourceType,
		res.ResourceName,
		res.ResourceAddress,
		res.Provider,
		string(configJSON),
		res.CreatedAt,
	)

	if err != nil {
		return errors.DatabaseError("Failed to create IaC resource", err)
	}

	return nil
}

// ListResourcesByDefinition lists all resources for an IaC definition
func (r *IaCRepository) ListResourcesByDefinition(ctx context.Context, userID, definitionID string) ([]*iac.IaCResource, error) {
	query := `
		SELECT id, iac_definition_id, user_id, resource_type, resource_name, resource_address, provider, configuration, created_at
		FROM iac_resources
		WHERE iac_definition_id = ? AND user_id = ?
		ORDER BY resource_type, resource_name
	`

	rows, err := r.db.QueryContext(ctx, query, definitionID, userID)
	if err != nil {
		return nil, errors.DatabaseError("Failed to list IaC resources", err)
	}
	defer rows.Close()

	resources := make([]*iac.IaCResource, 0)

	for rows.Next() {
		var res iac.IaCResource
		var configJSON string

		err := rows.Scan(
			&res.ID,
			&res.IaCDefinitionID,
			&res.UserID,
			&res.ResourceType,
			&res.ResourceName,
			&res.ResourceAddress,
			&res.Provider,
			&configJSON,
			&res.CreatedAt,
		)
		if err != nil {
			return nil, errors.DatabaseError("Failed to scan IaC resource", err)
		}

		// Deserialize configuration
		if err := json.Unmarshal([]byte(configJSON), &res.Configuration); err != nil {
			return nil, errors.DatabaseError("Failed to unmarshal configuration", err)
		}

		resources = append(resources, &res)
	}

	return resources, nil
}

// DeleteResourcesByDefinition deletes all resources for an IaC definition
func (r *IaCRepository) DeleteResourcesByDefinition(ctx context.Context, userID, definitionID string) error {
	query := `DELETE FROM iac_resources WHERE iac_definition_id = ? AND user_id = ?`

	_, err := r.db.ExecContext(ctx, query, definitionID, userID)
	if err != nil {
		return errors.DatabaseError("Failed to delete IaC resources", err)
	}

	return nil
}

// CreateDriftResult creates a new IaC drift result
func (r *IaCRepository) CreateDriftResult(ctx context.Context, drift *iac.IaCDriftResult) error {
	if drift.ID == "" {
		drift.ID = uuid.New().String()
	}

	now := time.Now()
	drift.DetectedAt = now

	// Serialize details
	var detailsJSON sql.NullString
	if drift.Details != nil {
		data, err := json.Marshal(drift.Details)
		if err != nil {
			return errors.DatabaseError("Failed to marshal details", err)
		}
		detailsJSON = sql.NullString{String: string(data), Valid: true}
	}

	// Convert pointers to nullable types
	var severityStr sql.NullString
	if drift.Severity != nil {
		severityStr = sql.NullString{String: string(*drift.Severity), Valid: true}
	}

	var iacResourceID sql.NullString
	if drift.IaCResourceID != nil {
		iacResourceID = sql.NullString{String: *drift.IaCResourceID, Valid: true}
	}

	var actualResourceID sql.NullString
	if drift.ActualResourceID != nil {
		actualResourceID = sql.NullString{String: *drift.ActualResourceID, Valid: true}
	}

	query := `
		INSERT INTO iac_drift_results
		(id, user_id, iac_definition_id, iac_resource_id, actual_resource_id, drift_category, severity, details, detected_at, status, resolved_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		drift.ID,
		drift.UserID,
		drift.IaCDefinitionID,
		iacResourceID,
		actualResourceID,
		drift.DriftCategory,
		severityStr,
		detailsJSON,
		drift.DetectedAt,
		drift.Status,
		drift.ResolvedAt,
	)

	if err != nil {
		return errors.DatabaseError("Failed to create IaC drift result", err)
	}

	return nil
}

// ListDriftResults lists drift results with optional filtering
func (r *IaCRepository) ListDriftResults(ctx context.Context, userID string, definitionID *string, category *iac.DriftCategory, status *iac.DriftStatus) ([]*iac.IaCDriftResult, error) {
	query := `
		SELECT id, user_id, iac_definition_id, iac_resource_id, actual_resource_id, drift_category, severity, details, detected_at, status, resolved_at
		FROM iac_drift_results
		WHERE user_id = ?
	`

	args := []interface{}{userID}

	if definitionID != nil {
		query += " AND iac_definition_id = ?"
		args = append(args, *definitionID)
	}

	if category != nil {
		query += " AND drift_category = ?"
		args = append(args, *category)
	}

	if status != nil {
		query += " AND status = ?"
		args = append(args, *status)
	}

	query += " ORDER BY detected_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.DatabaseError("Failed to list drift results", err)
	}
	defer rows.Close()

	drifts := make([]*iac.IaCDriftResult, 0)

	for rows.Next() {
		var drift iac.IaCDriftResult
		var detailsJSON sql.NullString
		var severityStr sql.NullString
		var iacResourceID sql.NullString
		var actualResourceID sql.NullString
		var resolvedAt sql.NullTime

		err := rows.Scan(
			&drift.ID,
			&drift.UserID,
			&drift.IaCDefinitionID,
			&iacResourceID,
			&actualResourceID,
			&drift.DriftCategory,
			&severityStr,
			&detailsJSON,
			&drift.DetectedAt,
			&drift.Status,
			&resolvedAt,
		)
		if err != nil {
			return nil, errors.DatabaseError("Failed to scan drift result", err)
		}

		// Deserialize details
		if detailsJSON.Valid {
			if err := json.Unmarshal([]byte(detailsJSON.String), &drift.Details); err != nil {
				return nil, errors.DatabaseError("Failed to unmarshal details", err)
			}
		}

		// Convert nullable types to pointers
		if severityStr.Valid {
			severity := iac.Severity(severityStr.String)
			drift.Severity = &severity
		}

		if iacResourceID.Valid {
			drift.IaCResourceID = &iacResourceID.String
		}

		if actualResourceID.Valid {
			drift.ActualResourceID = &actualResourceID.String
		}

		if resolvedAt.Valid {
			drift.ResolvedAt = &resolvedAt.Time
		}

		drifts = append(drifts, &drift)
	}

	return drifts, nil
}

// UpdateDriftStatus updates the status of a drift result
func (r *IaCRepository) UpdateDriftStatus(ctx context.Context, userID, driftID string, status iac.DriftStatus) error {
	query := `
		UPDATE iac_drift_results
		SET status = ?, resolved_at = ?
		WHERE id = ? AND user_id = ?
	`

	var resolvedAt sql.NullTime
	if status == iac.DriftStatusResolved {
		resolvedAt = sql.NullTime{Time: time.Now(), Valid: true}
	}

	result, err := r.db.ExecContext(ctx, query, status, resolvedAt, driftID, userID)
	if err != nil {
		return errors.DatabaseError("Failed to update drift status", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.DatabaseError("Failed to get rows affected", err)
	}

	if rowsAffected == 0 {
		return iac.ErrDriftNotFound
	}

	return nil
}

// GetDriftSummary returns a summary of drift results by category and severity
func (r *IaCRepository) GetDriftSummary(ctx context.Context, userID string, definitionID *string) (map[string]interface{}, error) {
	query := `
		SELECT
			drift_category,
			severity,
			COUNT(*) as count
		FROM iac_drift_results
		WHERE user_id = ? AND status = ?
	`

	args := []interface{}{userID, iac.DriftStatusDetected}

	if definitionID != nil {
		query += " AND iac_definition_id = ?"
		args = append(args, *definitionID)
	}

	query += " GROUP BY drift_category, severity"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.DatabaseError("Failed to get drift summary", err)
	}
	defer rows.Close()

	summary := map[string]interface{}{
		"total":    0,
		"by_category": make(map[string]int),
		"by_severity": make(map[string]int),
	}

	total := 0

	for rows.Next() {
		var category string
		var severity sql.NullString
		var count int

		if err := rows.Scan(&category, &severity, &count); err != nil {
			return nil, errors.DatabaseError("Failed to scan drift summary", err)
		}

		total += count

		// Add to category count
		if byCat, ok := summary["by_category"].(map[string]int); ok {
			byCat[category] = byCat[category] + count
		}

		// Add to severity count
		if severity.Valid {
			if bySev, ok := summary["by_severity"].(map[string]int); ok {
				bySev[severity.String] = bySev[severity.String] + count
			}
		}
	}

	summary["total"] = total

	return summary, nil
}
