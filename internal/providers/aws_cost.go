package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	cetypes "github.com/aws/aws-sdk-go-v2/service/costexplorer/types"

	"github.com/pratik-mahalle/infraudit/internal/domain/cost"
)

// FetchAWSCosts retrieves cost data from AWS Cost Explorer for the last 30 days.
// AWS Cost Explorer API is only accessible from us-east-1.
func FetchAWSCosts(ctx context.Context, creds AWSCredentials) ([]cost.Cost, error) {
	var cfg aws.Config
	var err error

	// Cost Explorer is only available in us-east-1
	region := "us-east-1"

	if creds.AccessKeyID != "" && creds.SecretAccessKey != "" {
		cfg, err = awsconfig.LoadDefaultConfig(ctx,
			awsconfig.WithRegion(region),
			awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(creds.AccessKeyID, creds.SecretAccessKey, "")),
		)
	} else {
		cfg, err = awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	}
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	ceClient := costexplorer.NewFromConfig(cfg)

	now := time.Now().UTC()
	startDate := now.AddDate(0, 0, -30).Format("2006-01-02")
	endDate := now.Format("2006-01-02")

	input := &costexplorer.GetCostAndUsageInput{
		TimePeriod: &cetypes.DateInterval{
			Start: aws.String(startDate),
			End:   aws.String(endDate),
		},
		Granularity: cetypes.GranularityDaily,
		Metrics:     []string{"UnblendedCost"},
		GroupBy: []cetypes.GroupDefinition{
			{
				Type: cetypes.GroupDefinitionTypeDimension,
				Key:  aws.String("SERVICE"),
			},
			{
				Type: cetypes.GroupDefinitionTypeDimension,
				Key:  aws.String("REGION"),
			},
		},
	}

	result, err := ceClient.GetCostAndUsage(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("AWS Cost Explorer API error: %w", err)
	}

	var costs []cost.Cost
	for _, resultByTime := range result.ResultsByTime {
		costDate, err := time.Parse("2006-01-02", *resultByTime.TimePeriod.Start)
		if err != nil {
			continue
		}

		for _, group := range resultByTime.Groups {
			serviceName := ""
			regionName := ""
			if len(group.Keys) > 0 {
				serviceName = group.Keys[0]
			}
			if len(group.Keys) > 1 {
				regionName = group.Keys[1]
			}

			amount := 0.0
			if metric, ok := group.Metrics["UnblendedCost"]; ok && metric.Amount != nil {
				amount, _ = strconv.ParseFloat(*metric.Amount, 64)
			}

			if amount == 0 {
				continue
			}

			details, _ := json.Marshal(map[string]interface{}{
				"service": serviceName,
				"region":  regionName,
				"metrics": group.Metrics,
			})

			costs = append(costs, cost.Cost{
				Provider:    cost.ProviderAWS,
				ServiceName: serviceName,
				Region:      regionName,
				CostDate:    costDate,
				DailyCost:   amount,
				Currency:    "USD",
				CostDetails: json.RawMessage(details),
			})
		}
	}

	return costs, nil
}
