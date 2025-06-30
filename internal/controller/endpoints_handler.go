/*
GNU GENERAL PUBLIC LICENSE
Version 3, 29 June 2007

Copyright (c) 2024-25 Norihiro Seto

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
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
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	api "github.com/norseto/oci-lb-controller/api/v1alpha1"
)

// EndpointsHandler handles Endpoints resource events for service-based node filtering.
type EndpointsHandler struct {
	client.Client
	Recorder record.EventRecorder
}

// Create handles endpoints creation events.
func (eh *EndpointsHandler) Create(ctx context.Context, evt event.TypedCreateEvent[client.Object], _ workqueue.TypedRateLimitingInterface[reconcile.Request]) {
	logger := log.FromContext(ctx, "endpoints", evt.Object.GetName(), "namespace", evt.Object.GetNamespace())
	logger.V(1).Info("endpoints creation")
	eh.handleEndpointsChange(ctx, evt.Object)
}

// Update handles endpoints update events.
func (eh *EndpointsHandler) Update(ctx context.Context, evt event.TypedUpdateEvent[client.Object], _ workqueue.TypedRateLimitingInterface[reconcile.Request]) {
	logger := log.FromContext(ctx, "endpoints", evt.ObjectNew.GetName(), "namespace", evt.ObjectNew.GetNamespace())
	logger.V(1).Info("endpoints update")
	eh.handleEndpointsChange(ctx, evt.ObjectNew)
}

// Delete handles endpoints deletion events.
func (eh *EndpointsHandler) Delete(ctx context.Context, evt event.TypedDeleteEvent[client.Object], _ workqueue.TypedRateLimitingInterface[reconcile.Request]) {
	logger := log.FromContext(ctx, "endpoints", evt.Object.GetName(), "namespace", evt.Object.GetNamespace())
	logger.V(1).Info("endpoints deletion")
	eh.handleEndpointsChange(ctx, evt.Object)
}

func (eh *EndpointsHandler) Generic(ctx context.Context, evt event.TypedGenericEvent[client.Object], _ workqueue.TypedRateLimitingInterface[reconcile.Request]) {
	// Do nothing
}

// handleEndpointsChange processes endpoints changes and updates affected LBRegistrar resources.
func (eh *EndpointsHandler) handleEndpointsChange(ctx context.Context, obj client.Object) {
	logger := log.FromContext(ctx, "endpoints", obj.GetName(), "namespace", obj.GetNamespace())

	// Find LBRegistrar resources that reference this service
	list := &api.LBRegistrarList{}
	if err := eh.List(ctx, list); err != nil {
		logger.Error(err, "failed to list LBRegistrar resources")
		return
	}

	affectedCount := 0
	for _, lb := range list.Items {
		shouldUpdate := false

		// Check single service configuration (backward compatibility)
		if lb.Spec.Service != nil &&
			lb.Spec.Service.FilterByEndpoints &&
			lb.Spec.Service.Namespace == obj.GetNamespace() &&
			lb.Spec.Service.Name == obj.GetName() {
			shouldUpdate = true
		}

		// Check multi-service configuration
		if len(lb.Spec.Services) > 0 {
			for _, service := range lb.Spec.Services {
				if service.FilterByEndpoints &&
					service.Namespace == obj.GetNamespace() &&
					service.Name == obj.GetName() {
					shouldUpdate = true
					break
				}
			}
		}

		if shouldUpdate {
			// Update status to trigger reconciliation
			if lb.Status.Phase != api.PhasePending && lb.Status.Phase != api.PhaseNew {
				lb.Status.Phase = api.PhasePending
				if err := eh.Status().Update(ctx, &lb); err != nil {
					logger.Error(err, "failed to update LBRegistrar status", "registrar", lb.Name)
					continue
				}
				eh.Recorder.Event(&lb, corev1.EventTypeNormal, "EndpointsChanged",
					fmt.Sprintf("Service %s/%s endpoints changed, triggering reconciliation",
						obj.GetNamespace(), obj.GetName()))
				affectedCount++
			}
		}
	}

	if affectedCount > 0 {
		logger.Info("triggered reconciliation for affected LBRegistrar resources", "count", affectedCount)
	}
}
