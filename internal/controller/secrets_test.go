package controller

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestGetSecretValue(t *testing.T) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "secret", Namespace: "ns"},
		Data:       map[string][]byte{"key": []byte("value")},
	}
	c := fake.NewClientBuilder().WithObjects(secret).Build()
	ctx := context.Background()

	val, err := GetSecretValue(ctx, c, "ns", &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "secret"}, Key: "key"})
	if err != nil {
		t.Fatalf("GetSecretValue returned error: %v", err)
	}
	if val != "value" {
		t.Errorf("expected value, got %s", val)
	}
}

func TestGetSecretValueMissingKey(t *testing.T) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "secret", Namespace: "ns"},
		Data:       map[string][]byte{"other": []byte("value")},
	}
	c := fake.NewClientBuilder().WithObjects(secret).Build()
	ctx := context.Background()

	_, err := GetSecretValue(ctx, c, "ns", &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "secret"}, Key: "missing"})
	if err == nil {
		t.Fatalf("expected error when key missing")
	}
}
