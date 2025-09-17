package networkloadbalancer

import (
	"context"
	"testing"
	"time"

	"github.com/oracle/oci-go-sdk/v65/common"
	ocilb "github.com/oracle/oci-go-sdk/v65/networkloadbalancer"
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
	defer func() { newNLBClient = orig }()
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
