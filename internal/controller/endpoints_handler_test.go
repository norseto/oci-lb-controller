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
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/event"

	api "github.com/norseto/oci-lb-controller/api/v1alpha1"
)

var _ = Describe("EndpointsHandler", func() {
	var (
		ctx            context.Context
		scheme         *runtime.Scheme
		recorder       *record.FakeRecorder
		handler        *EndpointsHandler
		testEndpoints  *corev1.Endpoints //nolint:staticcheck
		testRegistrar1 *api.LBRegistrar
		testRegistrar2 *api.LBRegistrar
	)

	BeforeEach(func() {
		ctx = context.Background()
		scheme = runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		Expect(api.AddToScheme(scheme)).To(Succeed())

		recorder = record.NewFakeRecorder(10)

		testEndpoints = &corev1.Endpoints{ //nolint:staticcheck
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-service",
				Namespace: "test-ns",
			},
		}

		// LBRegistrar with FilterByEndpoints enabled for the test service
		testRegistrar1 = &api.LBRegistrar{
			ObjectMeta: metav1.ObjectMeta{
				Name: "registrar1",
			},
			Spec: api.LBRegistrarSpec{
				LoadBalancerId: "lb1",
				BackendSetName: "backend1",
				Service: &api.ServiceSpec{
					Name:              "test-service",
					Namespace:         "test-ns",
					Port:              intstr.FromInt(80),
					FilterByEndpoints: true,
				},
			},
			Status: api.LBRegistrarStatus{
				Phase: api.PhaseReady,
			},
		}

		// LBRegistrar with FilterByEndpoints disabled
		testRegistrar2 = &api.LBRegistrar{
			ObjectMeta: metav1.ObjectMeta{
				Name: "registrar2",
			},
			Spec: api.LBRegistrarSpec{
				LoadBalancerId: "lb2",
				BackendSetName: "backend2",
				Service: &api.ServiceSpec{
					Name:              "test-service",
					Namespace:         "test-ns",
					Port:              intstr.FromInt(80),
					FilterByEndpoints: false,
				},
			},
			Status: api.LBRegistrarStatus{
				Phase: api.PhaseReady,
			},
		}
	})

	Describe("handleEndpointsChange", func() {
		Context("when endpoints change for a service with FilterByEndpoints enabled", func() {
			It("should trigger reconciliation for affected LBRegistrar", func() {
				client := fake.NewClientBuilder().
					WithScheme(scheme).
					WithRuntimeObjects(testRegistrar1, testRegistrar2).
					WithStatusSubresource(&api.LBRegistrar{}).
					Build()

				handler = &EndpointsHandler{
					Client:   client,
					Recorder: recorder,
				}

				handler.handleEndpointsChange(ctx, testEndpoints)

				// Check that registrar1 was updated to Pending phase
				updatedRegistrar1 := &api.LBRegistrar{}
				key1 := types.NamespacedName{Name: testRegistrar1.Name, Namespace: testRegistrar1.Namespace}
				err := client.Get(ctx, key1, updatedRegistrar1)
				Expect(err).NotTo(HaveOccurred())
				Expect(updatedRegistrar1.Status.Phase).To(Equal(api.PhasePending))

				// Check that registrar2 was not affected (FilterByEndpoints is false)
				updatedRegistrar2 := &api.LBRegistrar{}
				key2 := types.NamespacedName{Name: testRegistrar2.Name, Namespace: testRegistrar2.Namespace}
				err = client.Get(ctx, key2, updatedRegistrar2)
				Expect(err).NotTo(HaveOccurred())
				Expect(updatedRegistrar2.Status.Phase).To(Equal(api.PhaseReady))

				// Check that an event was recorded
				Eventually(recorder.Events).Should(Receive(ContainSubstring("EndpointsChanged")))
			})
		})

		Context("when endpoints change for a different service", func() {
			It("should not affect any LBRegistrar", func() {
				differentEndpoints := &corev1.Endpoints{ //nolint:staticcheck
					ObjectMeta: metav1.ObjectMeta{
						Name:      "different-service",
						Namespace: "test-ns",
					},
				}

				client := fake.NewClientBuilder().
					WithScheme(scheme).
					WithRuntimeObjects(testRegistrar1, testRegistrar2).
					WithStatusSubresource(&api.LBRegistrar{}).
					Build()

				handler = &EndpointsHandler{
					Client:   client,
					Recorder: recorder,
				}

				handler.handleEndpointsChange(ctx, differentEndpoints)

				// Check that both registrars remain unchanged
				updatedRegistrar1 := &api.LBRegistrar{}
				key1 := types.NamespacedName{Name: testRegistrar1.Name, Namespace: testRegistrar1.Namespace}
				err := client.Get(ctx, key1, updatedRegistrar1)
				Expect(err).NotTo(HaveOccurred())
				Expect(updatedRegistrar1.Status.Phase).To(Equal(api.PhaseReady))

				updatedRegistrar2 := &api.LBRegistrar{}
				key2 := types.NamespacedName{Name: testRegistrar2.Name, Namespace: testRegistrar2.Namespace}
				err = client.Get(ctx, key2, updatedRegistrar2)
				Expect(err).NotTo(HaveOccurred())
				Expect(updatedRegistrar2.Status.Phase).To(Equal(api.PhaseReady))
			})
		})

		Context("when LBRegistrar is already in Pending phase", func() {
			It("should not update status again", func() {
				testRegistrar1.Status.Phase = api.PhasePending

				client := fake.NewClientBuilder().
					WithScheme(scheme).
					WithRuntimeObjects(testRegistrar1).
					WithStatusSubresource(&api.LBRegistrar{}).
					Build()

				handler = &EndpointsHandler{
					Client:   client,
					Recorder: recorder,
				}

				handler.handleEndpointsChange(ctx, testEndpoints)

				// Check that the phase remains Pending
				updatedRegistrar := &api.LBRegistrar{}
				key := types.NamespacedName{Name: testRegistrar1.Name, Namespace: testRegistrar1.Namespace}
				err := client.Get(ctx, key, updatedRegistrar)
				Expect(err).NotTo(HaveOccurred())
				Expect(updatedRegistrar.Status.Phase).To(Equal(api.PhasePending))
			})
		})
	})

	Describe("Event handlers", func() {
		var testClient client.Client

		BeforeEach(func() {
			testClient = fake.NewClientBuilder().
				WithScheme(scheme).
				WithRuntimeObjects(testRegistrar1).
				WithStatusSubresource(&api.LBRegistrar{}).
				Build()

			handler = &EndpointsHandler{
				Client:   testClient,
				Recorder: recorder,
			}
		})

		It("should handle endpoints changes directly", func() {
			// Test the handleEndpointsChange method directly
			handler.handleEndpointsChange(ctx, testEndpoints)

			// Verify the registrar was updated
			updatedRegistrar := &api.LBRegistrar{}
			key := types.NamespacedName{Name: testRegistrar1.Name, Namespace: testRegistrar1.Namespace}
			err := testClient.Get(ctx, key, updatedRegistrar)
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedRegistrar.Status.Phase).To(Equal(api.PhasePending))
		})
	})
})

func TestEndpointsHandlerEvents(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add core scheme: %v", err)
	}
	if err := api.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add api scheme: %v", err)
	}

	handler := &EndpointsHandler{
		Client:   fake.NewClientBuilder().WithScheme(scheme).Build(),
		Recorder: record.NewFakeRecorder(1),
	}

	ctx := context.Background()
	ep := &corev1.Endpoints{ObjectMeta: metav1.ObjectMeta{Name: "svc", Namespace: "default"}}

	handler.Create(ctx, event.TypedCreateEvent[client.Object]{Object: ep}, nil)
	handler.Update(ctx, event.TypedUpdateEvent[client.Object]{ObjectOld: ep, ObjectNew: ep}, nil)
	handler.Delete(ctx, event.TypedDeleteEvent[client.Object]{Object: ep}, nil)
	handler.Generic(ctx, event.TypedGenericEvent[client.Object]{Object: ep}, nil)
}
