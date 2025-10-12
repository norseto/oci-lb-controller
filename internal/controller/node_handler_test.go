package controller

import (
	"context"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/event"

	api "github.com/norseto/oci-lb-controller/api/v1alpha1"
)

func TestGetNode(t *testing.T) {
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: "node1"},
	}
	c := fake.NewClientBuilder().WithObjects(node).Build()
	ctx := context.Background()
	got, err := getNode(ctx, c, "node1")
	if err != nil {
		t.Fatalf("getNode returned error: %v", err)
	}
	if got.Name != "node1" {
		t.Errorf("expected node1, got %s", got.Name)
	}
}

func TestGetNodeNotFound(t *testing.T) {
	c := fake.NewClientBuilder().Build()
	ctx := context.Background()
	_, err := getNode(ctx, c, "missing")
	if err == nil {
		t.Fatalf("expected error when node not found")
	}
}

func TestRefreshToPending(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add core scheme: %v", err)
	}
	if err := api.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add api scheme: %v", err)
	}

	reg1 := &api.LBRegistrar{
		ObjectMeta: metav1.ObjectMeta{Name: "lb1", Namespace: "default"},
		Status:     api.LBRegistrarStatus{Phase: api.PhaseReady},
	}
	reg2 := &api.LBRegistrar{
		ObjectMeta: metav1.ObjectMeta{Name: "lb2", Namespace: "default"},
		Status:     api.LBRegistrarStatus{Phase: api.PhasePending},
	}

	c := fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(reg1, reg2).
		WithStatusSubresource(&api.LBRegistrar{}).
		Build()

	recorder := record.NewFakeRecorder(10)
	handler := &NodeHandler{Client: c, Recorder: recorder}
	ctx := context.Background()

	handler.refreshToPending(ctx, "node1")

	updated1 := &api.LBRegistrar{}
	_ = c.Get(ctx, types.NamespacedName{Name: "lb1", Namespace: "default"}, updated1)
	if updated1.Status.Phase != api.PhasePending {
		t.Errorf("expected phase %s, got %s", api.PhasePending, updated1.Status.Phase)
	}

	updated2 := &api.LBRegistrar{}
	_ = c.Get(ctx, types.NamespacedName{Name: "lb2", Namespace: "default"}, updated2)
	if updated2.Status.Phase != api.PhasePending {
		t.Errorf("expected phase %s, got %s", api.PhasePending, updated2.Status.Phase)
	}

	select {
	case e := <-recorder.Events:
		if !strings.Contains(e, "PhaseChange") {
			t.Errorf("unexpected event %q", e)
		}
	default:
		t.Errorf("expected PhaseChange event")
	}
}

func TestNodeHandlerEvents(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add core scheme: %v", err)
	}
	if err := api.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add api scheme: %v", err)
	}

	handler := &NodeHandler{
		Client:   fake.NewClientBuilder().WithScheme(scheme).Build(),
		Recorder: record.NewFakeRecorder(1),
	}
	ctx := context.Background()
	node := &corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node1"}}

	handler.Create(ctx, event.TypedCreateEvent[client.Object]{Object: node}, nil)
	handler.Delete(ctx, event.TypedDeleteEvent[client.Object]{Object: node}, nil)
	handler.Update(ctx, event.TypedUpdateEvent[client.Object]{ObjectOld: node, ObjectNew: node}, nil)
	handler.Generic(ctx, event.TypedGenericEvent[client.Object]{Object: node}, nil)
}
