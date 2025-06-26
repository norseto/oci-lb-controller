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

func TestGetNodePortFromServiceByNumber(t *testing.T) {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "svc", Namespace: "ns"},
		Spec: corev1.ServiceSpec{
			Type:  corev1.ServiceTypeNodePort,
			Ports: []corev1.ServicePort{{Port: 80, NodePort: 30080}},
		},
	}
	cl := fake.NewClientBuilder().WithObjects(svc).Build()
	spec := &api.ServiceSpec{Name: "svc", Namespace: "ns", Port: intstr.FromInt(80)}
	port, err := getNodePortFromService(context.Background(), cl, spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if port != 30080 {
		t.Errorf("expected 30080 got %d", port)
	}
}

func TestGetNodePortFromServiceByName(t *testing.T) {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "svc", Namespace: "ns"},
		Spec: corev1.ServiceSpec{
			Type:  corev1.ServiceTypeNodePort,
			Ports: []corev1.ServicePort{{Name: "http", NodePort: 30081}},
		},
	}
	cl := fake.NewClientBuilder().WithObjects(svc).Build()
	spec := &api.ServiceSpec{Name: "svc", Namespace: "ns", Port: intstr.FromString("http")}
	port, err := getNodePortFromService(context.Background(), cl, spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if port != 30081 {
		t.Errorf("expected 30081 got %d", port)
	}
}

func TestGetNodePortFromServiceErrors(t *testing.T) {
	svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "svc", Namespace: "ns"}}
	cl := fake.NewClientBuilder().WithObjects(svc).Build()
	spec := &api.ServiceSpec{Name: "svc", Namespace: "ns", Port: intstr.FromInt(80)}
	if _, err := getNodePortFromService(context.Background(), cl, spec); err == nil {
		t.Fatal("expected error, got nil")
	}
}
