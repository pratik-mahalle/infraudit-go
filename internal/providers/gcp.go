package providers

import (
	"context"
	"encoding/json"
	"log"
	"strconv"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"

	"infraudit/backend/internal/services"
)

type GCPCredentials struct {
	ProjectID          string
	ServiceAccountJSON string
	Region             string
}

func GCPListResources(ctx context.Context, creds GCPCredentials) ([]services.CloudResource, error) {
	var opts []option.ClientOption
	if creds.ServiceAccountJSON != "" {
		opts = append(opts, option.WithCredentialsJSON([]byte(creds.ServiceAccountJSON)))
	}

	var out []services.CloudResource

	// Compute Engine instances across all zones via AggregatedList
	instClient, err := compute.NewInstancesRESTClient(ctx, opts...)
	if err == nil {
		defer instClient.Close()
		aggReq := &computepb.AggregatedListInstancesRequest{Project: creds.ProjectID}
		it := instClient.AggregatedList(ctx, aggReq)
		for {
			pair, err := it.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				log.Printf("gcp compute aggregated list error: %v", err)
				break
			}
			if pair.Value == nil || pair.Value.Instances == nil {
				continue
			}
			for _, inst := range pair.Value.Instances {
				name := inst.GetName()
				id := strconv.FormatInt(int64(inst.GetId()), 10)
				zone := inst.GetZone()
				status := inst.GetStatus()

				// Build full configuration
				config := buildGCEConfiguration(inst)
				configJSON, _ := json.Marshal(config)

				out = append(out, services.CloudResource{
					ID:            id,
					Name:          name,
					Type:          "GCE",
					Provider:      "gcp",
					Region:        zone,
					Status:        status,
					Configuration: string(configJSON),
				})
			}
		}
	} else {
		log.Printf("gcp compute client error: %v", err)
	}

	// Cloud Storage buckets
	stClient, err := storage.NewClient(ctx, opts...)
	if err == nil {
		defer stClient.Close()
		it := stClient.Buckets(ctx, creds.ProjectID)
		for {
			battrs, err := it.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				log.Printf("gcp storage list error: %v", err)
				break
			}

			// Build full configuration
			config := buildGCSConfiguration(battrs)
			configJSON, _ := json.Marshal(config)

			out = append(out, services.CloudResource{
				ID:            "gcs-" + battrs.Name,
				Name:          battrs.Name,
				Type:          "GCS",
				Provider:      "gcp",
				Region:        battrs.Location,
				Status:        "available",
				Configuration: string(configJSON),
			})
		}
	} else {
		log.Printf("gcp storage client error: %v", err)
	}

	return out, nil
}

// buildGCEConfiguration creates a comprehensive configuration object for GCE instances
func buildGCEConfiguration(inst *computepb.Instance) map[string]interface{} {
	config := map[string]interface{}{
		"instance_id":   strconv.FormatUint(inst.GetId(), 10),
		"name":          inst.GetName(),
		"machine_type":  inst.GetMachineType(),
		"status":        inst.GetStatus(),
		"zone":          inst.GetZone(),
	}

	// Tags
	if inst.Tags != nil && len(inst.Tags.Items) > 0 {
		config["tags"] = inst.Tags.Items
	}

	// Labels
	if len(inst.Labels) > 0 {
		config["labels"] = inst.Labels
	}

	// Network interfaces
	if len(inst.NetworkInterfaces) > 0 {
		networks := make([]map[string]interface{}, 0, len(inst.NetworkInterfaces))
		for _, ni := range inst.NetworkInterfaces {
			network := map[string]interface{}{
				"network":      ni.GetNetwork(),
				"subnetwork":   ni.GetSubnetwork(),
				"network_ip":   ni.GetNetworkIP(),
			}

			// Access configs (public IPs)
			if len(ni.AccessConfigs) > 0 {
				accessConfigs := make([]map[string]string, 0)
				for _, ac := range ni.AccessConfigs {
					accessConfigs = append(accessConfigs, map[string]string{
						"name":       ac.GetName(),
						"nat_ip":     ac.GetNatIP(),
						"type":       ac.GetType(),
					})
				}
				network["access_configs"] = accessConfigs
			}
			networks = append(networks, network)
		}
		config["network_interfaces"] = networks
	}

	// Service accounts (IAM)
	if len(inst.ServiceAccounts) > 0 {
		serviceAccounts := make([]map[string]interface{}, 0, len(inst.ServiceAccounts))
		for _, sa := range inst.ServiceAccounts {
			serviceAccounts = append(serviceAccounts, map[string]interface{}{
				"email":  sa.GetEmail(),
				"scopes": sa.Scopes,
			})
		}
		config["service_accounts"] = serviceAccounts
	}

	// Disks (encryption info)
	if len(inst.Disks) > 0 {
		disks := make([]map[string]interface{}, 0, len(inst.Disks))
		for _, disk := range inst.Disks {
			diskInfo := map[string]interface{}{
				"device_name": disk.GetDeviceName(),
				"boot":        disk.GetBoot(),
				"mode":        disk.GetMode(),
			}
			if disk.DiskEncryptionKey != nil {
				diskInfo["encryption"] = map[string]interface{}{
					"encrypted": true,
				}
			}
			disks = append(disks, diskInfo)
		}
		config["disks"] = disks
	}

	// Metadata (SSH keys, etc.)
	if inst.Metadata != nil && len(inst.Metadata.Items) > 0 {
		metadata := make(map[string]string)
		for _, item := range inst.Metadata.Items {
			metadata[item.GetKey()] = item.GetValue()
		}
		config["metadata"] = metadata
	}

	// Deletion protection
	config["deletion_protection"] = inst.GetDeletionProtection()

	// Shielded instance config
	if inst.ShieldedInstanceConfig != nil {
		config["shielded_instance"] = map[string]interface{}{
			"enable_secure_boot":          inst.ShieldedInstanceConfig.GetEnableSecureBoot(),
			"enable_vtpm":                 inst.ShieldedInstanceConfig.GetEnableVtpm(),
			"enable_integrity_monitoring": inst.ShieldedInstanceConfig.GetEnableIntegrityMonitoring(),
		}
	}

	return config
}

// buildGCSConfiguration creates a comprehensive configuration object for GCS buckets
func buildGCSConfiguration(attrs *storage.BucketAttrs) map[string]interface{} {
	config := map[string]interface{}{
		"bucket_name": attrs.Name,
		"location":    attrs.Location,
		"storage_class": attrs.StorageClass,
		"created":     attrs.Created.String(),
	}

	// Encryption
	if attrs.Encryption != nil {
		config["encryption"] = map[string]interface{}{
			"enabled":           true,
			"default_kms_key":   attrs.Encryption.DefaultKMSKeyName,
		}
	} else {
		config["encryption"] = map[string]interface{}{
			"enabled": false,
		}
	}

	// Versioning
	if attrs.VersioningEnabled {
		config["versioning"] = map[string]interface{}{
			"enabled": true,
		}
	} else {
		config["versioning"] = map[string]interface{}{
			"enabled": false,
		}
	}

	// Public access prevention
	config["public_access_prevention"] = string(attrs.PublicAccessPrevention)

	// IAM configuration
	if attrs.UniformBucketLevelAccess.Enabled {
		config["uniform_bucket_level_access"] = map[string]interface{}{
			"enabled": true,
			"locked_time": attrs.UniformBucketLevelAccess.LockedTime.String(),
		}
	} else {
		config["uniform_bucket_level_access"] = map[string]interface{}{
			"enabled": false,
		}
	}

	// Labels
	if len(attrs.Labels) > 0 {
		config["labels"] = attrs.Labels
	}

	// Lifecycle rules
	if len(attrs.Lifecycle.Rules) > 0 {
		rules := make([]map[string]interface{}, 0, len(attrs.Lifecycle.Rules))
		for _, rule := range attrs.Lifecycle.Rules {
			rules = append(rules, map[string]interface{}{
				"action": map[string]string{
					"type":          rule.Action.Type,
					"storage_class": rule.Action.StorageClass,
				},
			})
		}
		config["lifecycle_rules"] = rules
	}

	// CORS
	if len(attrs.CORS) > 0 {
		config["cors_enabled"] = true
	}

	// Logging
	if attrs.Logging != nil {
		config["logging"] = map[string]interface{}{
			"enabled":       true,
			"log_bucket":    attrs.Logging.LogBucket,
			"log_prefix":    attrs.Logging.LogObjectPrefix,
		}
	}

	return config
}
