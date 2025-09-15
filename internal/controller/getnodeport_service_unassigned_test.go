package controller

import (
	"context"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	api "github.com/norseto/oci-lb-controller/api/v1alpha1"
)

func TestGetNodePortFromServiceUnassigned(t *testing.T) {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "svc", Namespace: "default"},
		Spec: corev1.ServiceSpec{
			Type:  corev1.ServiceTypeNodePort,
			Ports: []corev1.ServicePort{{Port: 80, NodePort: 0}},
		},
	}
	c := fake.NewClientBuilder().WithObjects(svc).Build()
	ctx := context.Background()
	spec := &api.ServiceSpec{Name: "svc", Namespace: "default", Port: intstr.FromInt(80)}
	_, err := getNodePortFromService(ctx, c, spec)
	if err == nil || !strings.Contains(err.Error(), "nodePort is not allocated") {
		t.Fatalf("expected error indicating unallocated nodePort, got %v", err)
	}
}
