package controller

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/oracle/oci-go-sdk/v65/common"

	api "github.com/norseto/oci-lb-controller/api/v1alpha1"
)

func setupScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add core scheme: %v", err)
	}
	if err := api.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add api scheme: %v", err)
	}
	return scheme
}

func TestRegisterAllNodes(t *testing.T) {
	scheme := setupScheme(t)

	nodes := []runtime.Object{
		&corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node1"}},
		&corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node2"}},
	}

	registrar := &api.LBRegistrar{
		ObjectMeta: metav1.ObjectMeta{Name: "sample"},
		Spec: api.LBRegistrarSpec{
			LoadBalancerId: "lb",
			BackendSetName: "backend",
			ApiKey:         api.ApiKeySpec{User: "user", Fingerprint: "fp", Tenancy: "ten", Region: "reg", PrivateKey: api.PrivateKeySpec{}},
		},
	}

	builder := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(append(nodes, registrar)...)
	clnt := builder.Build()

	originalProvider := getConfigurationProviderFunc
	originalRegister := registerBackendsFunc
	defer func() {
		getConfigurationProviderFunc = originalProvider
		registerBackendsFunc = originalRegister
	}()

	getConfigurationProviderFunc = func(context.Context, client.Client, *api.LBRegistrar) (common.ConfigurationProvider, error) {
		return nil, nil
	}

	var receivedNodes int
	registerBackendsFunc = func(_ context.Context, _ common.ConfigurationProvider, spec api.LBRegistrarSpec, nodes *corev1.NodeList) error {
		receivedNodes = len(nodes.Items)
		if spec.BackendSetName != "backend" {
			t.Fatalf("unexpected backend set %s", spec.BackendSetName)
		}
		return nil
	}

	configErr, regErr := register(context.Background(), clnt, registrar)
	if configErr != nil || regErr != nil {
		t.Fatalf("register returned errors %v %v", configErr, regErr)
	}
	if receivedNodes != 2 {
		t.Fatalf("expected 2 nodes, got %d", receivedNodes)
	}
}

func TestRegisterMultipleServices(t *testing.T) {
	scheme := setupScheme(t)

	node1 := &corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node1"}}
	node2 := &corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node2"}}

	svc1 := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "svc1", Namespace: "default"},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeNodePort,
			Ports: []corev1.ServicePort{{
				Port:     80,
				NodePort: 30080,
			}},
		},
	}

	svc2 := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "svc2", Namespace: "default"},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeNodePort,
			Ports: []corev1.ServicePort{{
				Name:     "http",
				Port:     8080,
				NodePort: 31080,
			}},
		},
	}

	endpoints := &corev1.Endpoints{ //nolint:staticcheck
		ObjectMeta: metav1.ObjectMeta{Name: "svc2", Namespace: "default"},
		Subsets: []corev1.EndpointSubset{{
			Addresses: []corev1.EndpointAddress{{IP: "10.0.0.2"}},
		}},
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "pod", Namespace: "default"},
		Status:     corev1.PodStatus{PodIP: "10.0.0.2"},
		Spec:       corev1.PodSpec{NodeName: "node2"},
	}

	registrar := &api.LBRegistrar{
		ObjectMeta: metav1.ObjectMeta{Name: "multi"},
		Spec: api.LBRegistrarSpec{
			LoadBalancerId: "lb",
			BackendSetName: "backend",
			Services: []api.ServiceSpec{
				{
					Name:      "svc1",
					Namespace: "default",
					Port:      intstr.FromInt(80),
				},
				{
					Name:              "svc2",
					Namespace:         "default",
					Port:              intstr.FromString("http"),
					FilterByEndpoints: true,
					Weight:            3,
					BackendSetName:    "override",
				},
			},
			ApiKey: api.ApiKeySpec{User: "user", Fingerprint: "fp", Tenancy: "ten", Region: "reg", PrivateKey: api.PrivateKeySpec{}},
		},
	}

	builder := fake.NewClientBuilder().WithScheme(scheme).
		WithRuntimeObjects(node1, node2, svc1, svc2, endpoints, pod)
	clnt := builder.Build()

	originalRegister := registerBackendsFunc
	defer func() { registerBackendsFunc = originalRegister }()

	calls := make([]struct {
		backend string
		nodes   int
		weight  int
	}, 0)

	registerBackendsFunc = func(_ context.Context, _ common.ConfigurationProvider, spec api.LBRegistrarSpec, nodes *corev1.NodeList) error {
		calls = append(calls, struct {
			backend string
			nodes   int
			weight  int
		}{spec.BackendSetName, len(nodes.Items), spec.Weight})
		return nil
	}

	var provider common.ConfigurationProvider
	if err := registerMultipleServices(context.Background(), clnt, provider, registrar); err != nil {
		t.Fatalf("registerMultipleServices error: %v", err)
	}

	if len(calls) != 2 {
		t.Fatalf("expected 2 registration calls, got %d", len(calls))
	}
	if calls[0].backend != "backend" || calls[0].nodes == 0 {
		t.Fatalf("unexpected first call %+v", calls[0])
	}
	if calls[1].backend != "override" || calls[1].nodes != 1 || calls[1].weight != 3 {
		t.Fatalf("unexpected second call %+v", calls[1])
	}
}
