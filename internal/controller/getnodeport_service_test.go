package controller

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	api "github.com/norseto/oci-lb-controller/api/v1alpha1"
)

func TestGetNodePortFromServiceNumber(t *testing.T) {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "svc", Namespace: "default"},
		Spec: corev1.ServiceSpec{
			Type:  corev1.ServiceTypeNodePort,
			Ports: []corev1.ServicePort{{Port: 80, NodePort: 30000}},
		},
	}
	c := fake.NewClientBuilder().WithObjects(svc).Build()
	ctx := context.Background()
	spec := &api.ServiceSpec{Name: "svc", Namespace: "default", Port: intstr.FromInt(80)}
	got, err := getNodePortFromService(ctx, c, spec)
	if err != nil {
		t.Fatalf("got error: %v", err)
	}
	if got != 30000 {
		t.Errorf("expected 30000, got %d", got)
	}
}

func TestGetNodePortFromServiceName(t *testing.T) {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "svc", Namespace: "default"},
		Spec: corev1.ServiceSpec{
			Type:  corev1.ServiceTypeNodePort,
			Ports: []corev1.ServicePort{{Name: "http", NodePort: 30001}},
		},
	}
	c := fake.NewClientBuilder().WithObjects(svc).Build()
	ctx := context.Background()
	spec := &api.ServiceSpec{Name: "svc", Namespace: "default", Port: intstr.FromString("http")}
	got, err := getNodePortFromService(ctx, c, spec)
	if err != nil {
		t.Fatalf("got error: %v", err)
	}
	if got != 30001 {
		t.Errorf("expected 30001, got %d", got)
	}
}

func TestGetNodePortFromServiceNotNodePort(t *testing.T) {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "svc", Namespace: "default"},
		Spec:       corev1.ServiceSpec{Type: corev1.ServiceTypeClusterIP},
	}
	c := fake.NewClientBuilder().WithObjects(svc).Build()
	ctx := context.Background()
	spec := &api.ServiceSpec{Name: "svc", Namespace: "default", Port: intstr.FromInt(80)}
	_, err := getNodePortFromService(ctx, c, spec)
	if err == nil {
		t.Fatalf("expected error for non NodePort service")
	}
}

func TestGetNodePortFromServiceMissing(t *testing.T) {
	c := fake.NewClientBuilder().Build()
	ctx := context.Background()
	spec := &api.ServiceSpec{Name: "svc", Namespace: "default", Port: intstr.FromInt(80)}
	_, err := getNodePortFromService(ctx, c, spec)
	if err == nil {
		t.Fatalf("expected error when service missing")
	}
}

func TestGetNodePortFromServiceNoMatch(t *testing.T) {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "svc", Namespace: "default"},
		Spec: corev1.ServiceSpec{
			Type:  corev1.ServiceTypeNodePort,
			Ports: []corev1.ServicePort{{Port: 81, NodePort: 30002}},
		},
	}
	c := fake.NewClientBuilder().WithObjects(svc).Build()
	ctx := context.Background()
	spec := &api.ServiceSpec{Name: "svc", Namespace: "default", Port: intstr.FromInt(80)}
	_, err := getNodePortFromService(ctx, c, spec)
	if err == nil {
		t.Fatalf("expected error for no matching port")
	}
}
