package providers

import (
	"context"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	armcompute "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5"
	armresources "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	armstorage "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"

	"infraaudit/backend/internal/services"
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
						out = append(out, services.CloudResource{ID: id, Name: name, Type: "VM", Provider: "azure", Region: region, Status: "unknown"})
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
						out = append(out, services.CloudResource{ID: id, Name: name, Type: "VMSS", Provider: "azure", Region: region, Status: "unknown"})
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
						out = append(out, services.CloudResource{ID: id, Name: name, Type: "Storage", Provider: "azure", Region: region, Status: "available"})
					}
				}
			} else {
				log.Printf("azure storage client error: %v", err)
			}
		}
	}

	return out, nil
}
