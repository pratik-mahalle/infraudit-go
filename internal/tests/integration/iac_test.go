package integration

import (
	"context"
	"os"
	"testing"

	"github.com/pratik-mahalle/infraudit/internal/domain/iac"
	"github.com/pratik-mahalle/infraudit/internal/repository/postgres"
	"github.com/pratik-mahalle/infraudit/internal/services"
	"github.com/pratik-mahalle/infraudit/internal/testutil"
)

func TestIaCParsing(t *testing.T) {
	db := testutil.NewTestDB(t)
	defer testutil.CleanupDB(db)

	iacRepo := postgres.NewIaCRepository(db)
	iacSvc := services.NewIaCService(iacRepo, nil, nil)

	ctx := context.Background()
	userID := "1"

	// Read Sample TF
	path := "../../../testdata/iac/example.tf"
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read sample tf: %v", err)
	}

	t.Run("Parse Terraform Success", func(t *testing.T) {
		def, err := iacSvc.UploadAndParse(ctx, userID, "test-definition", iac.IaCTypeTerraform, string(content))
		if err != nil {
			t.Fatalf("UploadAndParse failed: %v", err)
		}

		if def.ID == "" {
			t.Error("ID should be set")
		}
		if def.ParsedResources == nil {
			t.Error("ParsedResources should not be nil")
		}

		// Verify resources saved in DB
		resources, err := iacRepo.ListResourcesByDefinition(ctx, userID, def.ID)
		if err != nil {
			t.Fatalf("Failed to list resources: %v", err)
		}

		if len(resources) < 2 {
			t.Errorf("Expected at least 2 resources, got %d", len(resources))
		}

		foundInstance := false
		foundBucket := false
		for _, r := range resources {
			if r.ResourceType == "ec2_instance" {
				foundInstance = true
			}
			if r.ResourceType == "s3_bucket" {
				foundBucket = true
			}
		}

		if !foundInstance {
			t.Error("Did not find aws_instance")
		}
		if !foundBucket {
			t.Error("Did not find aws_s3_bucket")
		}
	})
}
