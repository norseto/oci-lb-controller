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
	"strings"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	api "github.com/norseto/oci-lb-controller/api/v1alpha1"
)

func TestLBRegistrarReconciler_Reconcile(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = api.AddToScheme(scheme)

	baseRegistrar := func() *api.LBRegistrar {
		return &api.LBRegistrar{
			ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
			Spec: api.LBRegistrarSpec{
				LoadBalancerId: "lb",
				BackendSetName: "bs",
				ApiKey: api.ApiKeySpec{
					PrivateKey: api.PrivateKeySpec{
						Namespace: "default",
						SecretKeyRef: corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{Name: "secret"},
							Key:                  "key",
						},
					},
				},
			},
		}
	}

	t.Run("New to Pending", func(t *testing.T) {
		registrar := baseRegistrar()
		registrar.Status.Phase = api.PhaseNew
		c := fake.NewClientBuilder().WithScheme(scheme).
			WithObjects(registrar).
			WithStatusSubresource(&api.LBRegistrar{}).Build()
		r := &LBRegistrarReconciler{Client: c, Scheme: scheme, Recorder: record.NewFakeRecorder(10)}

		_, err := r.Reconcile(
			context.Background(),
			ctrl.Request{
				NamespacedName: types.NamespacedName{Name: "test", Namespace: "default"},
			},
		)
		if err != nil {
			t.Fatalf("Reconcile returned error: %v", err)
		}
		updated := &api.LBRegistrar{}
		_ = c.Get(context.Background(), types.NamespacedName{Name: "test", Namespace: "default"}, updated)
		if updated.Status.Phase != api.PhasePending {
			t.Fatalf("expected phase Pending, got %s", updated.Status.Phase)
		}
	})

	checkSecretError := func(t *testing.T, phase string) {
		registrar := baseRegistrar()
		registrar.Status.Phase = phase
		client := fake.NewClientBuilder().WithScheme(scheme).
			WithObjects(registrar).
			WithStatusSubresource(&api.LBRegistrar{}).Build()
		recorder := record.NewFakeRecorder(10)
		reconciler := &LBRegistrarReconciler{Client: client, Scheme: scheme, Recorder: recorder}

		_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
			NamespacedName: types.NamespacedName{Name: "test", Namespace: "default"},
		})
		if err == nil {
			t.Fatalf("expected error but got none")
		}

		updated := &api.LBRegistrar{}
		_ = client.Get(context.Background(), types.NamespacedName{Name: "test", Namespace: "default"}, updated)
		if updated.Status.Phase != api.PhasePending {
			t.Fatalf("expected phase Pending, got %s", updated.Status.Phase)
		}

		select {
		case e := <-recorder.Events:
			if !strings.Contains(e, "unable to create configuration provider") {
				t.Fatalf("unexpected event %s", e)
			}
		default:
			t.Fatalf("expected event not recorded")
		}
	}

	t.Run("Pending secret retrieval error", func(t *testing.T) {
		checkSecretError(t, api.PhasePending)
	})

	t.Run("Registering configuration provider error", func(t *testing.T) {
		checkSecretError(t, api.PhaseRegistering)
	})

	t.Run("Registering register failure requeues", func(t *testing.T) {
		registrar := baseRegistrar()
		registrar.Status.Phase = api.PhaseRegistering
		registrar.Spec.Service = &api.ServiceSpec{
			Namespace:         "default",
			Name:              "svc",
			Port:              intstr.FromInt(80),
			FilterByEndpoints: true,
		}
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "secret", Namespace: "default"},
			Data:       map[string][]byte{"key": []byte("value")},
		}
		service := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "svc", Namespace: "default"},
			Spec: corev1.ServiceSpec{
				Type:  corev1.ServiceTypeNodePort,
				Ports: []corev1.ServicePort{{Port: 80, NodePort: 30080}},
			},
		}
		c := fake.NewClientBuilder().WithScheme(scheme).
			WithObjects(registrar, secret, service).
			WithStatusSubresource(&api.LBRegistrar{}).Build()
		recorder := record.NewFakeRecorder(10)
		r := &LBRegistrarReconciler{Client: c, Scheme: scheme, Recorder: recorder}

		result, err := r.Reconcile(
			context.Background(),
			ctrl.Request{
				NamespacedName: types.NamespacedName{Name: "test", Namespace: "default"},
			},
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.RequeueAfter != 90*time.Second {
			t.Fatalf("expected RequeueAfter 90s, got %v", result.RequeueAfter)
		}
		updated := &api.LBRegistrar{}
		_ = c.Get(context.Background(), types.NamespacedName{Name: "test", Namespace: "default"}, updated)
		if updated.Status.Phase != api.PhaseRegistering {
			t.Fatalf("expected phase Registering, got %s", updated.Status.Phase)
		}
		select {
		case e := <-recorder.Events:
			if !strings.Contains(e, "unable to register backends") {
				t.Fatalf("unexpected event %s", e)
			}
		default:
			t.Fatalf("expected event not recorded")
		}
	})
}
