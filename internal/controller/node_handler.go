/*
GNU GENERAL PUBLIC LICENSE
Version 3, 29 June 2007

Copyright (c) 2024-25 Norihiro Seto

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in hope that it will be useful,
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

	corev1 "k8s.io/api/core/v1"

	api "github.com/norseto/oci-lb-controller/api/v1alpha1"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// NodeHandler is a struct that implements the TypedEventHandler interface.
type NodeHandler struct {
	client.Client
	Recorder record.EventRecorder
}

// Create handles node creation events.
// Each event triggers refreshToPending to update LBRegistrar resources.
func (nh *NodeHandler) Create(ctx context.Context, evt event.TypedCreateEvent[client.Object], _ workqueue.TypedRateLimitingInterface[reconcile.Request]) {
	object := evt.Object
	logger := log.FromContext(ctx, "node", object.GetName())
	logger.V(1).Info("node creation", "resVer", object.GetResourceVersion())

	nh.refreshToPending(ctx, object.GetName())
}

// Update node itsself does not cause any changes to LBRegistrar resources.
func (nh *NodeHandler) Update(ctx context.Context, evt event.TypedUpdateEvent[client.Object], _ workqueue.TypedRateLimitingInterface[reconcile.Request]) {
	// Do nothing
}

// Delete handles node deletion events.
// Deletion also triggers refreshToPending so that LBRegistrar resources
// are updated when a node disappears.
func (nh *NodeHandler) Delete(ctx context.Context, evt event.TypedDeleteEvent[client.Object], _ workqueue.TypedRateLimitingInterface[reconcile.Request]) {
	logger := log.FromContext(ctx, "node", evt.Object.GetName())
	node := evt.Object
	logger.V(1).Info("node delete", "node", node.GetName(), "resver", node.GetResourceVersion())
	nh.refreshToPending(ctx, node.GetName())
}

func (nh *NodeHandler) Generic(ctx context.Context, evt event.TypedGenericEvent[client.Object], _ workqueue.TypedRateLimitingInterface[reconcile.Request]) {
	// Do nothing
}

// refreshToPending refreshes the LBRegistrar objects to Pending phase.
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
