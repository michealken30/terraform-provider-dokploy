package provider

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestNew(t *testing.T) {
	version := "1.0.0"
	factory := New(version)

	if factory == nil {
		t.Fatal("expected factory function, got nil")
	}

	p := factory()
	if p == nil {
		t.Fatal("expected provider, got nil")
	}

	dokployProvider, ok := p.(*DokployProvider)
	if !ok {
		t.Fatal("expected *DokployProvider type")
	}

	if dokployProvider.version != version {
		t.Errorf("expected version %s, got %s", version, dokployProvider.version)
	}
}

func TestMetadata(t *testing.T) {
	version := "1.2.3"
	p := &DokployProvider{version: version}

	req := provider.MetadataRequest{}
	resp := &provider.MetadataResponse{}

	p.Metadata(context.Background(), req, resp)

	if resp.TypeName != "dokploy" {
		t.Errorf("expected TypeName 'dokploy', got %s", resp.TypeName)
	}

	if resp.Version != version {
		t.Errorf("expected Version %s, got %s", version, resp.Version)
	}
}

func TestSchema(t *testing.T) {
	p := &DokployProvider{}

	req := provider.SchemaRequest{}
	resp := &provider.SchemaResponse{}

	p.Schema(context.Background(), req, resp)

	// Check that schema has the expected attributes
	schema := resp.Schema

	if schema.Description == "" {
		t.Error("expected non-empty schema description")
	}

	// Check host attribute
	hostAttr, ok := schema.Attributes["host"]
	if !ok {
		t.Fatal("expected 'host' attribute in schema")
	}
	if hostAttr.GetDescription() == "" {
		t.Error("expected non-empty host description")
	}

	// Check api_key attribute
	apiKeyAttr, ok := schema.Attributes["api_key"]
	if !ok {
		t.Fatal("expected 'api_key' attribute in schema")
	}
	if apiKeyAttr.GetDescription() == "" {
		t.Error("expected non-empty api_key description")
	}
}

func TestConfigure_Success(t *testing.T) {
	// Clear environment variables to ensure test isolation
	os.Unsetenv("DOKPLOY_HOST")
	os.Unsetenv("DOKPLOY_API_KEY")

	p := &DokployProvider{version: "1.0.0"}

	// Create a mock config with host and api_key
	configType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"host":    tftypes.String,
			"api_key": tftypes.String,
		},
	}

	configValue := tftypes.NewValue(configType, map[string]tftypes.Value{
		"host":    tftypes.NewValue(tftypes.String, "https://dokploy.example.com"),
		"api_key": tftypes.NewValue(tftypes.String, "test-api-key"),
	})

	// Get schema to create proper config
	schemaReq := provider.SchemaRequest{}
	schemaResp := &provider.SchemaResponse{}
	p.Schema(context.Background(), schemaReq, schemaResp)

	config := tfsdk.Config{
		Schema: schemaResp.Schema,
		Raw:    configValue,
	}

	req := provider.ConfigureRequest{
		Config: config,
	}
	resp := &provider.ConfigureResponse{}

	p.Configure(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error: %v", resp.Diagnostics)
	}

	if resp.DataSourceData == nil {
		t.Error("expected DataSourceData to be set")
	}

	if resp.ResourceData == nil {
		t.Error("expected ResourceData to be set")
	}
}

func TestConfigure_FromEnv(t *testing.T) {
	// Set environment variables
	os.Setenv("DOKPLOY_HOST", "https://env.dokploy.example.com")
	os.Setenv("DOKPLOY_API_KEY", "env-api-key")
	defer func() {
		os.Unsetenv("DOKPLOY_HOST")
		os.Unsetenv("DOKPLOY_API_KEY")
	}()

	p := &DokployProvider{version: "1.0.0"}

	// Create config with null values (to use env vars)
	configType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"host":    tftypes.String,
			"api_key": tftypes.String,
		},
	}

	configValue := tftypes.NewValue(configType, map[string]tftypes.Value{
		"host":    tftypes.NewValue(tftypes.String, nil),
		"api_key": tftypes.NewValue(tftypes.String, nil),
	})

	schemaReq := provider.SchemaRequest{}
	schemaResp := &provider.SchemaResponse{}
	p.Schema(context.Background(), schemaReq, schemaResp)

	config := tfsdk.Config{
		Schema: schemaResp.Schema,
		Raw:    configValue,
	}

	req := provider.ConfigureRequest{
		Config: config,
	}
	resp := &provider.ConfigureResponse{}

	p.Configure(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error: %v", resp.Diagnostics)
	}

	if resp.DataSourceData == nil {
		t.Error("expected DataSourceData to be set from env vars")
	}

	if resp.ResourceData == nil {
		t.Error("expected ResourceData to be set from env vars")
	}
}

func TestConfigure_MissingHost(t *testing.T) {
	// Clear environment variables
	os.Unsetenv("DOKPLOY_HOST")
	os.Unsetenv("DOKPLOY_API_KEY")

	p := &DokployProvider{version: "1.0.0"}

	// Create config with null values (no host provided)
	configType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"host":    tftypes.String,
			"api_key": tftypes.String,
		},
	}

	configValue := tftypes.NewValue(configType, map[string]tftypes.Value{
		"host":    tftypes.NewValue(tftypes.String, nil),
		"api_key": tftypes.NewValue(tftypes.String, nil),
	})

	schemaReq := provider.SchemaRequest{}
	schemaResp := &provider.SchemaResponse{}
	p.Schema(context.Background(), schemaReq, schemaResp)

	config := tfsdk.Config{
		Schema: schemaResp.Schema,
		Raw:    configValue,
	}

	req := provider.ConfigureRequest{
		Config: config,
	}
	resp := &provider.ConfigureResponse{}

	p.Configure(context.Background(), req, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error for missing host")
	}

	// Check that the error message is about missing host
	foundHostError := false
	for _, diag := range resp.Diagnostics {
		if diag.Summary() == "Missing Host Configuration" {
			foundHostError = true
			break
		}
	}

	if !foundHostError {
		t.Error("expected 'Missing Host Configuration' error")
	}
}

func TestConfigure_OptionalAPIKey(t *testing.T) {
	// Clear environment variables
	os.Unsetenv("DOKPLOY_HOST")
	os.Unsetenv("DOKPLOY_API_KEY")

	p := &DokployProvider{version: "1.0.0"}

	// Create config with host but no api_key
	configType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"host":    tftypes.String,
			"api_key": tftypes.String,
		},
	}

	configValue := tftypes.NewValue(configType, map[string]tftypes.Value{
		"host":    tftypes.NewValue(tftypes.String, "https://dokploy.example.com"),
		"api_key": tftypes.NewValue(tftypes.String, nil),
	})

	schemaReq := provider.SchemaRequest{}
	schemaResp := &provider.SchemaResponse{}
	p.Schema(context.Background(), schemaReq, schemaResp)

	config := tfsdk.Config{
		Schema: schemaResp.Schema,
		Raw:    configValue,
	}

	req := provider.ConfigureRequest{
		Config: config,
	}
	resp := &provider.ConfigureResponse{}

	p.Configure(context.Background(), req, resp)

	// Should not error - API key is optional for bootstrap resource
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error: %v", resp.Diagnostics)
	}

	if resp.DataSourceData == nil {
		t.Error("expected DataSourceData to be set even without API key")
	}
}

func TestResources(t *testing.T) {
	p := &DokployProvider{version: "1.0.0"}

	resources := p.Resources(context.Background())

	// Provider should return 29 resources
	expectedCount := 29
	if len(resources) != expectedCount {
		t.Errorf("expected %d resources, got %d", expectedCount, len(resources))
	}

	// Verify all resource factories work
	for i, factory := range resources {
		if factory == nil {
			t.Errorf("resource factory at index %d is nil", i)
			continue
		}

		r := factory()
		if r == nil {
			t.Errorf("resource at index %d returned nil", i)
		}
	}
}

func TestDataSources(t *testing.T) {
	p := &DokployProvider{version: "1.0.0"}

	dataSources := p.DataSources(context.Background())

	// Provider should return 20 data sources
	expectedCount := 20
	if len(dataSources) != expectedCount {
		t.Errorf("expected %d data sources, got %d", expectedCount, len(dataSources))
	}

	// Verify all data source factories work
	for i, factory := range dataSources {
		if factory == nil {
			t.Errorf("data source factory at index %d is nil", i)
			continue
		}

		ds := factory()
		if ds == nil {
			t.Errorf("data source at index %d returned nil", i)
		}
	}
}

func TestProviderImplementsInterface(t *testing.T) {
	var _ provider.Provider = &DokployProvider{}
}
