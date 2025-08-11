package main

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// AWSCredentials holds simple static credentials and region.
// In production prefer role-based auth or secure storage.
type AWSCredentials struct {
	AccessKeyID     string
	SecretAccessKey string
	Region          string
}

func awsListResources(ctx context.Context, creds AWSCredentials) ([]CloudResource, error) {
	// Load AWS config. Prefer provided static creds; fallback to default chain.
	var (
		cfg aws.Config
		err error
	)
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

	var out []CloudResource

	// Discover all regions
	ec2Global := ec2.NewFromConfig(cfg)
	regions := []string{}
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

	// EC2 instances across all regions
	for _, region := range regions {
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
					// Try to find Name tag
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
					out = append(out, CloudResource{ID: id, Name: name, Type: "EC2", Provider: "aws", Region: region, Status: state})
				}
			}
		}
	}

	// S3 buckets (account-wide)
	s3c := s3.NewFromConfig(cfg)
	if resp, err := s3c.ListBuckets(ctx, &s3.ListBucketsInput{}); err == nil {
		for _, b := range resp.Buckets {
			name := "bucket"
			if b.Name != nil {
				name = *b.Name
			}
			out = append(out, CloudResource{ID: "s3-" + name, Name: name, Type: "S3", Provider: "aws", Region: cfg.Region, Status: "available"})
		}
	} else {
		log.Printf("aws s3 list error: %v", err)
	}

	return out, nil
}

func nonEmpty(v string, def string) string {
	if v == "" {
		return def
	}
	return v
}
