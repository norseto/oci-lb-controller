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
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/oracle/oci-go-sdk/v65/common"

	api "github.com/norseto/oci-lb-controller/api/v1alpha1"
	"github.com/norseto/oci-lb-controller/internal/controller/cloud/oci"
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
	result := ctrl.Result{}
	shouldUpdate := false

	registrar := &api.LBRegistrar{}
	if err := r.Get(ctx, req.NamespacedName, registrar); err != nil {
		log.FromContext(ctx).Error(err, "unable to fetch LBRegistrar")
		return result, client.IgnoreNotFound(err)
	}

	logger := log.FromContext(ctx,
		"lbId", registrar.Spec.LoadBalancerId,
		"backeneset", registrar.Spec.BackendSetName)
	ctx = log.IntoContext(ctx, logger)

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
		logger.Info("got current backends", "backends", backends)
		registrar.Status.Phase = api.PhaseRegistering
		shouldUpdate = true
	case api.PhaseRegistering:
		logger.Info("reconciling registering registrar")
		confErr, regErr := register(ctx, r.Client, registrar)
		if confErr != nil {
			logger.Error(confErr, "unable to create configuration provider")
			r.Recorder.Eventf(registrar, corev1.EventTypeWarning, "Failed", "unable to create configuration provider: %v", confErr)
			registrar.Status.Phase = api.PhasePending
			shouldUpdate = true
			return result, confErr
		} else if regErr != nil {
			logger.Error(regErr, "unable to register backends")
			r.Recorder.Eventf(registrar, corev1.EventTypeWarning, "Failed", "unable to register backends: %v", regErr)
			result.RequeueAfter = 90 * time.Second
			return result, regErr
		}
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

func register(ctx context.Context, clnt client.Client, registrar *api.LBRegistrar) (configErr error, regErr error) {
	logger := log.FromContext(ctx)

	provider, configErr := getConfigurationProvider(ctx, clnt, registrar)
	if configErr != nil {
		return
	}

	nodes := &corev1.NodeList{}
	configErr = clnt.List(ctx, nodes)
	if configErr != nil {
		configErr = client.IgnoreNotFound(configErr)
		return
	}

	logger.V(1).Info("found node", "count", len(nodes.Items))
	logger.V(2).Info("found nodes", "nodes", nodes.Items)
	regErr = clnt.List(ctx, nodes)
	if regErr != nil {
		regErr = client.IgnoreNotFound(regErr)
		return
	}

	regErr = oci.RegisterBackends(ctx, provider, registrar.Spec, nodes)
	return
}
