package controller

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	api "github.com/norseto/oci-lb-controller/api/v1alpha1"
)

var _ = Describe("EndpointsHandler multi-service", func() {
	It("updates only LBRegistrar matching modified Endpoints", func() {
		ctx := context.Background()
		scheme := runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		Expect(api.AddToScheme(scheme)).To(Succeed())

		recorder := record.NewFakeRecorder(10)

		testEndpoints := &corev1.Endpoints{ //nolint:staticcheck
			ObjectMeta: metav1.ObjectMeta{
				Name:      "svc-target",
				Namespace: "test-ns",
			},
		}

		registrarTarget := &api.LBRegistrar{
			ObjectMeta: metav1.ObjectMeta{
				Name: "registrar-target",
			},
			Spec: api.LBRegistrarSpec{
				LoadBalancerId: "lb1",
				BackendSetName: "backend1",
				Services: []api.ServiceSpec{{
					Name:              "svc-target",
					Namespace:         "test-ns",
					Port:              intstr.FromInt(80),
					FilterByEndpoints: true,
				}},
			},
			Status: api.LBRegistrarStatus{
				Phase: api.PhaseReady,
			},
		}

		registrarOther := &api.LBRegistrar{
			ObjectMeta: metav1.ObjectMeta{
				Name: "registrar-other",
			},
			Spec: api.LBRegistrarSpec{
				LoadBalancerId: "lb2",
				BackendSetName: "backend2",
				Services: []api.ServiceSpec{{
					Name:              "svc-other",
					Namespace:         "test-ns",
					Port:              intstr.FromInt(80),
					FilterByEndpoints: true,
				}},
			},
			Status: api.LBRegistrarStatus{
				Phase: api.PhaseReady,
			},
		}

		registrarNoFilter := &api.LBRegistrar{
			ObjectMeta: metav1.ObjectMeta{
				Name: "registrar-nofilter",
			},
			Spec: api.LBRegistrarSpec{
				LoadBalancerId: "lb3",
				BackendSetName: "backend3",
				Services: []api.ServiceSpec{{
					Name:              "svc-target",
					Namespace:         "test-ns",
					Port:              intstr.FromInt(80),
					FilterByEndpoints: false,
				}},
			},
			Status: api.LBRegistrarStatus{
				Phase: api.PhaseReady,
			},
		}

		client := fake.NewClientBuilder().
			WithScheme(scheme).
			WithRuntimeObjects(registrarTarget, registrarOther, registrarNoFilter).
			WithStatusSubresource(&api.LBRegistrar{}).
			Build()

		handler := &EndpointsHandler{
			Client:   client,
			Recorder: recorder,
		}

		handler.handleEndpointsChange(ctx, testEndpoints)

		updatedTarget := &api.LBRegistrar{}
		keyTarget := types.NamespacedName{Name: registrarTarget.Name, Namespace: registrarTarget.Namespace}
		Expect(client.Get(ctx, keyTarget, updatedTarget)).To(Succeed())
		Expect(updatedTarget.Status.Phase).To(Equal(api.PhasePending))

		updatedOther := &api.LBRegistrar{}
		keyOther := types.NamespacedName{Name: registrarOther.Name, Namespace: registrarOther.Namespace}
		Expect(client.Get(ctx, keyOther, updatedOther)).To(Succeed())
		Expect(updatedOther.Status.Phase).To(Equal(api.PhaseReady))

		updatedNoFilter := &api.LBRegistrar{}
		keyNoFilter := types.NamespacedName{Name: registrarNoFilter.Name, Namespace: registrarNoFilter.Namespace}
		Expect(client.Get(ctx, keyNoFilter, updatedNoFilter)).To(Succeed())
		Expect(updatedNoFilter.Status.Phase).To(Equal(api.PhaseReady))
	})
})
