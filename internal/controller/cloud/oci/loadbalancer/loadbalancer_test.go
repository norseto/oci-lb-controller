package loadbalancer

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"testing"

	"github.com/oracle/oci-go-sdk/v65/common"
	ocilb "github.com/oracle/oci-go-sdk/v65/loadbalancer"
	corev1 "k8s.io/api/core/v1"

	api "github.com/norseto/oci-lb-controller/api/v1alpha1"
)

type fakeLBClient struct {
	getResp   ocilb.GetBackendSetResponse
	getErr    error
	updateReq *ocilb.UpdateBackendSetRequest
	updateErr error
}

func (f *fakeLBClient) GetBackendSet(
	ctx context.Context,
	req ocilb.GetBackendSetRequest,
) (ocilb.GetBackendSetResponse, error) {
	if f.getErr != nil {
		return ocilb.GetBackendSetResponse{}, f.getErr
	}
	return f.getResp, nil
}

func (f *fakeLBClient) UpdateBackendSet(
	ctx context.Context,
	req ocilb.UpdateBackendSetRequest,
) (ocilb.UpdateBackendSetResponse, error) {
	f.updateReq = &req
	if f.updateErr != nil {
		return ocilb.UpdateBackendSetResponse{}, f.updateErr
	}
	return ocilb.UpdateBackendSetResponse{}, nil
}

func TestGetBackendSet(t *testing.T) {
	orig := newLBClient
	defer func() { newLBClient = orig }()
	fake := &fakeLBClient{
		getResp: ocilb.GetBackendSetResponse{
			BackendSet: ocilb.BackendSet{
				Backends: []ocilb.Backend{
					{
						Name:      common.String("b1"),
						IpAddress: common.String("1.1.1.1"),
						Port:      common.Int(80),
						Weight:    common.Int(1),
					},
					{
						Name:      common.String("b2"),
						IpAddress: common.String("2.2.2.2"),
						Port:      common.Int(8080),
						Weight:    common.Int(2),
					},
				},
			},
		},
	}
	newLBClient = func(provider common.ConfigurationProvider) (LoadBalancerClient, error) { return fake, nil }
	spec := api.LBRegistrarSpec{LoadBalancerId: "lb", BackendSetName: "bs"}
	targets, err := GetBackendSet(context.Background(), nil, spec)
	if err != nil {
		t.Fatalf("GetBackendSet error: %v", err)
	}
	if len(targets) != 2 {
		t.Fatalf("expected 2 targets, got %d", len(targets))
	}
	if targets[0].Name != "b1" || targets[0].IpAddress != "1.1.1.1" || targets[0].Port != 80 || targets[0].Weight != 1 {
		t.Fatalf("unexpected target: %#v", targets[0])
	}
}

func TestRegisterBackends(t *testing.T) {
	orig := newLBClient
	defer func() { newLBClient = orig }()
	hcPort := 1024
	fake := &fakeLBClient{
		getResp: ocilb.GetBackendSetResponse{
			BackendSet: ocilb.BackendSet{
				Policy:        common.String("ROUND_ROBIN"),
				HealthChecker: &ocilb.HealthChecker{Protocol: common.String("HTTP"), Port: common.Int(hcPort)},
			},
		},
	}
	newLBClient = func(provider common.ConfigurationProvider) (LoadBalancerClient, error) { return fake, nil }
	nodes := &corev1.NodeList{Items: []corev1.Node{
		{Status: corev1.NodeStatus{Addresses: []corev1.NodeAddress{{Type: corev1.NodeInternalIP, Address: "10.0.0.1"}}}},
		{Status: corev1.NodeStatus{Addresses: []corev1.NodeAddress{{Type: corev1.NodeInternalIP, Address: "10.0.0.2"}}}},
	}}
	spec := api.LBRegistrarSpec{
		LoadBalancerId: "lb",
		BackendSetName: "bs",
		NodePort:       30000,
		Port:           80,
		Weight:         5,
	}
	if err := RegisterBackends(context.Background(), nil, spec, nodes); err != nil {
		t.Fatalf("RegisterBackends error: %v", err)
	}
	req := fake.updateReq
	if req == nil {
		t.Fatalf("UpdateBackendSet not called")
	}
	if *req.LoadBalancerId != spec.LoadBalancerId {
		t.Errorf("LoadBalancerId = %s, want %s", *req.LoadBalancerId, spec.LoadBalancerId)
	}
	if *req.BackendSetName != spec.BackendSetName {
		t.Errorf("BackendSetName = %s, want %s", *req.BackendSetName, spec.BackendSetName)
	}
	if len(req.Backends) != len(nodes.Items) {
		t.Fatalf("backends len = %d, want %d", len(req.Backends), len(nodes.Items))
	}
	for i, b := range req.Backends {
		expectedIP := nodes.Items[i].Status.Addresses[0].Address
		if *b.IpAddress != expectedIP {
			t.Errorf("backend %d ip = %s, want %s", i, *b.IpAddress, expectedIP)
		}
		if *b.Port != spec.Port {
			t.Errorf("backend %d port = %d, want %d", i, *b.Port, spec.Port)
		}
		if *b.Weight != spec.Weight {
			t.Errorf("backend %d weight = %d, want %d", i, *b.Weight, spec.Weight)
		}
	}
	if req.Policy == nil || *req.Policy != "ROUND_ROBIN" {
		t.Errorf("policy = %v, want ROUND_ROBIN", req.Policy)
	}
	if req.HealthChecker == nil ||
		*req.HealthChecker.Protocol != "HTTP" ||
		*req.HealthChecker.Port != hcPort {
		t.Errorf("unexpected health checker %+v", req.HealthChecker)
	}
}

func TestRegisterBackendsPreservesSettings(t *testing.T) {
	orig := newLBClient
	defer func() { newLBClient = orig }()
	ssl := &ocilb.SslConfiguration{
		VerifyDepth:                    common.Int(4),
		VerifyPeerCertificate:          common.Bool(true),
		HasSessionResumption:           common.Bool(true),
		TrustedCertificateAuthorityIds: []string{"ca-1", "ca-2"},
		CertificateIds:                 []string{"certificate-1"},
		CertificateName:                common.String("backend-certificate"),
		CipherSuiteName:                common.String("backend-cipher-suite"),
		Protocols:                      []string{"TLSv1.2", "TLSv1.3"},
		ServerOrderPreference:          ocilb.SslConfigurationServerOrderPreferenceEnabled,
	}
	wantSSL := &ocilb.SslConfigurationDetails{
		VerifyDepth:                    common.Int(4),
		VerifyPeerCertificate:          common.Bool(true),
		HasSessionResumption:           common.Bool(true),
		TrustedCertificateAuthorityIds: []string{"ca-1", "ca-2"},
		CertificateIds:                 []string{"certificate-1"},
		CertificateName:                common.String("backend-certificate"),
		CipherSuiteName:                common.String("backend-cipher-suite"),
		Protocols:                      []string{"TLSv1.2", "TLSv1.3"},
		ServerOrderPreference:          ocilb.SslConfigurationDetailsServerOrderPreferenceEnabled,
	}
	wantSSLJSON := map[string]any{
		"verifyDepth":                    float64(4),
		"verifyPeerCertificate":          true,
		"hasSessionResumption":           true,
		"trustedCertificateAuthorityIds": []any{"ca-1", "ca-2"},
		"certificateIds":                 []any{"certificate-1"},
		"certificateName":                "backend-certificate",
		"cipherSuiteName":                "backend-cipher-suite",
		"protocols":                      []any{"TLSv1.2", "TLSv1.3"},
		"serverOrderPreference":          "ENABLED",
	}
	applicationCookie := &ocilb.SessionPersistenceConfigurationDetails{
		CookieName:      common.String("application-session"),
		DisableFallback: common.Bool(true),
	}
	lbCookie := &ocilb.LbCookieSessionPersistenceConfigurationDetails{
		CookieName:      common.String("lb-session"),
		DisableFallback: common.Bool(true),
		Domain:          common.String("example.com"),
		Path:            common.String("/application"),
		MaxAgeInSeconds: common.Int(3600),
		IsSecure:        common.Bool(true),
		IsHttpOnly:      common.Bool(true),
	}
	tests := []struct {
		name           string
		settings       ocilb.BackendSet
		wantSSL        *ocilb.SslConfigurationDetails
		wantJSON       map[string]any
		removeBackends bool
	}{
		{
			name: "application cookie",
			settings: ocilb.BackendSet{
				SslConfiguration:                ssl,
				SessionPersistenceConfiguration: applicationCookie,
				BackendMaxConnections:           common.Int(150),
			},
			wantSSL: wantSSL,
			wantJSON: map[string]any{
				"sslConfiguration": wantSSLJSON,
				"sessionPersistenceConfiguration": map[string]any{
					"cookieName": "application-session", "disableFallback": true,
				},
				"backendMaxConnections": float64(150),
			},
		},
		{
			name: "load balancer cookie",
			settings: ocilb.BackendSet{
				SslConfiguration:                        ssl,
				LbCookieSessionPersistenceConfiguration: lbCookie,
				BackendMaxConnections:                   common.Int(300),
			},
			wantSSL: wantSSL,
			wantJSON: map[string]any{
				"sslConfiguration": wantSSLJSON,
				"lbCookieSessionPersistenceConfiguration": map[string]any{
					"cookieName": "lb-session", "disableFallback": true,
					"domain": "example.com", "path": "/application",
					"maxAgeInSeconds": float64(3600), "isSecure": true, "isHttpOnly": true,
				},
				"backendMaxConnections": float64(300),
			},
		},
		{
			name: "zero false and empty values",
			settings: ocilb.BackendSet{
				SslConfiguration: &ocilb.SslConfiguration{
					VerifyDepth:                    common.Int(0),
					VerifyPeerCertificate:          common.Bool(false),
					HasSessionResumption:           common.Bool(false),
					TrustedCertificateAuthorityIds: []string{},
					CertificateIds:                 []string{},
					CertificateName:                common.String(""),
					CipherSuiteName:                common.String(""),
					Protocols:                      []string{},
				},
				SessionPersistenceConfiguration: &ocilb.SessionPersistenceConfigurationDetails{
					CookieName: common.String("application-session"), DisableFallback: common.Bool(false),
				},
				BackendMaxConnections: common.Int(0),
			},
			wantSSL: &ocilb.SslConfigurationDetails{
				VerifyDepth:                    common.Int(0),
				VerifyPeerCertificate:          common.Bool(false),
				HasSessionResumption:           common.Bool(false),
				TrustedCertificateAuthorityIds: []string{},
				CertificateIds:                 []string{},
				CertificateName:                common.String(""),
				CipherSuiteName:                common.String(""),
				Protocols:                      []string{},
			},
			wantJSON: map[string]any{
				"sslConfiguration": map[string]any{
					"verifyDepth": float64(0), "verifyPeerCertificate": false, "hasSessionResumption": false,
					"trustedCertificateAuthorityIds": []any{}, "certificateIds": []any{},
					"certificateName": "", "cipherSuiteName": "", "protocols": []any{},
				},
				"sessionPersistenceConfiguration": map[string]any{
					"cookieName": "application-session", "disableFallback": false,
				},
				"backendMaxConnections": float64(0),
			},
		},
		{
			name: "unset SSL fields and application fallback",
			settings: ocilb.BackendSet{
				SslConfiguration: &ocilb.SslConfiguration{},
				SessionPersistenceConfiguration: &ocilb.SessionPersistenceConfigurationDetails{
					CookieName: common.String("application-session"),
				},
			},
			wantSSL: &ocilb.SslConfigurationDetails{},
			wantJSON: map[string]any{
				"sslConfiguration":                map[string]any{},
				"sessionPersistenceConfiguration": map[string]any{"cookieName": "application-session"},
			},
		},
		{
			name: "load balancer cookie false and unset fields",
			settings: ocilb.BackendSet{
				LbCookieSessionPersistenceConfiguration: &ocilb.LbCookieSessionPersistenceConfigurationDetails{
					DisableFallback: common.Bool(false), IsSecure: common.Bool(false), IsHttpOnly: common.Bool(false),
				},
			},
			wantJSON: map[string]any{
				"lbCookieSessionPersistenceConfiguration": map[string]any{
					"disableFallback": false, "isSecure": false, "isHttpOnly": false,
				},
			},
		},
		{
			name:     "absent settings",
			wantJSON: map[string]any{},
		},
		{
			name: "remove all backends",
			settings: ocilb.BackendSet{
				SslConfiguration:                ssl,
				SessionPersistenceConfiguration: applicationCookie,
				BackendMaxConnections:           common.Int(150),
			},
			wantSSL: wantSSL,
			wantJSON: map[string]any{
				"sslConfiguration": wantSSLJSON,
				"sessionPersistenceConfiguration": map[string]any{
					"cookieName": "application-session", "disableFallback": true,
				},
				"backendMaxConnections": float64(150),
			},
			removeBackends: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backendSet := tt.settings
			backendSet.Policy = common.String("LEAST_CONNECTIONS")
			backendSet.HealthChecker = &ocilb.HealthChecker{
				Protocol: common.String("HTTP"), Port: common.Int(1024), UrlPath: common.String("/ready"),
				ReturnCode: common.Int(204), Retries: common.Int(5), TimeoutInMillis: common.Int(2500),
				IntervalInMillis: common.Int(15000), ResponseBodyRegex: common.String("ready"), IsForcePlainText: common.Bool(false),
			}
			backendSet.Backends = []ocilb.Backend{{
				IpAddress: common.String("10.0.0.9"), Port: common.Int(8080), Weight: common.Int(1),
			}}
			before, err := json.Marshal(backendSet)
			if err != nil {
				t.Fatalf("marshal fetched BackendSet: %v", err)
			}
			fake := &fakeLBClient{getResp: ocilb.GetBackendSetResponse{BackendSet: backendSet}}
			newLBClient = func(common.ConfigurationProvider) (LoadBalancerClient, error) { return fake, nil }
			nodes := &corev1.NodeList{}
			wantBackends := []ocilb.BackendDetails{}
			wantBackendsJSON := []any{}
			if !tt.removeBackends {
				nodes.Items = []corev1.Node{{Status: corev1.NodeStatus{Addresses: []corev1.NodeAddress{{
					Type: corev1.NodeInternalIP, Address: "10.0.0.1",
				}}}}}
				wantBackends = append(wantBackends, ocilb.BackendDetails{
					IpAddress: common.String("10.0.0.1"), Port: common.Int(80), Weight: common.Int(5),
				})
				wantBackendsJSON = append(wantBackendsJSON, map[string]any{
					"ipAddress": "10.0.0.1", "port": float64(80), "weight": float64(5),
				})
			}
			spec := api.LBRegistrarSpec{LoadBalancerId: "lb", BackendSetName: "bs", Port: 80, NodePort: 30000, Weight: 5}
			if err := RegisterBackends(context.Background(), nil, spec, nodes); err != nil {
				t.Fatalf("RegisterBackends error: %v", err)
			}
			req := fake.updateReq
			if req == nil {
				t.Fatal("UpdateBackendSet not called")
			}
			wantChecker := &ocilb.HealthCheckerDetails{
				Protocol: common.String("HTTP"), Port: common.Int(1024), UrlPath: common.String("/ready"),
				ReturnCode: common.Int(204), Retries: common.Int(5), TimeoutInMillis: common.Int(2500),
				IntervalInMillis: common.Int(15000), ResponseBodyRegex: common.String("ready"), IsForcePlainText: common.Bool(false),
			}
			want := ocilb.UpdateBackendSetDetails{
				Policy: common.String("LEAST_CONNECTIONS"), HealthChecker: wantChecker, Backends: wantBackends,
				SslConfiguration: tt.wantSSL, BackendMaxConnections: tt.settings.BackendMaxConnections,
				SessionPersistenceConfiguration:         tt.settings.SessionPersistenceConfiguration,
				LbCookieSessionPersistenceConfiguration: tt.settings.LbCookieSessionPersistenceConfiguration,
			}
			if !reflect.DeepEqual(req.UpdateBackendSetDetails, want) {
				t.Errorf("update details = %+v, want %+v", req.UpdateBackendSetDetails, want)
			}
			httpReq, err := req.HTTPRequest(http.MethodPut, "/loadBalancers/{loadBalancerId}/backendSets/{backendSetName}", nil, nil)
			if err != nil {
				t.Fatalf("HTTPRequest error: %v", err)
			}
			defer func() {
				if err := httpReq.Body.Close(); err != nil {
					t.Errorf("close request body: %v", err)
				}
			}()
			if httpReq.URL.Path != "/loadBalancers/lb/backendSets/bs" {
				t.Errorf("request path = %q", httpReq.URL.Path)
			}
			var body map[string]any
			if err := json.NewDecoder(httpReq.Body).Decode(&body); err != nil {
				t.Fatalf("decode request body: %v", err)
			}
			wantBody := map[string]any{
				"policy": "LEAST_CONNECTIONS", "backends": wantBackendsJSON,
				"healthChecker": map[string]any{
					"protocol": "HTTP", "port": float64(1024), "urlPath": "/ready", "returnCode": float64(204),
					"retries": float64(5), "timeoutInMillis": float64(2500), "intervalInMillis": float64(15000),
					"responseBodyRegex": "ready", "isForcePlainText": false,
				},
			}
			for key, value := range tt.wantJSON {
				wantBody[key] = value
			}
			if !reflect.DeepEqual(body, wantBody) {
				t.Errorf("request body = %#v, want %#v", body, wantBody)
			}
			after, err := json.Marshal(fake.getResp.BackendSet)
			if err != nil {
				t.Fatalf("marshal fetched BackendSet after registration: %v", err)
			}
			if string(after) != string(before) {
				t.Error("RegisterBackends mutated the fetched BackendSet")
			}
		})
	}
}

func TestRegisterBackendsGetError(t *testing.T) {
	orig := newLBClient
	defer func() { newLBClient = orig }()
	getErr := errors.New("get")
	fake := &fakeLBClient{getErr: getErr}
	newLBClient = func(common.ConfigurationProvider) (LoadBalancerClient, error) { return fake, nil }
	err := RegisterBackends(context.Background(), nil, api.LBRegistrarSpec{}, &corev1.NodeList{})
	if !errors.Is(err, getErr) {
		t.Fatalf("RegisterBackends error = %v, want %v", err, getErr)
	}
	if fake.updateReq != nil {
		t.Error("UpdateBackendSet called after GetBackendSet failed")
	}
}

func TestGetBackendSetClientError(t *testing.T) {
	orig := newLBClient
	defer func() { newLBClient = orig }()
	newLBClient = func(common.ConfigurationProvider) (LoadBalancerClient, error) {
		return nil, errors.New("client")
	}
	if _, err := GetBackendSet(context.Background(), nil, api.LBRegistrarSpec{}); err == nil {
		t.Fatalf("expected error when client creation fails")
	}
}

func TestCurrentBackendSetError(t *testing.T) {
	orig := newLBClient
	defer func() { newLBClient = orig }()
	fake := &fakeLBClient{getErr: errors.New("get")}
	newLBClient = func(common.ConfigurationProvider) (LoadBalancerClient, error) { return fake, nil }
	if _, err := GetBackendSet(context.Background(), nil, api.LBRegistrarSpec{}); err == nil {
		t.Fatalf("expected error when GetBackendSet fails")
	}
}

func TestRegisterBackendsUpdateError(t *testing.T) {
	orig := newLBClient
	defer func() { newLBClient = orig }()
	updateErr := errors.New("update")
	fake := &fakeLBClient{
		getResp: ocilb.GetBackendSetResponse{
			BackendSet: ocilb.BackendSet{
				Policy:        common.String("ROUND_ROBIN"),
				HealthChecker: &ocilb.HealthChecker{},
			},
		},
		updateErr: updateErr,
	}
	newLBClient = func(common.ConfigurationProvider) (LoadBalancerClient, error) { return fake, nil }
	err := RegisterBackends(
		context.Background(),
		nil,
		api.LBRegistrarSpec{
			BackendSetName: "bs",
			LoadBalancerId: "lb",
			Weight:         1,
			NodePort:       42,
		},
		&corev1.NodeList{},
	)
	if !errors.Is(err, updateErr) {
		t.Fatalf("RegisterBackends error = %v, want %v", err, updateErr)
	}
}

func TestRegisterBackendsUsesNodePort(t *testing.T) {
	orig := newLBClient
	defer func() { newLBClient = orig }()
	fake := &fakeLBClient{
		getResp: ocilb.GetBackendSetResponse{
			BackendSet: ocilb.BackendSet{
				Policy:        common.String("ROUND_ROBIN"),
				HealthChecker: &ocilb.HealthChecker{},
			},
		},
	}
	newLBClient = func(common.ConfigurationProvider) (LoadBalancerClient, error) { return fake, nil }
	nodes := &corev1.NodeList{
		Items: []corev1.Node{
			{
				Status: corev1.NodeStatus{
					Addresses: []corev1.NodeAddress{{
						Type:    corev1.NodeInternalIP,
						Address: "10.0.0.1",
					}},
				},
			},
		},
	}
	spec := api.LBRegistrarSpec{LoadBalancerId: "lb", BackendSetName: "bs", NodePort: 31234, Weight: 3}
	if err := RegisterBackends(context.Background(), nil, spec, nodes); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fake.updateReq == nil || *fake.updateReq.UpdateBackendSetDetails.Backends[0].Port != spec.NodePort {
		t.Fatalf("expected node port to be used")
	}
}
