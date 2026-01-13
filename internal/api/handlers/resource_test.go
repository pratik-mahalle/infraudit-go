package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/pratik-mahalle/infraudit/internal/api/dto"
	"github.com/pratik-mahalle/infraudit/internal/api/middleware"
	"github.com/pratik-mahalle/infraudit/internal/domain/resource"
	"github.com/pratik-mahalle/infraudit/internal/pkg/logger"
	"github.com/pratik-mahalle/infraudit/internal/pkg/validator"
	"github.com/pratik-mahalle/infraudit/internal/services"
	"github.com/pratik-mahalle/infraudit/internal/testutil"
)

func TestResourceHandler_List(t *testing.T) {
	mockRepo := testutil.NewMockResourceRepository()
	log := logger.New(logger.Config{Level: "error", Format: "json"})
	service := services.NewResourceService(mockRepo, log)
	val := validator.New()
	handler := NewResourceHandler(service, log, val)

	// Create test resources
	ctx := context.Background()
	service.Create(ctx, &resource.Resource{
		UserID:     1,
		Provider:   "aws",
		ResourceID: "i-1",
		Name:       "instance-1",
		Type:       "ec2_instance",
		Region:     "us-east-1",
		Status:     "running",
	})

	tests := []struct {
		name           string
		userID         int64
		queryParams    string
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "list all resources",
			userID:         1,
			queryParams:    "",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "list with pagination",
			userID:         1,
			queryParams:    "?page=1&page_size=10",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/resources"+tt.queryParams, nil)
			req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, tt.userID))
			rr := httptest.NewRecorder()

			handler.List(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			if rr.Code == http.StatusOK {
				var response map[string]interface{}
				if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
					t.Errorf("failed to decode response: %v", err)
				}
			}
		})
	}
}

func TestResourceHandler_Get(t *testing.T) {
	mockRepo := testutil.NewMockResourceRepository()
	log := logger.New(logger.Config{Level: "error", Format: "json"})
	service := services.NewResourceService(mockRepo, log)
	val := validator.New()
	handler := NewResourceHandler(service, log, val)

	// Create test resource
	ctx := context.Background()
	service.Create(ctx, &resource.Resource{
		UserID:     1,
		Provider:   "aws",
		ResourceID: "i-test",
		Name:       "test-instance",
		Type:       "ec2_instance",
		Region:     "us-east-1",
		Status:     "running",
	})

	tests := []struct {
		name           string
		userID         int64
		resourceID     string
		expectedStatus int
	}{
		{
			name:           "get existing resource",
			userID:         1,
			resourceID:     "i-test",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get non-existing resource",
			userID:         1,
			resourceID:     "i-nonexistent",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/resources/"+tt.resourceID, nil)
			req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, tt.userID))

			// Add chi URL params
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.resourceID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()

			handler.Get(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}
		})
	}
}

func TestResourceHandler_Create(t *testing.T) {
	mockRepo := testutil.NewMockResourceRepository()
	log := logger.New(logger.Config{Level: "error", Format: "json"})
	service := services.NewResourceService(mockRepo, log)
	val := validator.New()
	handler := NewResourceHandler(service, log, val)

	tests := []struct {
		name           string
		userID         int64
		requestBody    dto.CreateResourceRequest
		expectedStatus int
	}{
		{
			name:   "create valid resource",
			userID: 1,
			requestBody: dto.CreateResourceRequest{
				Provider:   "aws",
				ResourceID: "i-new",
				Name:       "new-instance",
				Type:       "ec2_instance",
				Region:     "us-east-1",
				Status:     "running",
			},
			expectedStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/resources", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, tt.userID))

			rr := httptest.NewRecorder()

			handler.Create(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v, body: %s", status, tt.expectedStatus, rr.Body.String())
			}
		})
	}
}

func TestResourceHandler_Update(t *testing.T) {
	mockRepo := testutil.NewMockResourceRepository()
	log := logger.New(logger.Config{Level: "error", Format: "json"})
	service := services.NewResourceService(mockRepo, log)
	val := validator.New()
	handler := NewResourceHandler(service, log, val)

	// Create test resource
	ctx := context.Background()
	service.Create(ctx, &resource.Resource{
		UserID:     1,
		Provider:   "aws",
		ResourceID: "i-update",
		Name:       "old-name",
		Type:       "ec2_instance",
		Region:     "us-east-1",
		Status:     "running",
	})

	newName := "new-name"
	newStatus := "stopped"
	requestBody := dto.UpdateResourceRequest{
		Name:   &newName,
		Status: &newStatus,
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/resources/i-update", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, int64(1)))

	// Add chi URL params
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "i-update")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handler.Update(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestResourceHandler_Delete(t *testing.T) {
	mockRepo := testutil.NewMockResourceRepository()
	log := logger.New(logger.Config{Level: "error", Format: "json"})
	service := services.NewResourceService(mockRepo, log)
	val := validator.New()
	handler := NewResourceHandler(service, log, val)

	// Create test resource
	ctx := context.Background()
	service.Create(ctx, &resource.Resource{
		UserID:     1,
		Provider:   "aws",
		ResourceID: "i-delete",
		Name:       "to-delete",
		Type:       "ec2_instance",
		Region:     "us-east-1",
		Status:     "running",
	})

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/resources/i-delete", nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, int64(1)))

	// Add chi URL params
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "i-delete")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handler.Delete(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}
