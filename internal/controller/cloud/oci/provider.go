package oci

import (
	"context"

	"github.com/oracle/oci-go-sdk/v65/common"
	"sigs.k8s.io/controller-runtime/pkg/log"

	api "github.com/norseto/oci-lb-controller/api/v1alpha1"
)

// NewConfigurationProvider is a function that creates a new instance of the ConfigurationProvider interface.
// It takes in a context.Context object, a pointer to an api.ApiKeySpec object
func NewConfigurationProvider(ctx context.Context, spec *api.ApiKeySpec, privateKey string) (common.ConfigurationProvider, error) {
	_ = log.FromContext(ctx)

	key := api.ApiKeySpec{}
	var pass string
	spec.DeepCopyInto(&key)

	provider := common.NewRawConfigurationProvider(
		key.Tenancy, key.User, key.Region, key.Fingerprint, privateKey, &pass)
	return provider, nil
}
