package acctest

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/reserve-protocol/terraform-provider-dokploy/internal/client"
)

func init() {
	resource.AddTestSweepers("dokploy_project", &resource.Sweeper{
		Name: "dokploy_project",
		F:    sweepProjects,
	})

	resource.AddTestSweepers("dokploy_registry", &resource.Sweeper{
		Name: "dokploy_registry",
		F:    sweepRegistries,
	})

	resource.AddTestSweepers("dokploy_certificate", &resource.Sweeper{
		Name: "dokploy_certificate",
		F:    sweepCertificates,
	})

	resource.AddTestSweepers("dokploy_destination", &resource.Sweeper{
		Name: "dokploy_destination",
		F:    sweepDestinations,
	})

	resource.AddTestSweepers("dokploy_sshkey", &resource.Sweeper{
		Name: "dokploy_sshkey",
		F:    sweepSSHKeys,
	})
}

func getClient() (*client.Client, error) {
	host := os.Getenv("DOKPLOY_HOST")
	apiKey := os.Getenv("DOKPLOY_API_KEY")

	if host == "" || apiKey == "" {
		return nil, nil // Skip sweep if credentials not configured
	}

	return client.New(host, apiKey), nil
}

// sweepProjects removes test projects created during acceptance tests.
func sweepProjects(_ string) error {
	c, err := getClient()
	if err != nil || c == nil {
		return err
	}

	ctx := context.Background()
	projects, err := c.GetProjects(ctx)
	if err != nil {
		return err
	}

	for _, project := range projects {
		if strings.HasPrefix(project.Name, "tf-test-") {
			log.Printf("[INFO] Sweeping project: %s (%s)", project.Name, project.ProjectID)
			if err := c.DeleteProject(ctx, project.ProjectID); err != nil {
				log.Printf("[WARN] Failed to sweep project %s: %v", project.ProjectID, err)
			}
		}
	}

	return nil
}

// sweepRegistries removes test registries created during acceptance tests.
func sweepRegistries(_ string) error {
	c, err := getClient()
	if err != nil || c == nil {
		return err
	}

	ctx := context.Background()
	registries, err := c.GetRegistries(ctx)
	if err != nil {
		return err
	}

	for _, registry := range registries {
		if strings.HasPrefix(registry.RegistryName, "tf-test-") {
			log.Printf("[INFO] Sweeping registry: %s (%s)", registry.RegistryName, registry.RegistryID)
			if err := c.DeleteRegistry(ctx, registry.RegistryID); err != nil {
				log.Printf("[WARN] Failed to sweep registry %s: %v", registry.RegistryID, err)
			}
		}
	}

	return nil
}

// sweepCertificates removes test certificates created during acceptance tests.
func sweepCertificates(_ string) error {
	c, err := getClient()
	if err != nil || c == nil {
		return err
	}

	ctx := context.Background()
	certificates, err := c.GetCertificates(ctx)
	if err != nil {
		return err
	}

	for _, certificate := range certificates {
		if strings.HasPrefix(certificate.Name, "tf-test-") {
			log.Printf("[INFO] Sweeping certificate: %s (%s)", certificate.Name, certificate.CertificateID)
			if err := c.DeleteCertificate(ctx, certificate.CertificateID); err != nil {
				log.Printf("[WARN] Failed to sweep certificate %s: %v", certificate.CertificateID, err)
			}
		}
	}

	return nil
}

// sweepDestinations removes test destinations created during acceptance tests.
func sweepDestinations(_ string) error {
	c, err := getClient()
	if err != nil || c == nil {
		return err
	}

	ctx := context.Background()
	destinations, err := c.GetDestinations(ctx)
	if err != nil {
		return err
	}

	for _, destination := range destinations {
		if strings.HasPrefix(destination.Name, "tf-test-") {
			log.Printf("[INFO] Sweeping destination: %s (%s)", destination.Name, destination.DestinationID)
			if err := c.DeleteDestination(ctx, destination.DestinationID); err != nil {
				log.Printf("[WARN] Failed to sweep destination %s: %v", destination.DestinationID, err)
			}
		}
	}

	return nil
}

// sweepSSHKeys removes test SSH keys created during acceptance tests.
func sweepSSHKeys(_ string) error {
	c, err := getClient()
	if err != nil || c == nil {
		return err
	}

	ctx := context.Background()
	keys, err := c.GetSSHKeys(ctx)
	if err != nil {
		return err
	}

	for _, key := range keys {
		if strings.HasPrefix(key.Name, "tf-test-") {
			log.Printf("[INFO] Sweeping SSH key: %s (%s)", key.Name, key.SSHKeyID)
			if err := c.DeleteSSHKey(ctx, key.SSHKeyID); err != nil {
				log.Printf("[WARN] Failed to sweep SSH key %s: %v", key.SSHKeyID, err)
			}
		}
	}

	return nil
}
