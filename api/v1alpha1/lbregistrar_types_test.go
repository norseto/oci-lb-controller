package v1alpha1

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestDeepCopyFunctions(t *testing.T) {
	original := &LBRegistrar{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sample",
			Namespace: "system",
			Labels:    map[string]string{"env": "dev"},
		},
		Spec: LBRegistrarSpec{
			LoadBalancerId: "ocid1.loadbalancer.oc1..exampleuniqueID",
			NodePort:       30080,
			Port:           80,
			Service: &ServiceSpec{
				Name:              "svc",
				Namespace:         "default",
				FilterByEndpoints: true,
				Port:              intOrStringFromInt(80),
			},
			Services: []ServiceSpec{{
				Name:              "svc-2",
				Namespace:         "default",
				FilterByEndpoints: true,
				Port:              intOrStringFromString("http"),
			}},
			Weight:         5,
			BackendSetName: "backend",
			ApiKey: ApiKeySpec{
				User:        "user",
				Fingerprint: "fingerprint",
				Tenancy:     "tenancy",
				Region:      "region",
				PrivateKey: PrivateKeySpec{
					Namespace: "secrets",
					SecretKeyRef: corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: "key"},
						Key:                  "private",
					},
				},
			},
		},
		Status: LBRegistrarStatus{Phase: PhaseReady},
	}

	copied := original.DeepCopy()
	if copied == original {
		t.Fatalf("expected deep copy to allocate new structs")
	}

	copied.Labels["env"] = "prod"
	if original.Labels["env"] == "prod" {
		t.Fatalf("labels map should be copied")
	}

	copied.Spec.Service.Name = "mutated"
	if original.Spec.Service.Name == "mutated" {
		t.Fatalf("service mutation should not affect original")
	}

	if copied.Spec.Services[0].Name == "svc-2" {
		copied.Spec.Services[0].Name = "other"
		if original.Spec.Services[0].Name != "svc-2" {
			t.Fatalf("services slice should be copied")
		}
	}

	if copied.Spec.ApiKey.PrivateKey.SecretKeyRef.Key != "private" {
		t.Fatalf("unexpected private key copy")
	}

	list := &LBRegistrarList{Items: []LBRegistrar{*original}}
	if len(list.DeepCopy().Items) != 1 {
		t.Fatalf("expected DeepCopy to preserve items")
	}

	if original.Spec.ApiKey.DeepCopy() == &original.Spec.ApiKey {
		t.Fatalf("DeepCopy should return new pointer")
	}

	if original.Spec.Service.DeepCopy().Port != original.Spec.Service.Port {
		t.Fatalf("DeepCopy should preserve port")
	}

	if original.Spec.DeepCopy().Service == original.Spec.Service {
		t.Fatalf("expected service pointer to differ after copy")
	}

	if original.Spec.ApiKey.PrivateKey.DeepCopy().SecretKeyRef.Key != "private" {
		t.Fatalf("private key DeepCopy mismatch")
	}

	statusCopy := original.Status.DeepCopy()
	if statusCopy == &original.Status || statusCopy.Phase != original.Status.Phase {
		t.Fatalf("status DeepCopy failed")
	}

	if original.DeepCopyObject() == nil {
		t.Fatalf("DeepCopyObject should not return nil")
	}
	if (&LBRegistrarList{}).DeepCopyObject() == nil {
		t.Fatalf("DeepCopyObject on list should not be nil")
	}

	var nilRegistrar *LBRegistrar
	if nilRegistrar.DeepCopy() != nil {
		t.Fatalf("nil receiver should return nil copy")
	}

	var nilSpec *LBRegistrarSpec
	if nilSpec.DeepCopy() != nil {
		t.Fatalf("nil spec copy should be nil")
	}

	var nilStatus *LBRegistrarStatus
	if nilStatus.DeepCopy() != nil {
		t.Fatalf("nil status copy should be nil")
	}

	var nilService *ServiceSpec
	if nilService.DeepCopy() != nil {
		t.Fatalf("nil service copy should be nil")
	}

	var nilApiKey *ApiKeySpec
	if nilApiKey.DeepCopy() != nil {
		t.Fatalf("nil api key copy should be nil")
	}

	var nilPrivate *PrivateKeySpec
	if nilPrivate.DeepCopy() != nil {
		t.Fatalf("nil private key copy should be nil")
	}
}

func TestSchemeRegistration(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add API to scheme: %v", err)
	}

	gvks, _, err := scheme.ObjectKinds(&LBRegistrar{})
	if err != nil {
		t.Fatalf("failed to resolve GVK: %v", err)
	}
	if len(gvks) == 0 || gvks[0].Group != GroupVersion.Group {
		t.Fatalf("expected registered GVK, got %#v", gvks)
	}
}

func intOrStringFromInt(port int) intstr.IntOrString {
	return intstr.FromInt(port)
}

func intOrStringFromString(port string) intstr.IntOrString {
	return intstr.FromString(port)
}
