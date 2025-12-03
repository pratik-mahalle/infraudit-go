package testutil

import (
	"context"
	"fmt"

	"github.com/pratik-mahalle/infraudit/internal/domain/alert"
	"github.com/pratik-mahalle/infraudit/internal/domain/anomaly"
	"github.com/pratik-mahalle/infraudit/internal/domain/drift"
	"github.com/pratik-mahalle/infraudit/internal/domain/provider"
	"github.com/pratik-mahalle/infraudit/internal/domain/recommendation"
	"github.com/pratik-mahalle/infraudit/internal/domain/resource"
	"github.com/pratik-mahalle/infraudit/internal/domain/user"
)

// MockUserRepository is a mock implementation of user.Repository
type MockUserRepository struct {
	Users       map[int64]*user.User
	EmailIndex  map[string]*user.User
	NextID      int64
	CreateError error
	GetError    error
	UpdateError error
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		Users:      make(map[int64]*user.User),
		EmailIndex: make(map[string]*user.User),
		NextID:     1,
	}
}

func (m *MockUserRepository) Create(ctx context.Context, u *user.User) error {
	if m.CreateError != nil {
		return m.CreateError
	}
	u.ID = m.NextID
	m.NextID++
	m.Users[u.ID] = u
	m.EmailIndex[u.Email] = u
	return nil
}

func (m *MockUserRepository) GetByID(ctx context.Context, id int64) (*user.User, error) {
	if m.GetError != nil {
		return nil, m.GetError
	}
	u, ok := m.Users[id]
	if !ok {
		return nil, fmt.Errorf("user not found")
	}
	return u, nil
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	if m.GetError != nil {
		return nil, m.GetError
	}
	u, ok := m.EmailIndex[email]
	if !ok {
		return nil, fmt.Errorf("user not found")
	}
	return u, nil
}

func (m *MockUserRepository) Update(ctx context.Context, u *user.User) error {
	if m.UpdateError != nil {
		return m.UpdateError
	}
	if _, ok := m.Users[u.ID]; !ok {
		return fmt.Errorf("user not found")
	}
	m.Users[u.ID] = u
	m.EmailIndex[u.Email] = u
	return nil
}

func (m *MockUserRepository) Delete(ctx context.Context, id int64) error {
	if u, ok := m.Users[id]; ok {
		delete(m.EmailIndex, u.Email)
		delete(m.Users, id)
	}
	return nil
}

func (m *MockUserRepository) List(ctx context.Context, limit, offset int) ([]*user.User, int64, error) {
	var result []*user.User
	for _, u := range m.Users {
		result = append(result, u)
	}
	return result, int64(len(result)), nil
}

// MockResourceRepository is a mock implementation of resource.Repository
type MockResourceRepository struct {
	Resources   map[string]*resource.Resource
	CreateError error
	GetError    error
}

func NewMockResourceRepository() *MockResourceRepository {
	return &MockResourceRepository{
		Resources: make(map[string]*resource.Resource),
	}
}

func (m *MockResourceRepository) Create(ctx context.Context, r *resource.Resource) error {
	if m.CreateError != nil {
		return m.CreateError
	}
	key := r.ResourceID
	m.Resources[key] = r
	return nil
}

func (m *MockResourceRepository) GetByID(ctx context.Context, userID int64, resourceID string) (*resource.Resource, error) {
	if m.GetError != nil {
		return nil, m.GetError
	}
	r, ok := m.Resources[resourceID]
	if !ok || r.UserID != userID {
		return nil, fmt.Errorf("resource not found")
	}
	return r, nil
}

func (m *MockResourceRepository) List(ctx context.Context, userID int64, filter resource.Filter, limit, offset int) ([]*resource.Resource, int64, error) {
	var result []*resource.Resource
	for _, r := range m.Resources {
		if r.UserID == userID {
			result = append(result, r)
		}
	}
	return result, int64(len(result)), nil
}

func (m *MockResourceRepository) Update(ctx context.Context, r *resource.Resource) error {
	key := r.ResourceID
	if _, ok := m.Resources[key]; !ok {
		return fmt.Errorf("resource not found")
	}
	m.Resources[key] = r
	return nil
}

func (m *MockResourceRepository) Delete(ctx context.Context, userID int64, resourceID string) error {
	delete(m.Resources, resourceID)
	return nil
}

func (m *MockResourceRepository) DeleteByProvider(ctx context.Context, userID int64, provider string) error {
	for key, r := range m.Resources {
		if r.UserID == userID && r.Provider == provider {
			delete(m.Resources, key)
		}
	}
	return nil
}

func (m *MockResourceRepository) ListByProvider(ctx context.Context, userID int64, provider string) ([]*resource.Resource, error) {
	var result []*resource.Resource
	for _, r := range m.Resources {
		if r.UserID == userID && r.Provider == provider {
			result = append(result, r)
		}
	}
	return result, nil
}

func (m *MockResourceRepository) SaveBatch(ctx context.Context, userID int64, provider string, resources []*resource.Resource) error {
	for _, r := range resources {
		r.UserID = userID
		r.Provider = provider
		m.Resources[r.ResourceID] = r
	}
	return nil
}

// MockAlertRepository is a mock implementation of alert.Repository
type MockAlertRepository struct {
	Alerts map[int64]*alert.Alert
	NextID int64
}

func NewMockAlertRepository() *MockAlertRepository {
	return &MockAlertRepository{
		Alerts: make(map[int64]*alert.Alert),
		NextID: 1,
	}
}

func (m *MockAlertRepository) Create(ctx context.Context, a *alert.Alert) (int64, error) {
	id := m.NextID
	m.NextID++
	a.ID = id
	m.Alerts[id] = a
	return id, nil
}

func (m *MockAlertRepository) GetByID(ctx context.Context, userID int64, id int64) (*alert.Alert, error) {
	a, ok := m.Alerts[id]
	if !ok || a.UserID != userID {
		return nil, fmt.Errorf("alert not found")
	}
	return a, nil
}

func (m *MockAlertRepository) Update(ctx context.Context, a *alert.Alert) error {
	if _, ok := m.Alerts[a.ID]; !ok {
		return fmt.Errorf("alert not found")
	}
	m.Alerts[a.ID] = a
	return nil
}

func (m *MockAlertRepository) Delete(ctx context.Context, userID int64, id int64) error {
	delete(m.Alerts, id)
	return nil
}

func (m *MockAlertRepository) List(ctx context.Context, userID int64, filter alert.Filter) ([]*alert.Alert, error) {
	var result []*alert.Alert
	for _, a := range m.Alerts {
		if a.UserID == userID {
			result = append(result, a)
		}
	}
	return result, nil
}

func (m *MockAlertRepository) ListWithPagination(ctx context.Context, userID int64, filter alert.Filter, limit, offset int) ([]*alert.Alert, int64, error) {
	alerts, _ := m.List(ctx, userID, filter)
	return alerts, int64(len(alerts)), nil
}

func (m *MockAlertRepository) CountByStatus(ctx context.Context, userID int64) (map[string]int, error) {
	counts := make(map[string]int)
	for _, a := range m.Alerts {
		if a.UserID == userID {
			counts[a.Status]++
		}
	}
	return counts, nil
}

// MockRecommendationRepository mock
type MockRecommendationRepository struct {
	Recommendations map[int64]*recommendation.Recommendation
	NextID          int64
}

func NewMockRecommendationRepository() *MockRecommendationRepository {
	return &MockRecommendationRepository{
		Recommendations: make(map[int64]*recommendation.Recommendation),
		NextID:          1,
	}
}

func (m *MockRecommendationRepository) Create(ctx context.Context, r *recommendation.Recommendation) (int64, error) {
	id := m.NextID
	m.NextID++
	r.ID = id
	m.Recommendations[id] = r
	return id, nil
}

func (m *MockRecommendationRepository) GetByID(ctx context.Context, userID int64, id int64) (*recommendation.Recommendation, error) {
	r, ok := m.Recommendations[id]
	if !ok || r.UserID != userID {
		return nil, fmt.Errorf("recommendation not found")
	}
	return r, nil
}

func (m *MockRecommendationRepository) Update(ctx context.Context, r *recommendation.Recommendation) error {
	if _, ok := m.Recommendations[r.ID]; !ok {
		return fmt.Errorf("recommendation not found")
	}
	m.Recommendations[r.ID] = r
	return nil
}

func (m *MockRecommendationRepository) Delete(ctx context.Context, userID int64, id int64) error {
	delete(m.Recommendations, id)
	return nil
}

func (m *MockRecommendationRepository) List(ctx context.Context, userID int64, filter recommendation.Filter) ([]*recommendation.Recommendation, error) {
	var result []*recommendation.Recommendation
	for _, r := range m.Recommendations {
		if r.UserID == userID {
			result = append(result, r)
		}
	}
	return result, nil
}

func (m *MockRecommendationRepository) ListWithPagination(ctx context.Context, userID int64, filter recommendation.Filter, limit, offset int) ([]*recommendation.Recommendation, int64, error) {
	recs, _ := m.List(ctx, userID, filter)
	return recs, int64(len(recs)), nil
}

func (m *MockRecommendationRepository) GetTotalSavings(ctx context.Context, userID int64) (float64, error) {
	var total float64
	for _, r := range m.Recommendations {
		if r.UserID == userID {
			total += r.Savings
		}
	}
	return total, nil
}

// MockDriftRepository mock
type MockDriftRepository struct {
	Drifts map[int64]*drift.Drift
	NextID int64
}

func NewMockDriftRepository() *MockDriftRepository {
	return &MockDriftRepository{
		Drifts: make(map[int64]*drift.Drift),
		NextID: 1,
	}
}

func (m *MockDriftRepository) Create(ctx context.Context, d *drift.Drift) (int64, error) {
	id := m.NextID
	m.NextID++
	d.ID = id
	m.Drifts[id] = d
	return id, nil
}

func (m *MockDriftRepository) GetByID(ctx context.Context, userID int64, id int64) (*drift.Drift, error) {
	d, ok := m.Drifts[id]
	if !ok || d.UserID != userID {
		return nil, fmt.Errorf("drift not found")
	}
	return d, nil
}

func (m *MockDriftRepository) Update(ctx context.Context, d *drift.Drift) error {
	if _, ok := m.Drifts[d.ID]; !ok {
		return fmt.Errorf("drift not found")
	}
	m.Drifts[d.ID] = d
	return nil
}

func (m *MockDriftRepository) Delete(ctx context.Context, userID int64, id int64) error {
	delete(m.Drifts, id)
	return nil
}

func (m *MockDriftRepository) List(ctx context.Context, userID int64, filter drift.Filter) ([]*drift.Drift, error) {
	var result []*drift.Drift
	for _, d := range m.Drifts {
		if d.UserID == userID {
			result = append(result, d)
		}
	}
	return result, nil
}

func (m *MockDriftRepository) ListWithPagination(ctx context.Context, userID int64, filter drift.Filter, limit, offset int) ([]*drift.Drift, int64, error) {
	drifts, _ := m.List(ctx, userID, filter)
	return drifts, int64(len(drifts)), nil
}

func (m *MockDriftRepository) CountBySeverity(ctx context.Context, userID int64) (map[string]int, error) {
	counts := make(map[string]int)
	for _, d := range m.Drifts {
		if d.UserID == userID {
			counts[d.Severity]++
		}
	}
	return counts, nil
}

// MockAnomalyRepository mock
type MockAnomalyRepository struct {
	Anomalies map[int64]*anomaly.Anomaly
	NextID    int64
}

func NewMockAnomalyRepository() *MockAnomalyRepository {
	return &MockAnomalyRepository{
		Anomalies: make(map[int64]*anomaly.Anomaly),
		NextID:    1,
	}
}

func (m *MockAnomalyRepository) Create(ctx context.Context, a *anomaly.Anomaly) (int64, error) {
	id := m.NextID
	m.NextID++
	a.ID = id
	m.Anomalies[id] = a
	return id, nil
}

func (m *MockAnomalyRepository) GetByID(ctx context.Context, userID int64, id int64) (*anomaly.Anomaly, error) {
	a, ok := m.Anomalies[id]
	if !ok || a.UserID != userID {
		return nil, fmt.Errorf("anomaly not found")
	}
	return a, nil
}

func (m *MockAnomalyRepository) Update(ctx context.Context, a *anomaly.Anomaly) error {
	if _, ok := m.Anomalies[a.ID]; !ok {
		return fmt.Errorf("anomaly not found")
	}
	m.Anomalies[a.ID] = a
	return nil
}

func (m *MockAnomalyRepository) Delete(ctx context.Context, userID int64, id int64) error {
	delete(m.Anomalies, id)
	return nil
}

func (m *MockAnomalyRepository) List(ctx context.Context, userID int64, filter anomaly.Filter) ([]*anomaly.Anomaly, error) {
	var result []*anomaly.Anomaly
	for _, a := range m.Anomalies {
		if a.UserID == userID {
			result = append(result, a)
		}
	}
	return result, nil
}

func (m *MockAnomalyRepository) ListWithPagination(ctx context.Context, userID int64, filter anomaly.Filter, limit, offset int) ([]*anomaly.Anomaly, int64, error) {
	anomalies, _ := m.List(ctx, userID, filter)
	return anomalies, int64(len(anomalies)), nil
}

func (m *MockAnomalyRepository) CountBySeverity(ctx context.Context, userID int64) (map[string]int, error) {
	counts := make(map[string]int)
	for _, a := range m.Anomalies {
		if a.UserID == userID {
			counts[a.Severity]++
		}
	}
	return counts, nil
}

// MockProviderRepository mock
type MockProviderRepository struct {
	Providers map[string]*provider.Provider
}

func NewMockProviderRepository() *MockProviderRepository {
	return &MockProviderRepository{
		Providers: make(map[string]*provider.Provider),
	}
}

func (m *MockProviderRepository) Upsert(ctx context.Context, p *provider.Provider) error {
	key := p.Provider
	m.Providers[key] = p
	return nil
}

func (m *MockProviderRepository) GetByProvider(ctx context.Context, userID int64, providerType string) (*provider.Provider, error) {
	p, ok := m.Providers[providerType]
	if !ok || p.UserID != userID {
		return nil, fmt.Errorf("provider not found")
	}
	return p, nil
}

func (m *MockProviderRepository) List(ctx context.Context, userID int64) ([]*provider.Provider, error) {
	var result []*provider.Provider
	for _, p := range m.Providers {
		if p.UserID == userID {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *MockProviderRepository) Delete(ctx context.Context, userID int64, providerType string) error {
	delete(m.Providers, providerType)
	return nil
}

func (m *MockProviderRepository) UpdateSyncStatus(ctx context.Context, userID int64, providerType string, lastSynced interface{}) error {
	if p, ok := m.Providers[providerType]; ok && p.UserID == userID {
		// Update last synced
		return nil
	}
	return fmt.Errorf("provider not found")
}
