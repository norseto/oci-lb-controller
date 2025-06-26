package controller

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestGetSecretValueSuccess(t *testing.T) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Data:       map[string][]byte{"key": []byte("value")},
	}
	cl := fake.NewClientBuilder().WithObjects(secret).Build()

	sel := &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "test"}, Key: "key"}
	val, err := GetSecretValue(context.Background(), cl, "default", sel)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "value" {
		t.Errorf("expected value got %s", val)
	}
}

func TestGetSecretValueNotFound(t *testing.T) {
	cl := fake.NewClientBuilder().Build()
	sel := &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "missing"}, Key: "key"}
	if _, err := GetSecretValue(context.Background(), cl, "default", sel); err == nil {
		t.Fatal("expected error, got nil")
	}
}
