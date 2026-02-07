package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	"github.com/pratik-mahalle/infraudit/internal/domain/cost"
)

// GCPBillingCredentials holds credentials for GCP billing data access via BigQuery.
type GCPBillingCredentials struct {
	ProjectID          string
	ServiceAccountJSON string
	BillingDataset     string // e.g. "my_project.my_billing_dataset.gcp_billing_export_v1_XXXXXX"
}

// FetchGCPCosts retrieves cost data from a GCP BigQuery billing export table for the last 30 days.
func FetchGCPCosts(ctx context.Context, creds GCPBillingCredentials) ([]cost.Cost, error) {
	if creds.BillingDataset == "" {
		return nil, nil // No billing dataset configured, skip gracefully
	}

	var opts []option.ClientOption
	if creds.ServiceAccountJSON != "" {
		opts = append(opts, option.WithCredentialsJSON([]byte(creds.ServiceAccountJSON)))
	}

	client, err := bigquery.NewClient(ctx, creds.ProjectID, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create BigQuery client: %w", err)
	}
	defer client.Close()

	startDate := time.Now().UTC().AddDate(0, 0, -30).Format("2006-01-02")

	query := client.Query(fmt.Sprintf(`
		SELECT
			service.description AS service_name,
			IFNULL(location.region, 'global') AS region,
			DATE(usage_start_time) AS cost_date,
			SUM(cost) AS daily_cost,
			currency
		FROM %s
		WHERE DATE(usage_start_time) >= @start_date
		GROUP BY service_name, region, cost_date, currency
		ORDER BY cost_date ASC, daily_cost DESC
	`, fmt.Sprintf("`%s`", creds.BillingDataset)))

	query.Parameters = []bigquery.QueryParameter{
		{Name: "start_date", Value: startDate},
	}

	it, err := query.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("BigQuery query error: %w", err)
	}

	var costs []cost.Cost
	for {
		var row struct {
			ServiceName string           `bigquery:"service_name"`
			Region      string           `bigquery:"region"`
			CostDate    bigquery.NullDate `bigquery:"cost_date"`
			DailyCost   float64          `bigquery:"daily_cost"`
			Currency    string           `bigquery:"currency"`
		}

		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("BigQuery row read error: %w", err)
		}

		if row.DailyCost == 0 {
			continue
		}

		costDate := time.Date(
			row.CostDate.Date.Year, row.CostDate.Date.Month, row.CostDate.Date.Day,
			0, 0, 0, 0, time.UTC,
		)

		currency := row.Currency
		if currency == "" {
			currency = "USD"
		}

		details, _ := json.Marshal(map[string]interface{}{
			"service": row.ServiceName,
			"region":  row.Region,
		})

		costs = append(costs, cost.Cost{
			Provider:    cost.ProviderGCP,
			ServiceName: row.ServiceName,
			Region:      row.Region,
			CostDate:    costDate,
			DailyCost:   row.DailyCost,
			Currency:    currency,
			CostDetails: json.RawMessage(details),
		})
	}

	return costs, nil
}
