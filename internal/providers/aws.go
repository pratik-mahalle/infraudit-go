package providers

import (
	"context"
	"log"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"infraaudit/backend/internal/services"
)

type AWSCredentials struct {
	AccessKeyID     string
	SecretAccessKey string
	Region          string
}

// AWSListResources fetches EC2 instances and S3 buckets concurrently across regions.
func AWSListResources(ctx context.Context, creds AWSCredentials) ([]services.CloudResource, error) {
	var cfg aws.Config
	var err error
	if creds.AccessKeyID != "" && creds.SecretAccessKey != "" {
		cfg, err = awsconfig.LoadDefaultConfig(ctx,
			awsconfig.WithRegion(nonEmpty(creds.Region, "us-east-1")),
			awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(creds.AccessKeyID, creds.SecretAccessKey, "")),
		)
	} else {
		cfg, err = awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(nonEmpty(creds.Region, "us-east-1")))
	}
	if err != nil {
		return nil, err
	}

	resources := []services.CloudResource{}
	var mu sync.Mutex
	var wg sync.WaitGroup

	regions := []string{}
	ec2Global := ec2.NewFromConfig(cfg)
	if resp, err := ec2Global.DescribeRegions(ctx, &ec2.DescribeRegionsInput{}); err == nil {
		for _, r := range resp.Regions {
			if r.RegionName != nil {
				regions = append(regions, *r.RegionName)
			}
		}
	} else {
		log.Printf("aws describe regions error: %v", err)
	}
	if len(regions) == 0 {
		regions = []string{cfg.Region}
	}

	sem := make(chan struct{}, 5)
	for _, region := range regions {
		wg.Add(1)
		sem <- struct{}{}
		go func(region string) {
			defer wg.Done()
			defer func() { <-sem }()
			ec2Resources := fetchEC2InRegion(ctx, cfg, region)
			mu.Lock()
			resources = append(resources, ec2Resources...)
			mu.Unlock()
		}(region)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		s3Resources := fetchS3Buckets(ctx, cfg)
		mu.Lock()
		resources = append(resources, s3Resources...)
		mu.Unlock()
	}()

	wg.Wait()
	return resources, nil
}

func fetchEC2InRegion(ctx context.Context, cfg aws.Config, region string) []services.CloudResource {
	out := []services.CloudResource{}
	cfgRegional := cfg
	cfgRegional.Region = region
	ec2c := ec2.NewFromConfig(cfgRegional)
	p := ec2.NewDescribeInstancesPaginator(ec2c, &ec2.DescribeInstancesInput{})

	for p.HasMorePages() {
		page, err := p.NextPage(ctx)
		if err != nil {
			log.Printf("aws ec2 describe error in %s: %v", region, err)
			break
		}
		for _, res := range page.Reservations {
			for _, inst := range res.Instances {
				id := "unknown"
				if inst.InstanceId != nil {
					id = *inst.InstanceId
				}
				name := id
				for _, t := range inst.Tags {
					if t.Key != nil && *t.Key == "Name" && t.Value != nil {
						name = *t.Value
						break
					}
				}
				state := "unknown"
				if inst.State != nil && inst.State.Name != "" {
					state = string(inst.State.Name)
				}
				out = append(out, services.CloudResource{ID: id, Name: name, Type: "EC2", Provider: "aws", Region: region, Status: state})
			}
		}
	}
	return out
}

func fetchS3Buckets(ctx context.Context, cfg aws.Config) []services.CloudResource {
	out := []services.CloudResource{}
	s3c := s3.NewFromConfig(cfg)
	if resp, err := s3c.ListBuckets(ctx, &s3.ListBucketsInput{}); err == nil {
		for _, b := range resp.Buckets {
			name := "bucket"
			if b.Name != nil {
				name = *b.Name
			}
			out = append(out, services.CloudResource{ID: "s3-" + name, Name: name, Type: "S3", Provider: "aws", Region: cfg.Region, Status: "available"})
		}
	} else {
		log.Printf("aws s3 list error: %v", err)
	}
	return out
}

func nonEmpty(v string, def string) string {
	if v == "" {
		return def
	}
	return v
}
