package acctest

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"

	"github.com/reserve-protocol/terraform-provider-dokploy/internal/provider"
)

// ProtoV6ProviderFactories returns provider factories for acceptance tests.
var ProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"dokploy": providerserver.NewProtocol6WithError(provider.New("test")()),
}

// TestAccPreCheck validates required environment variables before running acceptance tests.
func TestAccPreCheck(t *testing.T) {
	t.Helper()

	if v := os.Getenv("DOKPLOY_HOST"); v == "" {
		t.Fatal("DOKPLOY_HOST must be set for acceptance tests")
	}

	if v := os.Getenv("DOKPLOY_API_KEY"); v == "" {
		t.Fatal("DOKPLOY_API_KEY must be set for acceptance tests")
	}
}

// SkipIfNotAccTest skips the test if TF_ACC is not set.
func SkipIfNotAccTest(t *testing.T) {
	t.Helper()

	if os.Getenv("TF_ACC") == "" {
		t.Skip("Skipping acceptance test; set TF_ACC=1 to run")
	}
}
