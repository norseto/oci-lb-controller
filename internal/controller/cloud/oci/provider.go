package oci

import (
	"context"
	"strings"

	"github.com/oracle/oci-go-sdk/v65/common"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"

	api "github.com/norseto/oci-lb-controller/api/v1alpha1"
	alb "github.com/norseto/oci-lb-controller/internal/controller/cloud/oci/loadbalancer"
	nlb "github.com/norseto/oci-lb-controller/internal/controller/cloud/oci/networkloadbalancer"
	"github.com/norseto/oci-lb-controller/internal/controller/models"
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

func isNetworkLoadBalancer(spec api.LBRegistrarSpec) bool {
	return strings.Index(spec.LoadBalancerId, ".networkloadbalancer.") >= 0
}

func GetBackendSet(ctx context.Context, provider common.ConfigurationProvider, spec api.LBRegistrarSpec) ([]*models.LoadBalanceTarget, error) {
	if isNetworkLoadBalancer(spec) {
		return nlb.GetBackendSet(ctx, provider, spec)
	}
	return alb.GetBackendSet(ctx, provider, spec)
}

func RegisterBackends(ctx context.Context, provider common.ConfigurationProvider,
	spec api.LBRegistrarSpec, targets *corev1.NodeList) error {
	if isNetworkLoadBalancer(spec) {
		return nlb.RegisterBackends(ctx, provider, spec, targets)
	}
	return alb.RegisterBackends(ctx, provider, spec, targets)
}
