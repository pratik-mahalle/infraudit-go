package cloudformation

import (
	"fmt"

	"github.com/pratik-mahalle/infraudit/internal/domain/iac"
)

// ResourceMapper maps CloudFormation resources to InfraAudit domain resources
type ResourceMapper struct{}

// NewResourceMapper creates a new resource mapper
func NewResourceMapper() *ResourceMapper {
	return &ResourceMapper{}
}

// MapToIaCResources converts parsed CloudFormation resources to IaC resources
func (m *ResourceMapper) MapToIaCResources(parsed *ParsedCloudFormation, definitionID string, userID string) []iac.IaCResource {
	resources := make([]iac.IaCResource, 0, len(parsed.Resources))

	for _, cfRes := range parsed.Resources {
		iacRes := m.mapCloudFormationResource(cfRes, definitionID, userID)
		resources = append(resources, iacRes)
	}

	return resources
}

// mapCloudFormationResource maps a single CloudFormation resource
func (m *ResourceMapper) mapCloudFormationResource(cfRes CloudFormationResource, definitionID string, userID string) iac.IaCResource {
	return iac.IaCResource{
		IaCDefinitionID: definitionID,
		UserID:          userID,
		ResourceType:    m.mapResourceType(cfRes.Type),
		ResourceName:    cfRes.LogicalID,
		ResourceAddress: cfRes.LogicalID, // CloudFormation uses logical ID as address
		Provider:        "aws",            // CloudFormation is AWS-only
		Configuration:   cfRes.Properties,
	}
}

// mapResourceType maps CloudFormation resource types to InfraAudit resource types
func (m *ResourceMapper) mapResourceType(cfType string) string {
	// Map of CloudFormation types to InfraAudit types
	typeMap := map[string]string{
		// EC2
		"AWS::EC2::Instance":              "ec2_instance",
		"AWS::EC2::Volume":                "ebs_volume",
		"AWS::EC2::SecurityGroup":         "security_group",
		"AWS::EC2::NetworkInterface":      "network_interface",
		"AWS::EC2::EIP":                   "elastic_ip",
		"AWS::EC2::KeyPair":               "key_pair",

		// VPC
		"AWS::EC2::VPC":                   "vpc",
		"AWS::EC2::Subnet":                "subnet",
		"AWS::EC2::RouteTable":            "route_table",
		"AWS::EC2::InternetGateway":       "internet_gateway",
		"AWS::EC2::NatGateway":            "nat_gateway",
		"AWS::EC2::VPCGatewayAttachment":  "vpc_gateway_attachment",

		// S3
		"AWS::S3::Bucket":                 "s3_bucket",
		"AWS::S3::BucketPolicy":           "s3_bucket_policy",

		// IAM
		"AWS::IAM::Role":                  "iam_role",
		"AWS::IAM::Policy":                "iam_policy",
		"AWS::IAM::User":                  "iam_user",
		"AWS::IAM::Group":                 "iam_group",
		"AWS::IAM::InstanceProfile":       "iam_instance_profile",
		"AWS::IAM::ManagedPolicy":         "iam_managed_policy",

		// RDS
		"AWS::RDS::DBInstance":            "rds_instance",
		"AWS::RDS::DBCluster":             "rds_cluster",
		"AWS::RDS::DBSubnetGroup":         "rds_subnet_group",
		"AWS::RDS::DBParameterGroup":      "rds_parameter_group",

		// Lambda
		"AWS::Lambda::Function":           "lambda_function",
		"AWS::Lambda::Permission":         "lambda_permission",
		"AWS::Lambda::EventSourceMapping": "lambda_event_source_mapping",

		// ELB
		"AWS::ElasticLoadBalancing::LoadBalancer":         "elb",
		"AWS::ElasticLoadBalancingV2::LoadBalancer":       "alb",
		"AWS::ElasticLoadBalancingV2::TargetGroup":        "target_group",
		"AWS::ElasticLoadBalancingV2::Listener":           "alb_listener",

		// Auto Scaling
		"AWS::AutoScaling::AutoScalingGroup":      "autoscaling_group",
		"AWS::AutoScaling::LaunchConfiguration":   "launch_configuration",

		// CloudWatch
		"AWS::CloudWatch::Alarm":          "cloudwatch_alarm",
		"AWS::Logs::LogGroup":             "cloudwatch_log_group",

		// DynamoDB
		"AWS::DynamoDB::Table":            "dynamodb_table",

		// SNS
		"AWS::SNS::Topic":                 "sns_topic",
		"AWS::SNS::Subscription":          "sns_subscription",

		// SQS
		"AWS::SQS::Queue":                 "sqs_queue",

		// ECS
		"AWS::ECS::Cluster":               "ecs_cluster",
		"AWS::ECS::Service":               "ecs_service",
		"AWS::ECS::TaskDefinition":        "ecs_task_definition",

		// EKS
		"AWS::EKS::Cluster":               "eks_cluster",
		"AWS::EKS::Nodegroup":             "eks_nodegroup",

		// CloudFront
		"AWS::CloudFront::Distribution":   "cloudfront_distribution",

		// Route53
		"AWS::Route53::HostedZone":        "route53_hosted_zone",
		"AWS::Route53::RecordSet":         "route53_record",

		// Secrets Manager
		"AWS::SecretsManager::Secret":     "secrets_manager_secret",

		// KMS
		"AWS::KMS::Key":                   "kms_key",
		"AWS::KMS::Alias":                 "kms_alias",
	}

	// Check if we have a mapping
	if mapped, ok := typeMap[cfType]; ok {
		return mapped
	}

	// Return original type if no mapping found
	return cfType
}

// ExtractResourceIdentifiers extracts identifiers from CloudFormation resource
func (m *ResourceMapper) ExtractResourceIdentifiers(cfRes CloudFormationResource) map[string]string {
	identifiers := make(map[string]string)

	// Add logical ID as primary identifier
	identifiers["logical_id"] = cfRes.LogicalID

	// Extract type-specific identifiers from properties
	props := cfRes.Properties

	switch cfRes.Type {
	case "AWS::EC2::Instance":
		if instanceID, ok := props["InstanceId"].(string); ok {
			identifiers["instance_id"] = instanceID
		}
		if tags, ok := props["Tags"].([]interface{}); ok {
			for _, tag := range tags {
				if tagMap, ok := tag.(map[string]interface{}); ok {
					if key, ok := tagMap["Key"].(string); ok && key == "Name" {
						if value, ok := tagMap["Value"].(string); ok {
							identifiers["name"] = value
						}
					}
				}
			}
		}

	case "AWS::S3::Bucket":
		if bucketName, ok := props["BucketName"].(string); ok {
			identifiers["bucket_name"] = bucketName
		}

	case "AWS::RDS::DBInstance":
		if dbInstanceID, ok := props["DBInstanceIdentifier"].(string); ok {
			identifiers["db_instance_id"] = dbInstanceID
		}

	case "AWS::Lambda::Function":
		if functionName, ok := props["FunctionName"].(string); ok {
			identifiers["function_name"] = functionName
		}

	case "AWS::DynamoDB::Table":
		if tableName, ok := props["TableName"].(string); ok {
			identifiers["table_name"] = tableName
		}

	case "AWS::IAM::Role":
		if roleName, ok := props["RoleName"].(string); ok {
			identifiers["role_name"] = roleName
		}
	}

	return identifiers
}

// CompareResources compares two CloudFormation resource configurations
func (m *ResourceMapper) CompareResources(cfConfig, actualConfig map[string]interface{}) []iac.FieldChange {
	changes := make([]iac.FieldChange, 0)

	// Check all fields in CF config
	for key, cfValue := range cfConfig {
		actualValue, exists := actualConfig[key]

		if !exists {
			changes = append(changes, iac.FieldChange{
				Field:       key,
				IaCValue:    cfValue,
				ActualValue: nil,
				ChangeType:  "removed",
			})
			continue
		}

		if !m.valuesEqual(cfValue, actualValue) {
			changes = append(changes, iac.FieldChange{
				Field:       key,
				IaCValue:    cfValue,
				ActualValue: actualValue,
				ChangeType:  "modified",
			})
		}
	}

	// Check for fields in actual config not in CF
	for key, actualValue := range actualConfig {
		if _, exists := cfConfig[key]; !exists {
			changes = append(changes, iac.FieldChange{
				Field:       key,
				IaCValue:    nil,
				ActualValue: actualValue,
				ChangeType:  "added",
			})
		}
	}

	return changes
}

// valuesEqual compares two values for equality
func (m *ResourceMapper) valuesEqual(v1, v2 interface{}) bool {
	return fmt.Sprintf("%v", v1) == fmt.Sprintf("%v", v2)
}
