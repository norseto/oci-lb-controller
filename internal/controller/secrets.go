/*
GNU GENERAL PUBLIC LICENSE
Version 3, 29 June 2007

Copyright (c) 2024-25 Norihiro Seto

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.

For the full license text, please visit: https://www.gnu.org/licenses/gpl-3.0.txt
*/

package controller

import (
	"context"
	"fmt"

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

	valueBytes, ok := secret.Data[sel.Key]
	if !ok {
		err := fmt.Errorf("secret key %s not found", sel.Key)
		logger.Error(err, "Failed to get secret key")
		return "", err
	}

	value := string(valueBytes)
	logger.V(2).Info("Got secret")

	return value, nil
}
