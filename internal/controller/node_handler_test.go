package controller

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestGetNodeFound(t *testing.T) {
	node := &corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "n1"}}
	cl := fake.NewClientBuilder().WithObjects(node).Build()
	got, err := getNode(context.Background(), cl, "n1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || got.Name != "n1" {
		t.Fatalf("unexpected node: %#v", got)
	}
}

func TestGetNodeNotFound(t *testing.T) {
	cl := fake.NewClientBuilder().Build()
	if _, err := getNode(context.Background(), cl, "missing"); err == nil {
		t.Fatal("expected error, got nil")
	}
}
