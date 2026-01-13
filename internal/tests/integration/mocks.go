package integration

import (
	"context"

	"github.com/pratik-mahalle/infraudit/internal/domain/resource"
	cloudproviders "github.com/pratik-mahalle/infraudit/internal/providers"
)

// MockCloudProviderClient implements services.CloudProviderClient for testing
type MockCloudProviderClient struct {
	AWSResources    []resource.Resource
	AzureResources  []resource.Resource
	GCPResources    []resource.Resource
	Err             error
	AWSListCalled   bool
	AzureListCalled bool
	GCPListCalled   bool
}

func (m *MockCloudProviderClient) AWSListResources(ctx context.Context, creds cloudproviders.AWSCredentials) ([]resource.Resource, error) {
	m.AWSListCalled = true
	if m.Err != nil {
		return nil, m.Err
	}
	return m.AWSResources, nil
}

func (m *MockCloudProviderClient) AzureListResources(ctx context.Context, creds cloudproviders.AzureCredentials) ([]resource.Resource, error) {
	m.AzureListCalled = true
	if m.Err != nil {
		return nil, m.Err
	}
	return m.AzureResources, nil
}

func (m *MockCloudProviderClient) GCPListResources(ctx context.Context, creds cloudproviders.GCPCredentials) ([]resource.Resource, error) {
	m.GCPListCalled = true
	if m.Err != nil {
		return nil, m.Err
	}
	return m.GCPResources, nil
}
