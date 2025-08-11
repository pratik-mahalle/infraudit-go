package main

import (
	"context"
	"encoding/json"
	"log"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
)

type GCPCredentials struct {
	ProjectID          string
	ServiceAccountJSON string
	Region             string
}

func gcpListResources(ctx context.Context, creds GCPCredentials) ([]CloudResource, error) {
	var opts []option.ClientOption
	if creds.ServiceAccountJSON != "" {
		opts = append(opts, option.WithCredentialsJSON([]byte(creds.ServiceAccountJSON)))
	}

	var out []CloudResource

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
			// pair.Key is zone/region, pair.Value.Instances is the list
			if pair.Value == nil || pair.Value.Instances == nil {
				continue
			}
			for _, inst := range pair.Value.Instances {
				name := inst.GetName()
				id := json.Number(inst.GetId()).String()
				zone := inst.GetZone()
				out = append(out, CloudResource{ID: id, Name: name, Type: "GCE", Provider: "gcp", Region: zone, Status: inst.GetStatus()})
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
			out = append(out, CloudResource{ID: "gcs-" + battrs.Name, Name: battrs.Name, Type: "GCS", Provider: "gcp", Region: battrs.Location, Status: "available"})
		}
	} else {
		log.Printf("gcp storage client error: %v", err)
	}

	return out, nil
}
