package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"infraaudit/backend/internal/domain/recommendation"
	"infraaudit/backend/internal/pkg/errors"
)

type RecommendationRepository struct {
	db *sql.DB
}

func NewRecommendationRepository(db *sql.DB) recommendation.Repository {
	return &RecommendationRepository{db: db}
}

func (r *RecommendationRepository) Create(ctx context.Context, rec *recommendation.Recommendation) (int64, error) {
	now := time.Now()
	rec.CreatedAt = now
	rec.UpdatedAt = now

	resourcesJSON, _ := json.Marshal(rec.Resources)

	query := `
		INSERT INTO recommendations (user_id, type, priority, title, description, savings, effort, impact, category, resources)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query,
		rec.UserID, rec.Type, rec.Priority, rec.Title, rec.Description, rec.Savings, rec.Effort, rec.Impact, rec.Category, string(resourcesJSON),
	)
	if err != nil {
		return 0, errors.DatabaseError("Failed to create recommendation", err)
	}

	return result.LastInsertId()
}

func (r *RecommendationRepository) GetByID(ctx context.Context, userID int64, id int64) (*recommendation.Recommendation, error) {
	query := `
		SELECT id, user_id, type, priority, title, description, savings, effort, impact, category, resources
		FROM recommendations WHERE user_id = ? AND id = ?
	`

	var rec recommendation.Recommendation
	var resourcesJSON string
	err := r.db.QueryRowContext(ctx, query, userID, id).Scan(
		&rec.ID, &rec.UserID, &rec.Type, &rec.Priority, &rec.Title, &rec.Description, &rec.Savings, &rec.Effort, &rec.Impact, &rec.Category, &resourcesJSON,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NotFound("Recommendation")
	}
	if err != nil {
		return nil, errors.DatabaseError("Failed to get recommendation", err)
	}

	json.Unmarshal([]byte(resourcesJSON), &rec.Resources)
	return &rec, nil
}

func (r *RecommendationRepository) Update(ctx context.Context, rec *recommendation.Recommendation) error {
	rec.UpdatedAt = time.Now()
	resourcesJSON, _ := json.Marshal(rec.Resources)

	query := `
		UPDATE recommendations SET type = ?, priority = ?, title = ?, description = ?, savings = ?, effort = ?, impact = ?, category = ?, resources = ?
		WHERE user_id = ? AND id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		rec.Type, rec.Priority, rec.Title, rec.Description, rec.Savings, rec.Effort, rec.Impact, rec.Category, string(resourcesJSON), rec.UserID, rec.ID,
	)
	if err != nil {
		return errors.DatabaseError("Failed to update recommendation", err)
	}

	rows, err := result.RowsAffected()
	if err != nil || rows == 0 {
		return errors.NotFound("Recommendation")
	}

	return nil
}

func (r *RecommendationRepository) Delete(ctx context.Context, userID int64, id int64) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM recommendations WHERE user_id = ? AND id = ?", userID, id)
	if err != nil {
		return errors.DatabaseError("Failed to delete recommendation", err)
	}

	rows, err := result.RowsAffected()
	if err != nil || rows == 0 {
		return errors.NotFound("Recommendation")
	}

	return nil
}

func (r *RecommendationRepository) List(ctx context.Context, userID int64, filter recommendation.Filter) ([]*recommendation.Recommendation, error) {
	where := []string{"user_id = ?"}
	args := []interface{}{userID}

	if filter.Type != "" {
		where = append(where, "type = ?")
		args = append(args, filter.Type)
	}
	if filter.Priority != "" {
		where = append(where, "priority = ?")
		args = append(args, filter.Priority)
	}
	if filter.Category != "" {
		where = append(where, "category = ?")
		args = append(args, filter.Category)
	}

	query := fmt.Sprintf(`
		SELECT id, user_id, type, priority, title, description, savings, effort, impact, category, resources
		FROM recommendations WHERE %s ORDER BY id DESC
	`, strings.Join(where, " AND "))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.DatabaseError("Failed to list recommendations", err)
	}
	defer rows.Close()

	var recs []*recommendation.Recommendation
	for rows.Next() {
		var rec recommendation.Recommendation
		var resourcesJSON string
		err := rows.Scan(&rec.ID, &rec.UserID, &rec.Type, &rec.Priority, &rec.Title, &rec.Description, &rec.Savings, &rec.Effort, &rec.Impact, &rec.Category, &resourcesJSON)
		if err != nil {
			return nil, errors.DatabaseError("Failed to scan recommendation", err)
		}
		json.Unmarshal([]byte(resourcesJSON), &rec.Resources)
		recs = append(recs, &rec)
	}

	return recs, rows.Err()
}

func (r *RecommendationRepository) ListWithPagination(ctx context.Context, userID int64, filter recommendation.Filter, limit, offset int) ([]*recommendation.Recommendation, int64, error) {
	where := []string{"user_id = ?"}
	args := []interface{}{userID}

	if filter.Type != "" {
		where = append(where, "type = ?")
		args = append(args, filter.Type)
	}
	if filter.Priority != "" {
		where = append(where, "priority = ?")
		args = append(args, filter.Priority)
	}

	whereClause := strings.Join(where, " AND ")

	var total int64
	err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM recommendations WHERE %s", whereClause), args...).Scan(&total)
	if err != nil {
		return nil, 0, errors.DatabaseError("Failed to count recommendations", err)
	}

	query := fmt.Sprintf(`
		SELECT id, user_id, type, priority, title, description, savings, effort, impact, category, resources
		FROM recommendations WHERE %s ORDER BY id DESC LIMIT ? OFFSET ?
	`, whereClause)

	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, errors.DatabaseError("Failed to list recommendations", err)
	}
	defer rows.Close()

	var recs []*recommendation.Recommendation
	for rows.Next() {
		var rec recommendation.Recommendation
		var resourcesJSON string
		err := rows.Scan(&rec.ID, &rec.UserID, &rec.Type, &rec.Priority, &rec.Title, &rec.Description, &rec.Savings, &rec.Effort, &rec.Impact, &rec.Category, &resourcesJSON)
		if err != nil {
			return nil, 0, errors.DatabaseError("Failed to scan recommendation", err)
		}
		json.Unmarshal([]byte(resourcesJSON), &rec.Resources)
		recs = append(recs, &rec)
	}

	return recs, total, rows.Err()
}

func (r *RecommendationRepository) GetTotalSavings(ctx context.Context, userID int64) (float64, error) {
	var total float64
	err := r.db.QueryRowContext(ctx, "SELECT COALESCE(SUM(savings), 0) FROM recommendations WHERE user_id = ?", userID).Scan(&total)
	if err != nil {
		return 0, errors.DatabaseError("Failed to get total savings", err)
	}
	return total, nil
}
