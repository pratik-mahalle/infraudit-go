package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/costmanagement/armcostmanagement"

	"github.com/pratik-mahalle/infraudit/internal/domain/cost"
)

// FetchAzureCosts retrieves cost data from Azure Cost Management for the last 30 days.
func FetchAzureCosts(ctx context.Context, creds AzureCredentials) ([]cost.Cost, error) {
	credential, err := azidentity.NewClientSecretCredential(creds.TenantID, creds.ClientID, creds.ClientSecret, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure credential: %w", err)
	}

	client, err := armcostmanagement.NewQueryClient(credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cost management client: %w", err)
	}

	now := time.Now().UTC()
	startDate := now.AddDate(0, 0, -30)

	scope := fmt.Sprintf("subscriptions/%s", creds.SubscriptionID)

	timePeriod := armcostmanagement.QueryTimePeriod{
		From: &startDate,
		To:   &now,
	}

	sumFunc := armcostmanagement.FunctionTypeSum
	queryAggregation := map[string]*armcostmanagement.QueryAggregation{
		"PreTaxCost": {
			Name:     ptrStr2("PreTaxCost"),
			Function: &sumFunc,
		},
	}

	dimGrouping := armcostmanagement.QueryColumnTypeDimension
	queryGrouping := []*armcostmanagement.QueryGrouping{
		{
			Type: &dimGrouping,
			Name: ptrStr2("ServiceName"),
		},
		{
			Type: &dimGrouping,
			Name: ptrStr2("ResourceLocation"),
		},
	}

	granularity := armcostmanagement.GranularityTypeDaily
	timeframeCustom := armcostmanagement.TimeframeTypeCustom
	exportTypeUsage := armcostmanagement.ExportTypeActualCost

	queryDef := armcostmanagement.QueryDefinition{
		Type:       &exportTypeUsage,
		Timeframe:  &timeframeCustom,
		TimePeriod: &timePeriod,
		Dataset: &armcostmanagement.QueryDataset{
			Granularity: &granularity,
			Aggregation: queryAggregation,
			Grouping:    queryGrouping,
		},
	}

	result, err := client.Usage(ctx, scope, queryDef, nil)
	if err != nil {
		return nil, fmt.Errorf("Azure Cost Management API error: %w", err)
	}

	if result.Properties == nil || result.Properties.Rows == nil {
		return nil, nil
	}

	// Build column index mapping
	colIndex := make(map[string]int)
	if result.Properties.Columns != nil {
		for i, col := range result.Properties.Columns {
			if col.Name != nil {
				colIndex[*col.Name] = i
			}
		}
	}

	costIdx, hasCost := colIndex["PreTaxCost"]
	serviceIdx, hasService := colIndex["ServiceName"]
	locationIdx, hasLocation := colIndex["ResourceLocation"]
	dateIdx, hasDate := colIndex["UsageDateKey"]
	if !hasDate {
		dateIdx, hasDate = colIndex["UsageDate"]
	}

	var costs []cost.Cost
	for _, row := range result.Properties.Rows {
		if len(row) == 0 {
			continue
		}

		var dailyCost float64
		if hasCost && costIdx < len(row) {
			if v, ok := row[costIdx].(float64); ok {
				dailyCost = v
			}
		}
		if dailyCost == 0 {
			continue
		}

		var serviceName, location string
		if hasService && serviceIdx < len(row) {
			if v, ok := row[serviceIdx].(string); ok {
				serviceName = v
			}
		}
		if hasLocation && locationIdx < len(row) {
			if v, ok := row[locationIdx].(string); ok {
				location = v
			}
		}

		var costDate time.Time
		if hasDate && dateIdx < len(row) {
			switch v := row[dateIdx].(type) {
			case float64:
				// Azure returns date as YYYYMMDD integer
				dateInt := int(v)
				year := dateInt / 10000
				month := (dateInt % 10000) / 100
				day := dateInt % 100
				costDate = time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
			case string:
				costDate, _ = time.Parse("2006-01-02", v)
			}
		}

		if costDate.IsZero() {
			continue
		}

		details, _ := json.Marshal(map[string]interface{}{
			"service":  serviceName,
			"location": location,
		})

		costs = append(costs, cost.Cost{
			Provider:    cost.ProviderAzure,
			ServiceName: serviceName,
			Region:      location,
			CostDate:    costDate,
			DailyCost:   dailyCost,
			Currency:    "USD",
			CostDetails: json.RawMessage(details),
		})
	}

	return costs, nil
}

func ptrStr2(s string) *string {
	return &s
}
