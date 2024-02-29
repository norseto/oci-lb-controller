package controller

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// GetSecretValue returns the value of a secret specified by the given secret key selector.
// It retrieves the secret from the provided client using the given context and namespace.
// If the secret cannot be retrieved, an error is returned.
func GetSecretValue(ctx context.Context, clnt client.Client, namespace string, sel *corev1.SecretKeySelector) (string, error) {
	logger := log.FromContext(ctx, "namespace", namespace, "name", sel.Name, "key", sel.Key)
	logger.V(2).Info("Getting secret")

	secret := corev1.Secret{}
	key := client.ObjectKey{Name: sel.Name, Namespace: namespace}
	if err := clnt.Get(ctx, key, &secret); err != nil {
		logger.Error(err, "Failed to get secret")
		return "", err
	}

	value := string(secret.Data[sel.Key])
	logger.V(2).Info("Got secret")

	return value, nil
}
