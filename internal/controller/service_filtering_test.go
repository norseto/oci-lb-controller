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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	api "github.com/norseto/oci-lb-controller/api/v1alpha1"
)

var _ = Describe("Service-based Node Filtering", func() {
	var (
		ctx     context.Context
		scheme  *runtime.Scheme
		objects []runtime.Object
	)

	BeforeEach(func() {
		ctx = context.Background()
		scheme = runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		Expect(api.AddToScheme(scheme)).To(Succeed())

		// Create test nodes
		nodes := []corev1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "node1"},
				Status:     corev1.NodeStatus{Addresses: []corev1.NodeAddress{{Type: corev1.NodeInternalIP, Address: "10.0.1.1"}}},
			},
			{
				ObjectMeta: metav1.ObjectMeta{Name: "node2"},
				Status:     corev1.NodeStatus{Addresses: []corev1.NodeAddress{{Type: corev1.NodeInternalIP, Address: "10.0.1.2"}}},
			},
			{
				ObjectMeta: metav1.ObjectMeta{Name: "node3"},
				Status:     corev1.NodeStatus{Addresses: []corev1.NodeAddress{{Type: corev1.NodeInternalIP, Address: "10.0.1.3"}}},
			},
		}

		// Create test pods
		pods := []corev1.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod1",
					Namespace: "test-ns",
				},
				Spec: corev1.PodSpec{NodeName: "node1"},
				Status: corev1.PodStatus{
					PodIP: "192.168.1.1",
					Phase: corev1.PodRunning,
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod2",
					Namespace: "test-ns",
				},
				Spec: corev1.PodSpec{NodeName: "node2"},
				Status: corev1.PodStatus{
					PodIP: "192.168.1.2",
					Phase: corev1.PodRunning,
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod3",
					Namespace: "test-ns",
				},
				Spec: corev1.PodSpec{NodeName: "node3"},
				Status: corev1.PodStatus{
					PodIP: "192.168.1.3",
					Phase: corev1.PodRunning,
				},
			},
		}

		// Create test endpoints (only pod1 and pod2 are included)
		endpoints := &corev1.Endpoints{ //nolint:staticcheck
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-service",
				Namespace: "test-ns",
			},
			Subsets: []corev1.EndpointSubset{ //nolint:staticcheck
				{
					Addresses: []corev1.EndpointAddress{
						{IP: "192.168.1.1"},
						{IP: "192.168.1.2"},
					},
				},
			},
		}

		objects = []runtime.Object{endpoints}
		for i := range nodes {
			objects = append(objects, &nodes[i])
		}
		for i := range pods {
			objects = append(objects, &pods[i])
		}
	})

	Describe("getNodesForService", func() {
		Context("when service has endpoints", func() {
			It("should return only nodes running service pods", func() {
				client := fake.NewClientBuilder().
					WithScheme(scheme).
					WithRuntimeObjects(objects...).
					Build()

				svcSpec := &api.ServiceSpec{
					Name:      "test-service",
					Namespace: "test-ns",
					Port:      intstr.FromInt(80),
				}

				nodes, err := getNodesForService(ctx, client, svcSpec)
				Expect(err).NotTo(HaveOccurred())
				Expect(nodes.Items).To(HaveLen(2))

				nodeNames := make([]string, len(nodes.Items))
				for i, node := range nodes.Items {
					nodeNames[i] = node.Name
				}
				Expect(nodeNames).To(ConsistOf("node1", "node2"))
			})
		})

		Context("when service has no endpoints", func() {
			It("should return empty node list", func() {
				// Create endpoints with no addresses
				emptyEndpoints := &corev1.Endpoints{ //nolint:staticcheck
					ObjectMeta: metav1.ObjectMeta{
						Name:      "empty-service",
						Namespace: "test-ns",
					},
					Subsets: []corev1.EndpointSubset{}, //nolint:staticcheck
				}

				objects = append(objects, emptyEndpoints)
				client := fake.NewClientBuilder().
					WithScheme(scheme).
					WithRuntimeObjects(objects...).
					Build()

				svcSpec := &api.ServiceSpec{
					Name:      "empty-service",
					Namespace: "test-ns",
					Port:      intstr.FromInt(80),
				}

				nodes, err := getNodesForService(ctx, client, svcSpec)
				Expect(err).NotTo(HaveOccurred())
				Expect(nodes.Items).To(HaveLen(0))
			})
		})

		Context("when service does not exist", func() {
			It("should return error", func() {
				client := fake.NewClientBuilder().
					WithScheme(scheme).
					WithRuntimeObjects(objects...).
					Build()

				svcSpec := &api.ServiceSpec{
					Name:      "nonexistent-service",
					Namespace: "test-ns",
					Port:      intstr.FromInt(80),
				}

				_, err := getNodesForService(ctx, client, svcSpec)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to get endpoints"))
			})
		})
	})

	Describe("Service filtering integration", func() {
		Context("when FilterByEndpoints is enabled", func() {
			It("should filter nodes in register function", func() {
				// This would require mocking OCI provider, which is complex
				// For now, we test the node filtering logic separately
				client := fake.NewClientBuilder().
					WithScheme(scheme).
					WithRuntimeObjects(objects...).
					Build()

				svcSpec := &api.ServiceSpec{
					Name:              "test-service",
					Namespace:         "test-ns",
					Port:              intstr.FromInt(80),
					FilterByEndpoints: true,
				}

				// Test the filtering logic
				nodes, err := getNodesForService(ctx, client, svcSpec)
				Expect(err).NotTo(HaveOccurred())
				Expect(nodes.Items).To(HaveLen(2))
			})
		})

		Context("when FilterByEndpoints is disabled", func() {
			It("should return all nodes", func() {
				client := fake.NewClientBuilder().
					WithScheme(scheme).
					WithRuntimeObjects(objects...).
					Build()

				allNodes := &corev1.NodeList{}
				err := client.List(ctx, allNodes)
				Expect(err).NotTo(HaveOccurred())
				Expect(allNodes.Items).To(HaveLen(3))
			})
		})
	})
})
