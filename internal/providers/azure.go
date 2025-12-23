package providers

import (
	"context"
	"encoding/json"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	armcompute "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5"
	armresources "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	armstorage "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"

	"github.com/pratik-mahalle/infraudit/internal/services"
)

type AzureCredentials struct {
	TenantID       string
	ClientID       string
	ClientSecret   string
	SubscriptionID string
	Location       string
}

func AzureListResources(ctx context.Context, creds AzureCredentials) ([]services.CloudResource, error) {
	cred, err := azidentity.NewClientSecretCredential(creds.TenantID, creds.ClientID, creds.ClientSecret, nil)
	if err != nil {
		return nil, err
	}

	var out []services.CloudResource

	rgClient, err := armresources.NewResourceGroupsClient(creds.SubscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}
	rgPager := rgClient.NewListPager(nil)
	for rgPager.More() {
		page, err := rgPager.NextPage(ctx)
		if err != nil {
			log.Printf("azure list resource groups error: %v", err)
			break
		}
		for _, rg := range page.Value {
			group := *rg.Name

			vmClient, err := armcompute.NewVirtualMachinesClient(creds.SubscriptionID, cred, nil)
			if err == nil {
				vmPager := vmClient.NewListPager(group, nil)
				for vmPager.More() {
					vmPage, err := vmPager.NextPage(ctx)
					if err != nil {
						log.Printf("azure list vms error: %v", err)
						break
					}
					for _, vm := range vmPage.Value {
						name := "vm"
						if vm.Name != nil {
							name = *vm.Name
						}
						id := ""
						if vm.ID != nil {
							id = *vm.ID
						}
						region := creds.Location
						if vm.Location != nil && *vm.Location != "" {
							region = *vm.Location
						}

						// Build full configuration
						config := buildAzureVMConfiguration(vm)
						configJSON, _ := json.Marshal(config)

						out = append(out, services.CloudResource{
							ID:            id,
							Name:          name,
							Type:          "VM",
							Provider:      "azure",
							Region:        region,
							Status:        "unknown",
							Configuration: string(configJSON),
						})
					}
				}
			} else {
				log.Printf("azure vm client error: %v", err)
			}

			vmssClient, err := armcompute.NewVirtualMachineScaleSetsClient(creds.SubscriptionID, cred, nil)
			if err == nil {
				vmssPager := vmssClient.NewListPager(group, nil)
				for vmssPager.More() {
					vmssPage, err := vmssPager.NextPage(ctx)
					if err != nil {
						log.Printf("azure list vmss error: %v", err)
						break
					}
					for _, ss := range vmssPage.Value {
						name := "vmss"
						if ss.Name != nil {
							name = *ss.Name
						}
						id := ""
						if ss.ID != nil {
							id = *ss.ID
						}
						region := creds.Location
						if ss.Location != nil && *ss.Location != "" {
							region = *ss.Location
						}

						// Build full configuration
						config := buildAzureVMSSConfiguration(ss)
						configJSON, _ := json.Marshal(config)

						out = append(out, services.CloudResource{
							ID:            id,
							Name:          name,
							Type:          "VMSS",
							Provider:      "azure",
							Region:        region,
							Status:        "unknown",
							Configuration: string(configJSON),
						})
					}
				}
			} else {
				log.Printf("azure vmss client error: %v", err)
			}

			stClient, err := armstorage.NewAccountsClient(creds.SubscriptionID, cred, nil)
			if err == nil {
				stPager := stClient.NewListByResourceGroupPager(group, nil)
				for stPager.More() {
					stPage, err := stPager.NextPage(ctx)
					if err != nil {
						log.Printf("azure list storage error: %v", err)
						break
					}
					for _, acc := range stPage.Value {
						name := "storage"
						if acc.Name != nil {
							name = *acc.Name
						}
						id := ""
						if acc.ID != nil {
							id = *acc.ID
						}
						region := creds.Location
						if acc.Location != nil && *acc.Location != "" {
							region = *acc.Location
						}

						// Build full configuration
						config := buildAzureStorageConfiguration(acc)
						configJSON, _ := json.Marshal(config)

						out = append(out, services.CloudResource{
							ID:            id,
							Name:          name,
							Type:          "Storage",
							Provider:      "azure",
							Region:        region,
							Status:        "available",
							Configuration: string(configJSON),
						})
					}
				}
			} else {
				log.Printf("azure storage client error: %v", err)
			}
		}
	}

	return out, nil
}

// buildAzureVMConfiguration creates a comprehensive configuration object for Azure VMs
func buildAzureVMConfiguration(vm *armcompute.VirtualMachine) map[string]interface{} {
	config := map[string]interface{}{
		"vm_id":    ptrStr(vm.ID),
		"name":     ptrStr(vm.Name),
		"location": ptrStr(vm.Location),
	}

	// VM size
	if vm.Properties != nil && vm.Properties.HardwareProfile != nil {
		config["vm_size"] = string(*vm.Properties.HardwareProfile.VMSize)
	}

	// Tags
	if vm.Tags != nil {
		tags := make(map[string]string)
		for k, v := range vm.Tags {
			if v != nil {
				tags[k] = *v
			}
		}
		config["tags"] = tags
	}

	// Network interfaces
	if vm.Properties != nil && vm.Properties.NetworkProfile != nil && vm.Properties.NetworkProfile.NetworkInterfaces != nil {
		nics := make([]map[string]string, 0)
		for _, nic := range vm.Properties.NetworkProfile.NetworkInterfaces {
			nics = append(nics, map[string]string{
				"id": ptrStr(nic.ID),
			})
		}
		config["network_interfaces"] = nics
	}

	// OS disk encryption
	if vm.Properties != nil && vm.Properties.StorageProfile != nil && vm.Properties.StorageProfile.OSDisk != nil {
		osDisk := vm.Properties.StorageProfile.OSDisk
		config["os_disk"] = map[string]interface{}{
			"name":          ptrStr(osDisk.Name),
			"create_option": string(*osDisk.CreateOption),
		}
		if osDisk.ManagedDisk != nil && osDisk.ManagedDisk.DiskEncryptionSet != nil {
			config["encryption"] = map[string]interface{}{
				"enabled":             true,
				"disk_encryption_set": ptrStr(osDisk.ManagedDisk.DiskEncryptionSet.ID),
			}
		} else {
			config["encryption"] = map[string]interface{}{
				"enabled": false,
			}
		}
	}

	// Identity (managed identity)
	if vm.Identity != nil {
		config["identity"] = map[string]interface{}{
			"type": string(*vm.Identity.Type),
		}
	}

	// Security profile
	if vm.Properties != nil && vm.Properties.SecurityProfile != nil {
		security := make(map[string]interface{})
		if vm.Properties.SecurityProfile.EncryptionAtHost != nil {
			security["encryption_at_host"] = *vm.Properties.SecurityProfile.EncryptionAtHost
		}
		if vm.Properties.SecurityProfile.SecurityType != nil {
			security["security_type"] = string(*vm.Properties.SecurityProfile.SecurityType)
		}
		config["security_profile"] = security
	}

	return config
}

// buildAzureVMSSConfiguration creates a comprehensive configuration object for Azure VMSS
func buildAzureVMSSConfiguration(vmss *armcompute.VirtualMachineScaleSet) map[string]interface{} {
	config := map[string]interface{}{
		"vmss_id":  ptrStr(vmss.ID),
		"name":     ptrStr(vmss.Name),
		"location": ptrStr(vmss.Location),
	}

	// SKU
	if vmss.SKU != nil {
		config["sku"] = map[string]interface{}{
			"name":     ptrStr(vmss.SKU.Name),
			"tier":     ptrStr(vmss.SKU.Tier),
			"capacity": ptrInt64(vmss.SKU.Capacity),
		}
	}

	// Tags
	if vmss.Tags != nil {
		tags := make(map[string]string)
		for k, v := range vmss.Tags {
			if v != nil {
				tags[k] = *v
			}
		}
		config["tags"] = tags
	}

	// Upgrade policy
	if vmss.Properties != nil && vmss.Properties.UpgradePolicy != nil {
		config["upgrade_policy"] = map[string]interface{}{
			"mode": string(*vmss.Properties.UpgradePolicy.Mode),
		}
	}

	// Identity
	if vmss.Identity != nil {
		config["identity"] = map[string]interface{}{
			"type": string(*vmss.Identity.Type),
		}
	}

	return config
}

// buildAzureStorageConfiguration creates a comprehensive configuration object for Azure Storage accounts
func buildAzureStorageConfiguration(acc *armstorage.Account) map[string]interface{} {
	config := map[string]interface{}{
		"account_id": ptrStr(acc.ID),
		"name":       ptrStr(acc.Name),
		"location":   ptrStr(acc.Location),
	}

	// SKU
	if acc.SKU != nil {
		config["sku"] = map[string]interface{}{
			"name": string(*acc.SKU.Name),
			"tier": string(*acc.SKU.Tier),
		}
	}

	// Kind
	if acc.Kind != nil {
		config["kind"] = string(*acc.Kind)
	}

	// Tags
	if acc.Tags != nil {
		tags := make(map[string]string)
		for k, v := range acc.Tags {
			if v != nil {
				tags[k] = *v
			}
		}
		config["tags"] = tags
	}

	// Properties
	if acc.Properties != nil {
		props := make(map[string]interface{})

		// Encryption
		if acc.Properties.Encryption != nil && acc.Properties.Encryption.Services != nil {
			encryption := make(map[string]interface{})
			if acc.Properties.Encryption.Services.Blob != nil {
				encryption["blob_enabled"] = ptrBool(acc.Properties.Encryption.Services.Blob.Enabled)
			}
			if acc.Properties.Encryption.Services.File != nil {
				encryption["file_enabled"] = ptrBool(acc.Properties.Encryption.Services.File.Enabled)
			}
			props["encryption"] = encryption
		}

		// HTTPS only
		if acc.Properties.EnableHTTPSTrafficOnly != nil {
			props["https_only"] = *acc.Properties.EnableHTTPSTrafficOnly
		}

		// Public network access
		if acc.Properties.PublicNetworkAccess != nil {
			props["public_network_access"] = string(*acc.Properties.PublicNetworkAccess)
		}

		// Minimum TLS version
		if acc.Properties.MinimumTLSVersion != nil {
			props["minimum_tls_version"] = string(*acc.Properties.MinimumTLSVersion)
		}

		// Allow blob public access
		if acc.Properties.AllowBlobPublicAccess != nil {
			props["allow_blob_public_access"] = *acc.Properties.AllowBlobPublicAccess
		}

		config["properties"] = props
	}

	return config
}

// Helper functions for pointer dereferencing
func ptrStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func ptrInt64(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}

func ptrBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}
