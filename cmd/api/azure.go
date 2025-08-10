package main

import (
	"context"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	armcompute "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5"
	armresources "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	armstorage "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
)

type AzureCredentials struct {
	TenantID       string
	ClientID       string
	ClientSecret   string
	SubscriptionID string
	Location       string
}

func azureListResources(ctx context.Context, creds AzureCredentials) ([]CloudResource, error) {
	cred, err := azidentity.NewClientSecretCredential(creds.TenantID, creds.ClientID, creds.ClientSecret, nil)
	if err != nil {
		return nil, err
	}

	var out []CloudResource

	// Enumerate resource groups and list VMs and Storage accounts within
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
						out = append(out, CloudResource{ID: id, Name: name, Type: "VM", Provider: "azure", Region: creds.Location, Status: "unknown"})
					}
				}
			} else {
				log.Printf("azure vm client error: %v", err)
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
						out = append(out, CloudResource{ID: id, Name: name, Type: "Storage", Provider: "azure", Region: creds.Location, Status: "available"})
					}
				}
			} else {
				log.Printf("azure storage client error: %v", err)
			}
		}
	}

	return out, nil
}
