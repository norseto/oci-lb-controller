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

package oci

import (
	"context"

	"github.com/oracle/oci-go-sdk/v65/common"
	"sigs.k8s.io/controller-runtime/pkg/log"

	api "github.com/norseto/oci-lb-controller/api/v1alpha2"
)

// NewConfigurationProvider is a function that creates a new instance of the ConfigurationProvider interface.
// It takes in a context.Context object, a pointer to an api.ApiKeySpec object
func NewConfigurationProvider(ctx context.Context, spec *api.ApiKeySpec, privateKey string) (common.ConfigurationProvider, error) {
	_ = log.FromContext(ctx)

	key := api.ApiKeySpec{}
	var pass string
	spec.DeepCopyInto(&key)

	provider := common.NewRawConfigurationProvider(
		key.Tenancy, key.User, key.Region, key.Fingerprint, privateKey, &pass)
	return provider, nil
}
