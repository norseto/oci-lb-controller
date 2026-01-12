package controller

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	api "github.com/norseto/oci-lb-controller/api/v1alpha1"
)

var _ = Describe("Service-based node filtering edge cases", func() {
	var (
		ctx    context.Context
		scheme *runtime.Scheme
		nodes  []corev1.Node
		pods   []corev1.Pod
	)

	BeforeEach(func() {
		ctx = context.Background()
		scheme = runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		Expect(api.AddToScheme(scheme)).To(Succeed())

		nodes = []corev1.Node{
			{ObjectMeta: metav1.ObjectMeta{Name: "node1"}},
			{ObjectMeta: metav1.ObjectMeta{Name: "node2"}},
		}

		pods = []corev1.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "test-ns"},
				Spec:       corev1.PodSpec{NodeName: "node1"},
				Status:     corev1.PodStatus{PodIP: "192.168.1.1", Phase: corev1.PodRunning},
			},
			{
				ObjectMeta: metav1.ObjectMeta{Name: "pod2", Namespace: "test-ns"},
				Spec:       corev1.PodSpec{NodeName: "node2"},
				Status:     corev1.PodStatus{PodIP: "192.168.1.2", Phase: corev1.PodRunning},
			},
		}
	})

	Context("duplicate pod IPs in endpoints", func() {
		It("returns unique nodes only", func() {
			endpoints := &corev1.Endpoints{ //nolint:staticcheck
				ObjectMeta: metav1.ObjectMeta{Name: "svc", Namespace: "test-ns"},
				Subsets: []corev1.EndpointSubset{ //nolint:staticcheck
					{
						Addresses: []corev1.EndpointAddress{
							{IP: "192.168.1.1"},
							{IP: "192.168.1.1"},
							{IP: "192.168.1.2"},
							{IP: "192.168.1.2"},
						},
					},
				},
			}

			objects := make([]runtime.Object, 0, 1+len(nodes)+len(pods))
			objects = append(objects, endpoints)
			for i := range nodes {
				objects = append(objects, &nodes[i])
			}
			for i := range pods {
				objects = append(objects, &pods[i])
			}

			client := fake.NewClientBuilder().
				WithScheme(scheme).
				WithRuntimeObjects(objects...).
				Build()

			svcSpec := &api.ServiceSpec{
				Name:      "svc",
				Namespace: "test-ns",
				Port:      intstr.FromInt(80),
			}

			nodeList, err := getNodesForService(ctx, client, svcSpec)
			Expect(err).NotTo(HaveOccurred())
			Expect(nodeList.Items).To(HaveLen(2))

			names := make([]string, len(nodeList.Items))
			for i, n := range nodeList.Items {
				names[i] = n.Name
			}
			Expect(names).To(ConsistOf("node1", "node2"))
		})
	})

	Context("endpoints without matching pods", func() {
		It("returns empty node list", func() {
			endpoints := &corev1.Endpoints{ //nolint:staticcheck
				ObjectMeta: metav1.ObjectMeta{Name: "no-match", Namespace: "test-ns"},
				Subsets: []corev1.EndpointSubset{ //nolint:staticcheck
					{
						Addresses: []corev1.EndpointAddress{{IP: "10.0.0.9"}},
					},
				},
			}

			objects := make([]runtime.Object, 0, 1+len(nodes)+len(pods))
			objects = append(objects, endpoints)
			for i := range nodes {
				objects = append(objects, &nodes[i])
			}
			for i := range pods {
				objects = append(objects, &pods[i])
			}

			client := fake.NewClientBuilder().
				WithScheme(scheme).
				WithRuntimeObjects(objects...).
				Build()

			svcSpec := &api.ServiceSpec{
				Name:      "no-match",
				Namespace: "test-ns",
				Port:      intstr.FromInt(80),
			}

			nodeList, err := getNodesForService(ctx, client, svcSpec)
			Expect(err).NotTo(HaveOccurred())
			Expect(nodeList.Items).To(HaveLen(0))
		})
	})
})
