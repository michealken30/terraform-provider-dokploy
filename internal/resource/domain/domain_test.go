package domain

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/thefrozenfire/terraform-provider-dokploy/internal/client"
)

func domainSchema(t *testing.T) schema.Schema {
	t.Helper()
	var response resource.SchemaResponse
	(&DomainResource{}).Schema(context.Background(), resource.SchemaRequest{}, &response)
	return response.Schema
}

func TestInternalPathSchemaAcceptsAPINormalizedValueWhenOmitted(t *testing.T) {
	attribute, ok := domainSchema(t).Attributes["internal_path"].(schema.StringAttribute)
	if !ok {
		t.Fatal("internal_path is not a string attribute")
	}
	if !attribute.Optional {
		t.Error("internal_path must remain optional")
	}
	if !attribute.Computed {
		t.Error("internal_path must be computed so an omitted configuration can retain Dokploy's normalized value")
	}
	if attribute.Required {
		t.Error("internal_path must not be required")
	}
}

func TestMapDomainToStateRetainsImportedInternalPath(t *testing.T) {
	internalPath := "/"
	model := DomainResourceModel{ID: types.StringValue("domain-1")}
	(&DomainResource{}).mapDomainToState(&client.Domain{
		DomainID:     "domain-1",
		Host:         "example.com",
		Path:         "/identity",
		InternalPath: &internalPath,
	}, &model)

	if model.InternalPath.IsNull() || model.InternalPath.ValueString() != "/" {
		t.Fatalf("import/read must retain API-selected internal_path, got %#v", model.InternalPath)
	}
}

func TestExplicitInternalPathRemainsConfigurationManaged(t *testing.T) {
	configured := types.StringValue("/foo")
	model := DomainResourceModel{InternalPath: configured}

	if model.InternalPath.IsNull() || model.InternalPath.IsUnknown() {
		t.Fatal("explicit internal_path must remain known")
	}
	if model.InternalPath.ValueString() != "/foo" {
		t.Fatalf("explicit internal_path changed: got %q", model.InternalPath.ValueString())
	}
}
