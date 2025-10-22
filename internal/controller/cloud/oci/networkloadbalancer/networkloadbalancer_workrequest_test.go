package networkloadbalancer

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/oracle/oci-go-sdk/v65/common"
	ocilb "github.com/oracle/oci-go-sdk/v65/networkloadbalancer"
	corev1 "k8s.io/api/core/v1"

	api "github.com/norseto/oci-lb-controller/api/v1alpha1"
)

type fakeNLBClient struct {
	statuses []ocilb.OperationStatusEnum
	idx      int
}

func (f *fakeNLBClient) GetBackendSet(ctx context.Context, req ocilb.GetBackendSetRequest) (ocilb.GetBackendSetResponse, error) {
	return ocilb.GetBackendSetResponse{}, nil
}

func (f *fakeNLBClient) UpdateBackendSet(ctx context.Context, req ocilb.UpdateBackendSetRequest) (ocilb.UpdateBackendSetResponse, error) {
	return ocilb.UpdateBackendSetResponse{}, nil
}

func (f *fakeNLBClient) GetWorkRequest(ctx context.Context, req ocilb.GetWorkRequestRequest) (ocilb.GetWorkRequestResponse, error) {
	status := f.statuses[f.idx]
	if f.idx < len(f.statuses)-1 {
		f.idx++
	}
	return ocilb.GetWorkRequestResponse{WorkRequest: ocilb.WorkRequest{Status: status}}, nil
}

func TestWaitForWorkRequestCompletion(t *testing.T) {
	orig := newNLBClient
	origAttempts := workRequestMaxAttempts
	origWait := workRequestWait
	defer func() {
		newNLBClient = orig
		workRequestMaxAttempts = origAttempts
		workRequestWait = origWait
	}()
	workRequestMaxAttempts = 5
	workRequestWait = func(d time.Duration) <-chan time.Time {
		ch := make(chan time.Time)
		go func() { ch <- time.Time{} }()
		return ch
	}
	cases := []struct {
		name     string
		statuses []ocilb.OperationStatusEnum
		makeCtx  func() (context.Context, context.CancelFunc)
		wantErr  bool
	}{
		{
			name:     "Succeeded",
			statuses: []ocilb.OperationStatusEnum{ocilb.OperationStatusSucceeded},
			makeCtx:  func() (context.Context, context.CancelFunc) { return context.Background(), func() {} },
		},
		{
			name:     "Failed",
			statuses: []ocilb.OperationStatusEnum{ocilb.OperationStatusFailed},
			makeCtx:  func() (context.Context, context.CancelFunc) { return context.Background(), func() {} },
			wantErr:  true,
		},
		{
			name:     "Canceled",
			statuses: []ocilb.OperationStatusEnum{ocilb.OperationStatusCanceled},
			makeCtx:  func() (context.Context, context.CancelFunc) { return context.Background(), func() {} },
			wantErr:  true,
		},
		{
			name:     "InProgress",
			statuses: []ocilb.OperationStatusEnum{ocilb.OperationStatusInProgress, ocilb.OperationStatusSucceeded},
			makeCtx:  func() (context.Context, context.CancelFunc) { return context.Background(), func() {} },
		},
		{
			name:     "Accepted",
			statuses: []ocilb.OperationStatusEnum{ocilb.OperationStatusAccepted, ocilb.OperationStatusSucceeded},
			makeCtx:  func() (context.Context, context.CancelFunc) { return context.Background(), func() {} },
		},
		{
			name:     "Unknown",
			statuses: []ocilb.OperationStatusEnum{ocilb.OperationStatusEnum("UNKNOWN"), ocilb.OperationStatusSucceeded},
			makeCtx:  func() (context.Context, context.CancelFunc) { return context.Background(), func() {} },
		},
		{
			name:     "ContextCanceled",
			statuses: []ocilb.OperationStatusEnum{ocilb.OperationStatusInProgress},
			makeCtx: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				time.AfterFunc(100*time.Millisecond, cancel)
				return ctx, cancel
			},
			wantErr: true,
		},
		{
			name:     "ContextTimeout",
			statuses: []ocilb.OperationStatusEnum{ocilb.OperationStatusInProgress},
			makeCtx: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 100*time.Millisecond)
			},
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeNLBClient{statuses: tt.statuses}
			newNLBClient = func(provider common.ConfigurationProvider) (NetworkLoadBalancerClient, error) { return fake, nil }
			ctx, cancel := tt.makeCtx()
			defer cancel()
			client, _ := newNLBClient(nil)
			err := waitForWorkRequestCompletion(ctx, client, common.String("wr"))
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}

type fullFakeClient struct {
	getResp   ocilb.GetBackendSetResponse
	getErr    error
	updateErr error
	fakeNLBClient
}

func (f *fullFakeClient) GetBackendSet(ctx context.Context, req ocilb.GetBackendSetRequest) (ocilb.GetBackendSetResponse, error) {
	if f.getErr != nil {
		return ocilb.GetBackendSetResponse{}, f.getErr
	}
	return f.getResp, nil
}

func (f *fullFakeClient) UpdateBackendSet(ctx context.Context, req ocilb.UpdateBackendSetRequest) (ocilb.UpdateBackendSetResponse, error) {
	if f.updateErr != nil {
		return ocilb.UpdateBackendSetResponse{}, f.updateErr
	}
	return ocilb.UpdateBackendSetResponse{OpcWorkRequestId: common.String("wr")}, nil
}

func TestGetBackendSetAndRegister(t *testing.T) {
	orig := newNLBClient
	origWait := workRequestWait
	defer func() {
		newNLBClient = orig
		workRequestWait = origWait
	}()
	newNLBClient = func(common.ConfigurationProvider) (NetworkLoadBalancerClient, error) {
		return &fullFakeClient{
			getResp: ocilb.GetBackendSetResponse{BackendSet: ocilb.BackendSet{
				Policy:        ocilb.NetworkLoadBalancingPolicyFiveTuple,
				HealthChecker: &ocilb.HealthChecker{Protocol: ocilb.HealthCheckProtocolsTcp},
				Backends: []ocilb.Backend{{
					Name: common.String("b"), IpAddress: common.String("10.0.0.1"), Port: common.Int(80), Weight: common.Int(1),
				}},
			}},
			fakeNLBClient: fakeNLBClient{statuses: []ocilb.OperationStatusEnum{ocilb.OperationStatusSucceeded}},
		}, nil
	}
	workRequestWait = func(time.Duration) <-chan time.Time {
		ch := make(chan time.Time)
		go func() { ch <- time.Time{} }()
		return ch
	}
	targets, err := GetBackendSet(context.Background(), nil, api.LBRegistrarSpec{LoadBalancerId: "ocid1.networkloadbalancer.oc1..x", BackendSetName: "bs"})
	if err != nil || len(targets) != 1 {
		t.Fatalf("unexpected targets %v err=%v", targets, err)
	}

	nodes := &corev1.NodeList{Items: []corev1.Node{{Status: corev1.NodeStatus{Addresses: []corev1.NodeAddress{{Type: corev1.NodeInternalIP, Address: "10.0.0.2"}}}}}}
	if err := RegisterBackends(context.Background(), nil, api.LBRegistrarSpec{LoadBalancerId: "ocid1.networkloadbalancer.oc1..x", BackendSetName: "bs", NodePort: 30000, Weight: 1}, nodes); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

type delegatingAPI struct {
	getCalled    bool
	updateCalled bool
	workCalled   bool
}

func (d *delegatingAPI) GetBackendSet(ctx context.Context, req ocilb.GetBackendSetRequest) (ocilb.GetBackendSetResponse, error) {
	d.getCalled = true
	if req.NetworkLoadBalancerId == nil {
		return ocilb.GetBackendSetResponse{}, errors.New("missing id")
	}
	return ocilb.GetBackendSetResponse{}, nil
}

func (d *delegatingAPI) UpdateBackendSet(ctx context.Context, req ocilb.UpdateBackendSetRequest) (ocilb.UpdateBackendSetResponse, error) {
	d.updateCalled = true
	return ocilb.UpdateBackendSetResponse{}, nil
}

func (d *delegatingAPI) GetWorkRequest(ctx context.Context, req ocilb.GetWorkRequestRequest) (ocilb.GetWorkRequestResponse, error) {
	d.workCalled = true
	return ocilb.GetWorkRequestResponse{}, nil
}

func TestOCINLBClientDelegates(t *testing.T) {
	api := &delegatingAPI{}
	client := &ociNLBClient{api: api}

	if _, err := client.GetBackendSet(context.Background(), ocilb.GetBackendSetRequest{NetworkLoadBalancerId: common.String("id")}); err != nil {
		t.Fatalf("GetBackendSet error: %v", err)
	}
	if _, err := client.UpdateBackendSet(context.Background(), ocilb.UpdateBackendSetRequest{}); err != nil {
		t.Fatalf("UpdateBackendSet error: %v", err)
	}
	if _, err := client.GetWorkRequest(context.Background(), ocilb.GetWorkRequestRequest{}); err != nil {
		t.Fatalf("GetWorkRequest error: %v", err)
	}

	if !api.getCalled || !api.updateCalled || !api.workCalled {
		t.Fatalf("expected all methods invoked %+v", api)
	}
}
