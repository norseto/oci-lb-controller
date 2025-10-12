package loadbalancer

import (
	"context"
	"errors"
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

func (f *fakeLBClient) GetBackendSet(ctx context.Context, req ocilb.GetBackendSetRequest) (ocilb.GetBackendSetResponse, error) {
	if f.getErr != nil {
		return ocilb.GetBackendSetResponse{}, f.getErr
	}
	return f.getResp, nil
}

func (f *fakeLBClient) UpdateBackendSet(ctx context.Context, req ocilb.UpdateBackendSetRequest) (ocilb.UpdateBackendSetResponse, error) {
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
					{Name: common.String("b1"), IpAddress: common.String("1.1.1.1"), Port: common.Int(80), Weight: common.Int(1)},
					{Name: common.String("b2"), IpAddress: common.String("2.2.2.2"), Port: common.Int(8080), Weight: common.Int(2)},
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
	spec := api.LBRegistrarSpec{LoadBalancerId: "lb", BackendSetName: "bs", NodePort: 30000, Port: 80, Weight: 5}
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
	if len(req.UpdateBackendSetDetails.Backends) != len(nodes.Items) {
		t.Fatalf("backends len = %d, want %d", len(req.UpdateBackendSetDetails.Backends), len(nodes.Items))
	}
	for i, b := range req.UpdateBackendSetDetails.Backends {
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
	if req.UpdateBackendSetDetails.Policy == nil || *req.UpdateBackendSetDetails.Policy != "ROUND_ROBIN" {
		t.Errorf("policy = %v, want ROUND_ROBIN", req.UpdateBackendSetDetails.Policy)
	}
	if req.UpdateBackendSetDetails.HealthChecker == nil || *req.UpdateBackendSetDetails.HealthChecker.Protocol != "HTTP" || *req.UpdateBackendSetDetails.HealthChecker.Port != hcPort {
		t.Errorf("unexpected health checker %+v", req.UpdateBackendSetDetails.HealthChecker)
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
	fake := &fakeLBClient{
		getResp:   ocilb.GetBackendSetResponse{BackendSet: ocilb.BackendSet{Policy: common.String("ROUND_ROBIN"), HealthChecker: &ocilb.HealthChecker{}}},
		updateErr: errors.New("update"),
	}
	newLBClient = func(common.ConfigurationProvider) (LoadBalancerClient, error) { return fake, nil }
	err := RegisterBackends(context.Background(), nil, api.LBRegistrarSpec{BackendSetName: "bs", LoadBalancerId: "lb", Weight: 1, NodePort: 42}, &corev1.NodeList{})
	if err == nil {
		t.Fatalf("expected update error")
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
	nodes := &corev1.NodeList{Items: []corev1.Node{
		{Status: corev1.NodeStatus{Addresses: []corev1.NodeAddress{{Type: corev1.NodeInternalIP, Address: "10.0.0.1"}}}},
	}}
	spec := api.LBRegistrarSpec{LoadBalancerId: "lb", BackendSetName: "bs", NodePort: 31234, Weight: 3}
	if err := RegisterBackends(context.Background(), nil, spec, nodes); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fake.updateReq == nil || *fake.updateReq.UpdateBackendSetDetails.Backends[0].Port != spec.NodePort {
		t.Fatalf("expected node port to be used")
	}
}
