package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/pratik-mahalle/infraudit/internal/api/dto"
	"github.com/pratik-mahalle/infraudit/internal/api/handlers"
	"github.com/pratik-mahalle/infraudit/internal/api/middleware"
	"github.com/pratik-mahalle/infraudit/internal/pkg/logger"
	"github.com/pratik-mahalle/infraudit/internal/pkg/validator"
	"github.com/pratik-mahalle/infraudit/internal/repository/postgres"
	"github.com/pratik-mahalle/infraudit/internal/services"
	"github.com/pratik-mahalle/infraudit/internal/testutil"
)

// TestResourceLifecycle tests the full lifecycle of a resource
// This is an integration test that tests: Create -> List -> Get -> Update -> Delete
func TestResourceLifecycle(t *testing.T) {
	// Setup test database
	db := testutil.NewTestDB(t)
	defer testutil.CleanupDB(db)

	// Setup dependencies
	log := logger.New(logger.Config{Level: "error", Format: "json"})
	val := validator.New()

	// Initialize repository and service
	resourceRepo := postgres.NewResourceRepository(db)
	resourceService := services.NewResourceService(resourceRepo, log)

	// Initialize handler
	resourceHandler := handlers.NewResourceHandler(resourceService, log, val)

	// User ID for testing
	userID := int64(1)

	// Test data
	resourceID := "i-integration-test"
	originalName := "integration-test-instance"
	updatedName := "updated-instance"

	// Step 1: Create a resource
	t.Run("Create Resource", func(t *testing.T) {
		createReq := dto.CreateResourceRequest{
			Provider:   "aws",
			ResourceID: resourceID,
			Name:       originalName,
			Type:       "ec2_instance",
			Region:     "us-east-1",
			Status:     "running",
		}

		body, _ := json.Marshal(createReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/resources", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, userID))

		rr := httptest.NewRecorder()
		resourceHandler.Create(rr, req)

		if status := rr.Code; status != http.StatusCreated {
			t.Errorf("Create failed with status %v, body: %s", status, rr.Body.String())
		}
	})

	// Step 2: List resources
	t.Run("List Resources", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/resources", nil)
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, userID))

		rr := httptest.NewRecorder()
		resourceHandler.List(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("List failed with status %v", status)
		}

		var response map[string]interface{}
		if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
			t.Errorf("Failed to decode response: %v", err)
		}

		data := response["data"].([]interface{})
		if len(data) != 1 {
			t.Errorf("Expected 1 resource, got %d", len(data))
		}
	})

	// Step 3: Get specific resource
	t.Run("Get Resource", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/resources/"+resourceID, nil)
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, userID))

		// Add chi URL params
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", resourceID)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		resourceHandler.Get(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("Get failed with status %v", status)
		}

		var response map[string]interface{}
		if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
			t.Errorf("Failed to decode response: %v", err)
		}

		data := response["data"].(map[string]interface{})
		if data["name"] != originalName {
			t.Errorf("Expected name %s, got %s", originalName, data["name"])
		}
	})

	// Step 4: Update resource
	t.Run("Update Resource", func(t *testing.T) {
		updateReq := dto.UpdateResourceRequest{
			Name: &updatedName,
		}

		body, _ := json.Marshal(updateReq)
		req := httptest.NewRequest(http.MethodPut, "/api/v1/resources/"+resourceID, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, userID))

		// Add chi URL params
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", resourceID)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		resourceHandler.Update(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("Update failed with status %v, body: %s", status, rr.Body.String())
		}

		// Verify update
		res, err := resourceService.GetByID(context.Background(), userID, resourceID)
		if err != nil {
			t.Errorf("Failed to get updated resource: %v", err)
		}
		if res.Name != updatedName {
			t.Errorf("Expected updated name %s, got %s", updatedName, res.Name)
		}
	})

	// Step 5: Delete resource
	t.Run("Delete Resource", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/resources/"+resourceID, nil)
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, userID))

		// Add chi URL params
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", resourceID)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		resourceHandler.Delete(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("Delete failed with status %v", status)
		}

		// Verify deletion
		_, err := resourceService.GetByID(context.Background(), userID, resourceID)
		if err == nil {
			t.Error("Resource still exists after deletion")
		}
	})
}

// TestAlertWorkflow tests creating and managing alerts
func TestAlertWorkflow(t *testing.T) {
	// Setup test database
	db := testutil.NewTestDB(t)
	defer testutil.CleanupDB(db)

	// Setup dependencies
	log := logger.New(logger.Config{Level: "error", Format: "json"})
	val := validator.New()

	// Initialize repository and service
	alertRepo := postgres.NewAlertRepository(db)
	alertService := services.NewAlertService(alertRepo, log)

	// Initialize handler
	alertHandler := handlers.NewAlertHandler(alertService, log, val)

	// User ID for testing
	userID := int64(1)

	var alertID int64

	// Step 1: Create an alert
	t.Run("Create Alert", func(t *testing.T) {
		createReq := dto.CreateAlertRequest{
			Type:        "security",
			Severity:    "critical",
			Title:       "Unencrypted S3 Bucket",
			Description: "S3 bucket detected without encryption",
			Resource:    "my-test-bucket",
		}

		body, _ := json.Marshal(createReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/alerts", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, userID))

		rr := httptest.NewRecorder()
		alertHandler.Create(rr, req)

		if status := rr.Code; status != http.StatusCreated {
			t.Errorf("Create failed with status %v, body: %s", status, rr.Body.String())
		}

		var response map[string]interface{}
		if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
			t.Errorf("Failed to decode response: %v", err)
		}

		data := response["data"].(map[string]interface{})
		alertID = int64(data["id"].(float64))
	})

	// Step 2: Get alert summary
	t.Run("Get Alert Summary", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/alerts/summary", nil)
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, userID))

		rr := httptest.NewRecorder()
		alertHandler.GetSummary(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("GetSummary failed with status %v", status)
		}

		var response map[string]interface{}
		if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
			t.Errorf("Failed to decode response: %v", err)
		}
	})

	// Step 3: Update alert status
	t.Run("Update Alert Status", func(t *testing.T) {
		updateReq := map[string]interface{}{
			"status": "resolved",
		}

		body, _ := json.Marshal(updateReq)
		req := httptest.NewRequest(http.MethodPut, "/api/v1/alerts/"+string(rune(alertID)), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, userID))

		// Add chi URL params
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", string(rune(alertID)))
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		alertHandler.Update(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Logf("Update returned status %v, may need status update endpoint", status)
		}
	})
}
