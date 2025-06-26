package controller

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestGetNode(t *testing.T) {
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: "node1"},
	}
	c := fake.NewClientBuilder().WithObjects(node).Build()
	ctx := context.Background()
	got, err := getNode(ctx, c, "node1")
	if err != nil {
		t.Fatalf("getNode returned error: %v", err)
	}
	if got.Name != "node1" {
		t.Errorf("expected node1, got %s", got.Name)
	}
}

func TestGetNodeNotFound(t *testing.T) {
	c := fake.NewClientBuilder().Build()
	ctx := context.Background()
	_, err := getNode(ctx, c, "missing")
	if err == nil {
		t.Fatalf("expected error when node not found")
	}
}
