package oci

import (
	"crypto/rsa"
	"github.com/oracle/oci-go-sdk/v65/common"
)

type AuthTokenConfigProvider struct {
	common.ConfigurationProvider
	authToken string
	region    string
	tenancy   string
}

func (a *AuthTokenConfigProvider) TenancyOCID() (string, error) {
	return a.tenancy, nil
}

func (a *AuthTokenConfigProvider) UserOCID() (string, error) {
	return "", nil
}
func (a *AuthTokenConfigProvider) KeyFingerprint() (string, error) {
	return "", nil
}
func (a *AuthTokenConfigProvider) Region() (string, error) {
	return a.region, nil
}
func (a *AuthTokenConfigProvider) PrivateRSAKey() (*rsa.PrivateKey, error) {
	return nil, nil
}
func (a *AuthTokenConfigProvider) KeyID() (string, error) {
	return "", nil
}

func (a *AuthTokenConfigProvider) AuthType() (common.AuthConfig, error) {
	return common.AuthConfig{
		AuthType:         common.InstancePrincipal,
		IsFromConfigFile: false,
		OboToken:         &a.authToken,
	}, nil
}

func NewConfigurationProvider(opts ...func(*AuthTokenConfigProvider)) common.ConfigurationProvider {
	provider := &AuthTokenConfigProvider{}
	for _, opt := range opts {
		opt(provider)
	}
	return provider
}

func Region(region string) func(*AuthTokenConfigProvider) {
	return func(provider *AuthTokenConfigProvider) {
		provider.region = region
	}
}

func AuthToken(authToken string) func(*AuthTokenConfigProvider) {
	return func(provider *AuthTokenConfigProvider) {
		provider.authToken = authToken
	}
}

func Tenancy(tenancy string) func(*AuthTokenConfigProvider) {
	return func(provider *AuthTokenConfigProvider) {
		provider.tenancy = tenancy
	}
}
