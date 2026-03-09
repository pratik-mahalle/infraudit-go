package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/pratik-mahalle/infraudit/internal/domain/recommendation"
	"github.com/pratik-mahalle/infraudit/internal/pkg/errors"
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

	status := rec.Status
	if status == "" {
		status = "pending"
	}

	query := `
		INSERT INTO recommendations (user_id, type, priority, title, description, estimated_savings, effort, impact, category, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`

	var id int64
	err := r.db.QueryRowContext(ctx, query,
		rec.UserID, rec.Type, rec.Priority, rec.Title, rec.Description, rec.Savings, rec.Effort, rec.Impact, rec.Category, status,
	).Scan(&id)
	if err != nil {
		return 0, errors.DatabaseError("Failed to create recommendation", err)
	}

	return id, nil
}

func (r *RecommendationRepository) GetByID(ctx context.Context, userID int64, id int64) (*recommendation.Recommendation, error) {
	query := `
		SELECT id, user_id, type, priority, title, description, estimated_savings, effort, impact, category, status
		FROM recommendations WHERE user_id = $1 AND id = $2
	`

	var rec recommendation.Recommendation
	err := r.db.QueryRowContext(ctx, query, userID, id).Scan(
		&rec.ID, &rec.UserID, &rec.Type, &rec.Priority, &rec.Title, &rec.Description, &rec.Savings, &rec.Effort, &rec.Impact, &rec.Category, &rec.Status,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NotFound("Recommendation")
	}
	if err != nil {
		return nil, errors.DatabaseError("Failed to get recommendation", err)
	}

	rec.Resources = []string{}
	return &rec, nil
}

func (r *RecommendationRepository) Update(ctx context.Context, rec *recommendation.Recommendation) error {
	rec.UpdatedAt = time.Now()

	status := rec.Status
	if status == "" {
		status = "pending"
	}

	query := `
		UPDATE recommendations SET type = $1, priority = $2, title = $3, description = $4, estimated_savings = $5, effort = $6, impact = $7, category = $8, status = $9
		WHERE user_id = $10 AND id = $11
	`

	result, err := r.db.ExecContext(ctx, query,
		rec.Type, rec.Priority, rec.Title, rec.Description, rec.Savings, rec.Effort, rec.Impact, rec.Category, status, rec.UserID, rec.ID,
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
	result, err := r.db.ExecContext(ctx, "DELETE FROM recommendations WHERE user_id = $1 AND id = $2", userID, id)
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
	paramN := 1
	where := []string{fmt.Sprintf("user_id = $%d", paramN)}
	args := []interface{}{userID}
	paramN++

	if filter.Type != "" {
		where = append(where, fmt.Sprintf("type = $%d", paramN))
		args = append(args, filter.Type)
		paramN++
	}
	if filter.Priority != "" {
		where = append(where, fmt.Sprintf("priority = $%d", paramN))
		args = append(args, filter.Priority)
		paramN++
	}
	if filter.Category != "" {
		where = append(where, fmt.Sprintf("category = $%d", paramN))
		args = append(args, filter.Category)
		paramN++
	}
	if filter.Status != "" {
		where = append(where, fmt.Sprintf("status = $%d", paramN))
		args = append(args, filter.Status)
		paramN++
	}

	query := fmt.Sprintf(`
		SELECT id, user_id, type, priority, title, description, estimated_savings, effort, impact, category, status
		FROM recommendations WHERE %s ORDER BY id DESC
	`, strings.Join(where, " AND "))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.DatabaseError("Failed to list recommendations", err)
	}
	defer rows.Close()

	// Pre-allocate slice with reasonable capacity
	recs := make([]*recommendation.Recommendation, 0, 100)
	for rows.Next() {
		var rec recommendation.Recommendation
		err := rows.Scan(&rec.ID, &rec.UserID, &rec.Type, &rec.Priority, &rec.Title, &rec.Description, &rec.Savings, &rec.Effort, &rec.Impact, &rec.Category, &rec.Status)
		if err != nil {
			return nil, errors.DatabaseError("Failed to scan recommendation", err)
		}
		rec.Resources = []string{}
		recs = append(recs, &rec)
	}

	return recs, rows.Err()
}

func (r *RecommendationRepository) ListWithPagination(ctx context.Context, userID int64, filter recommendation.Filter, limit, offset int) ([]*recommendation.Recommendation, int64, error) {
	paramN := 1
	where := []string{fmt.Sprintf("user_id = $%d", paramN)}
	args := []interface{}{userID}
	paramN++

	if filter.Type != "" {
		where = append(where, fmt.Sprintf("type = $%d", paramN))
		args = append(args, filter.Type)
		paramN++
	}
	if filter.Priority != "" {
		where = append(where, fmt.Sprintf("priority = $%d", paramN))
		args = append(args, filter.Priority)
		paramN++
	}
	if filter.Status != "" {
		where = append(where, fmt.Sprintf("status = $%d", paramN))
		args = append(args, filter.Status)
		paramN++
	}

	whereClause := strings.Join(where, " AND ")

	// Use a copy of args for count query (before appending limit/offset)
	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)

	var total int64
	err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM recommendations WHERE %s", whereClause), countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, errors.DatabaseError("Failed to count recommendations", err)
	}

	query := fmt.Sprintf(`
		SELECT id, user_id, type, priority, title, description, estimated_savings, effort, impact, category, status
		FROM recommendations WHERE %s ORDER BY id DESC LIMIT $%d OFFSET $%d
	`, whereClause, paramN, paramN+1)

	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, errors.DatabaseError("Failed to list recommendations", err)
	}
	defer rows.Close()

	// Pre-allocate slice with expected capacity
	recs := make([]*recommendation.Recommendation, 0, limit)
	for rows.Next() {
		var rec recommendation.Recommendation
		err := rows.Scan(&rec.ID, &rec.UserID, &rec.Type, &rec.Priority, &rec.Title, &rec.Description, &rec.Savings, &rec.Effort, &rec.Impact, &rec.Category, &rec.Status)
		if err != nil {
			return nil, 0, errors.DatabaseError("Failed to scan recommendation", err)
		}
		rec.Resources = []string{}
		recs = append(recs, &rec)
	}

	return recs, total, rows.Err()
}

func (r *RecommendationRepository) GetTotalSavings(ctx context.Context, userID int64) (float64, error) {
	var total float64
	err := r.db.QueryRowContext(ctx, "SELECT COALESCE(SUM(estimated_savings), 0) FROM recommendations WHERE user_id = $1", userID).Scan(&total)
	if err != nil {
		return 0, errors.DatabaseError("Failed to get total savings", err)
	}
	return total, nil
}
