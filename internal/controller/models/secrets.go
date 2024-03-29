/*
MIT License

Copyright (c) 2024 Norihiro Seto

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package models

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
