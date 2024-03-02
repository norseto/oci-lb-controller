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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	api "github.com/norseto/oci-lb-controller/api/v1alpha1"
	"github.com/norseto/oci-lb-controller/internal/controller/cloud/oci"

	"github.com/oracle/oci-go-sdk/v65/common"
)

// LBRegistrarReconciler reconciles a LBRegistrar object
type LBRegistrarReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=nodes.peppy-ratio.dev,resources=lbregistrars,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=nodes.peppy-ratio.dev,resources=lbregistrars/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=nodes.peppy-ratio.dev,resources=lbregistrars/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=create
//+kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the LBRegistrar object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.0/pkg/reconcile
func (r *LBRegistrarReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	result := ctrl.Result{}
	shouldUpdate := false

	registrar := &api.LBRegistrar{}
	if err := r.Get(ctx, req.NamespacedName, registrar); err != nil {
		logger.Error(err, "unable to fetch LBRegistrar")
		return result, client.IgnoreNotFound(err)
	}

	defer func() {
		if !shouldUpdate {
			return
		}
		if err := r.Status().Update(ctx, registrar); err != nil {
			logger.Error(err, "unable to update LBRegistrar status")
		}
	}()

	switch registrar.Status.Phase {
	case api.PhaseNew:
		logger.Info("reconciling pending registrar")
		registrar.Status.Phase = api.PhasePending
		shouldUpdate = true
	case api.PhasePending:
		logger.Info("reconciling pending registrar")
		provider, err := getConfigurationProvider(ctx, r.Client, registrar)
		if err != nil {
			logger.Error(err, "unable to create configuration provider")
			r.Recorder.Eventf(registrar, corev1.EventTypeWarning, "Failed", "unable to create configuration provider: %v", err)
			return result, err
		}
		backends, err := oci.GetBackendSet(ctx, provider, registrar.Spec)
		if err != nil {
			logger.Error(err, "unable to get backend set")
			r.Recorder.Eventf(registrar, corev1.EventTypeWarning, "Failed", "unable to get backend set: %v", err)
			return result, err
		}
		logger.Info("current backends", "backends", backends)
		registrar.Status.Phase = api.PhaseReady
		shouldUpdate = true
	}

	return result, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *LBRegistrarReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&api.LBRegistrar{}).
		Watches(&corev1.Node{}, &NodeHandler{Client: r.Client, Recorder: r.Recorder},
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{})).
		Complete(r)
}

func getConfigurationProvider(ctx context.Context, client client.Client, registrar *api.LBRegistrar) (common.ConfigurationProvider, error) {
	secSpec := registrar.Spec.ApiKey.PrivateKey
	privateKey, err := GetSecretValue(ctx, client, secSpec.Namespace, &secSpec.SecretKeyRef)
	if err != nil {
		return nil, err
	}

	provider, err := oci.NewConfigurationProvider(ctx, &registrar.Spec.ApiKey, privateKey)
	if err != nil {
		return nil, err
	}
	return provider, nil
}
