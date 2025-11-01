package providers

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/pratik-mahalle/infraudit/internal/services"
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
				tags := make(map[string]string)
				for _, t := range inst.Tags {
					if t.Key != nil && t.Value != nil {
						if *t.Key == "Name" {
							name = *t.Value
						}
						tags[*t.Key] = *t.Value
					}
				}
				state := "unknown"
				if inst.State != nil && inst.State.Name != "" {
					state = string(inst.State.Name)
				}

				// Build full configuration
				config := buildEC2Configuration(inst, tags)
				configJSON, _ := json.Marshal(config)

				out = append(out, services.CloudResource{
					ID:            id,
					Name:          name,
					Type:          "EC2",
					Provider:      "aws",
					Region:        region,
					Status:        state,
					Configuration: string(configJSON),
				})
			}
		}
	}
	return out
}

// buildEC2Configuration creates a comprehensive configuration object for EC2 instances
func buildEC2Configuration(inst types.Instance, tags map[string]string) map[string]interface{} {
	config := map[string]interface{}{
		"instance_id":   ptrString(inst.InstanceId),
		"instance_type": string(inst.InstanceType),
		"state":         string(inst.State.Name),
		"tags":          tags,
	}

	// Security groups
	if len(inst.SecurityGroups) > 0 {
		sgs := make([]map[string]string, 0, len(inst.SecurityGroups))
		for _, sg := range inst.SecurityGroups {
			sgs = append(sgs, map[string]string{
				"id":   ptrString(sg.GroupId),
				"name": ptrString(sg.GroupName),
			})
		}
		config["security_groups"] = sgs
	}

	// IAM instance profile
	if inst.IamInstanceProfile != nil {
		config["iam_instance_profile"] = map[string]string{
			"arn": ptrString(inst.IamInstanceProfile.Arn),
		}
	}

	// Network configuration
	config["network"] = map[string]interface{}{
		"vpc_id":              ptrString(inst.VpcId),
		"subnet_id":           ptrString(inst.SubnetId),
		"private_ip_address":  ptrString(inst.PrivateIpAddress),
		"public_ip_address":   ptrString(inst.PublicIpAddress),
		"private_dns_name":    ptrString(inst.PrivateDnsName),
		"public_dns_name":     ptrString(inst.PublicDnsName),
	}

	// Encryption
	if inst.RootDeviceName != nil {
		encrypted := false
		for _, bdm := range inst.BlockDeviceMappings {
			if bdm.Ebs != nil && bdm.Ebs.VolumeId != nil {
				// Note: Would need additional API call to check EBS encryption
				encrypted = false // Placeholder
			}
		}
		config["encryption"] = map[string]interface{}{
			"ebs_encrypted": encrypted,
		}
	}

	// SSH key
	config["key_name"] = ptrString(inst.KeyName)

	// Monitoring
	config["monitoring"] = map[string]interface{}{
		"enabled": inst.Monitoring != nil && inst.Monitoring.State == "enabled",
	}

	// Launch time
	if inst.LaunchTime != nil {
		config["launch_time"] = inst.LaunchTime.String()
	}

	return config
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

			// Build full configuration
			config := buildS3Configuration(ctx, s3c, name)
			configJSON, _ := json.Marshal(config)

			out = append(out, services.CloudResource{
				ID:            "s3-" + name,
				Name:          name,
				Type:          "S3",
				Provider:      "aws",
				Region:        cfg.Region,
				Status:        "available",
				Configuration: string(configJSON),
			})
		}
	} else {
		log.Printf("aws s3 list error: %v", err)
	}
	return out
}

// buildS3Configuration creates a comprehensive configuration object for S3 buckets
func buildS3Configuration(ctx context.Context, s3c *s3.Client, bucketName string) map[string]interface{} {
	config := map[string]interface{}{
		"bucket_name": bucketName,
	}

	// Get bucket encryption
	if encResp, err := s3c.GetBucketEncryption(ctx, &s3.GetBucketEncryptionInput{
		Bucket: &bucketName,
	}); err == nil && encResp.ServerSideEncryptionConfiguration != nil {
		encrypted := len(encResp.ServerSideEncryptionConfiguration.Rules) > 0
		config["encryption"] = map[string]interface{}{
			"enabled": encrypted,
		}
	} else {
		config["encryption"] = map[string]interface{}{
			"enabled": false,
		}
	}

	// Get bucket versioning
	if verResp, err := s3c.GetBucketVersioning(ctx, &s3.GetBucketVersioningInput{
		Bucket: &bucketName,
	}); err == nil {
		config["versioning"] = map[string]interface{}{
			"enabled": verResp.Status == "Enabled",
			"status":  string(verResp.Status),
		}
	} else {
		config["versioning"] = map[string]interface{}{
			"enabled": false,
		}
	}

	// Get public access block configuration
	if pabResp, err := s3c.GetPublicAccessBlock(ctx, &s3.GetPublicAccessBlockInput{
		Bucket: &bucketName,
	}); err == nil && pabResp.PublicAccessBlockConfiguration != nil {
		pab := pabResp.PublicAccessBlockConfiguration
		config["public_access"] = map[string]interface{}{
			"block_public_acls":        pab.BlockPublicAcls,
			"ignore_public_acls":       pab.IgnorePublicAcls,
			"block_public_policy":      pab.BlockPublicPolicy,
			"restrict_public_buckets":  pab.RestrictPublicBuckets,
			"public_access_blocked":    pab.BlockPublicAcls && pab.IgnorePublicAcls && pab.BlockPublicPolicy && pab.RestrictPublicBuckets,
		}
	} else {
		config["public_access"] = map[string]interface{}{
			"public_access_blocked": false,
		}
	}

	// Get bucket ACL
	if aclResp, err := s3c.GetBucketAcl(ctx, &s3.GetBucketAclInput{
		Bucket: &bucketName,
	}); err == nil && aclResp.Grants != nil {
		grants := make([]map[string]interface{}, 0, len(aclResp.Grants))
		for _, grant := range aclResp.Grants {
			g := map[string]interface{}{
				"permission": string(grant.Permission),
			}
			if grant.Grantee != nil {
				g["grantee_type"] = string(grant.Grantee.Type)
				if grant.Grantee.URI != nil {
					g["grantee_uri"] = *grant.Grantee.URI
				}
			}
			grants = append(grants, g)
		}
		config["acl"] = grants
	}

	// Get bucket tags
	if tagResp, err := s3c.GetBucketTagging(ctx, &s3.GetBucketTaggingInput{
		Bucket: &bucketName,
	}); err == nil && tagResp.TagSet != nil {
		tags := make(map[string]string)
		for _, tag := range tagResp.TagSet {
			if tag.Key != nil && tag.Value != nil {
				tags[*tag.Key] = *tag.Value
			}
		}
		config["tags"] = tags
	}

	return config
}

// ptrString returns the value of a string pointer or empty string
func ptrString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func nonEmpty(v string, def string) string {
	if v == "" {
		return def
	}
	return v
}
