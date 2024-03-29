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
)

func FindService(ctx context.Context, clnt client.Client, namespace, name string) (*corev1.Service, error) {
	svc := &corev1.Service{}
	err := clnt.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, svc)
	if err != nil {
		return nil, client.IgnoreNotFound(err)
	}
	return svc, nil
}

func FindServiceEndPoint(ctx context.Context, clnt client.Client, namespace, name string) (*corev1.Endpoints, error) {
	endpoint := &corev1.Endpoints{}
	err := clnt.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, endpoint)
	if err != nil {
		return nil, client.IgnoreNotFound(err)
	}
	return endpoint, nil
}
