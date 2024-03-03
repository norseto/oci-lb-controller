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

package controller

import (
	"context"

	api "github.com/norseto/oci-lb-controller/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// NodeHandler is a struct that implements the EventHandler interface.
type NodeHandler struct {
	client.Client
	Recorder record.EventRecorder
}

// Create is a method that handles node creation events.
// It creates a new NodeClean object if one does not already exist.
// It also registers the NodeClean object with the Kubernetes client.
// If an error occurs during the creation or registration process, it is logged.
func (nh *NodeHandler) Create(ctx context.Context, evt event.CreateEvent, _ workqueue.RateLimitingInterface) {
	object := evt.Object
	logger := log.FromContext(ctx, "node", object.GetName())
	logger.V(1).Info("node creation", "resVer", object.GetResourceVersion())

	nh.refreshToPending(ctx, object.GetName())
}

// Update is a method that handles node update events.
// It calls the checkPhaseReady method to ensure that the node's phase can be ready.
// The result of the checkPhaseReady method is ignored.
func (nh *NodeHandler) Update(ctx context.Context, evt event.UpdateEvent, _ workqueue.RateLimitingInterface) {
}

// Delete handles node deletion events.
// It deletes NodeClean object for deleted node.
func (nh *NodeHandler) Delete(ctx context.Context, evt event.DeleteEvent, _ workqueue.RateLimitingInterface) {
	logger := log.FromContext(ctx, "node", evt.Object.GetName())
	node := evt.Object
	logger.V(1).Info("node delete", "node", node.GetName(), "resver", node.GetResourceVersion())
	nh.refreshToPending(ctx, node.GetName())
}

func (nh *NodeHandler) refreshToPending(ctx context.Context, nodeName string) {
	logger := log.FromContext(ctx, "node", nodeName)
	logger.V(1).Info("Refreshing LBRegistrar")

	list := &api.LBRegistrarList{}
	if err := nh.List(ctx, list); err != nil {
		logger.Error(err, "failed to list LBRegistrar")
		return
	}
	clnt := nh.Client
	for _, lb := range list.Items {
		if lb.Status.Phase == api.PhasePending || lb.Status.Phase == api.PhaseNew {
			continue
		}
		lb.Status.Phase = api.PhasePending
		if err := clnt.Status().Update(ctx, &lb); err != nil {
			logger.Error(err, "failed to update LBRegistrar", "registrar", lb.Name)
			continue
		}
		nh.Recorder.Event(&lb, corev1.EventTypeNormal, "PhaseChange", lb.Status.Phase)
	}
}

func (nh *NodeHandler) Generic(context.Context, event.GenericEvent, workqueue.RateLimitingInterface) {
	// Do nothing
}

// getNode is a function that retrieves a corev1.Node object from the Kubernetes client based on the given name.
// It returns the found Node object and nil error if successful.
// If an error occurs during the retrieval process, it is logged and returns nil Node object and the error.
func getNode(ctx context.Context, client client.Client, name string) (*corev1.Node, error) {
	criterion := types.NamespacedName{
		Name: name,
	}
	found := &corev1.Node{}
	err := client.Get(ctx, criterion, found)
	if err != nil {
		log.FromContext(ctx).Error(err, "failed to get node", "name", name)
		return nil, err
	}
	return found, nil
}
