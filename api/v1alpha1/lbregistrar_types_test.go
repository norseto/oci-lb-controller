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

package v1alpha1

import (
	"testing"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestLBRegistrarSpec_DefaultValues(t *testing.T) {
	g := gomega.NewWithT(t)

	spec := LBRegistrarSpec{}

	// Test default weight value
	g.Expect(spec.Weight).To(gomega.Equal(0))

	// Test with default weight
	spec.Weight = 1
	g.Expect(spec.Weight).To(gomega.Equal(1))
}

func TestServiceSpec_DefaultValues(t *testing.T) {
	g := gomega.NewWithT(t)

	serviceSpec := ServiceSpec{
		Name:      "test-service",
		Namespace: "default",
		Port:      intstr.FromInt(8080),
	}

	// Test default weight value
	g.Expect(serviceSpec.Weight).To(gomega.Equal(0))

	// Test with default weight
	serviceSpec.Weight = 1
	g.Expect(serviceSpec.Weight).To(gomega.Equal(1))
}

func TestApiKeySpec_RequiredFields(t *testing.T) {
	g := gomega.NewWithT(t)

	apiKey := ApiKeySpec{
		User:        "test-user",
		Fingerprint: "test-fingerprint",
		Tenancy:     "test-tenancy",
		Region:      "us-ashburn-1",
		PrivateKey: PrivateKeySpec{
			Namespace: "default",
			SecretKeyRef: corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "test-secret",
				},
				Key: "private-key",
			},
		},
	}

	g.Expect(apiKey.User).To(gomega.Equal("test-user"))
	g.Expect(apiKey.Fingerprint).To(gomega.Equal("test-fingerprint"))
	g.Expect(apiKey.Tenancy).To(gomega.Equal("test-tenancy"))
	g.Expect(apiKey.Region).To(gomega.Equal("us-ashburn-1"))
	g.Expect(apiKey.PrivateKey.Namespace).To(gomega.Equal("default"))
	g.Expect(apiKey.PrivateKey.SecretKeyRef.Name).To(gomega.Equal("test-secret"))
	g.Expect(apiKey.PrivateKey.SecretKeyRef.Key).To(gomega.Equal("private-key"))
}

func TestLBRegistrarStatus_Phases(t *testing.T) {
	g := gomega.NewWithT(t)

	status := LBRegistrarStatus{}

	// Test initial phase
	g.Expect(status.Phase).To(gomega.Equal(""))

	// Test phase transitions
	status.Phase = PhasePending
	g.Expect(status.Phase).To(gomega.Equal(PhasePending))

	status.Phase = PhaseRegistering
	g.Expect(status.Phase).To(gomega.Equal(PhaseRegistering))

	status.Phase = PhaseReady
	g.Expect(status.Phase).To(gomega.Equal(PhaseReady))
}

func TestLBRegistrar_DeepCopy(t *testing.T) {
	g := gomega.NewWithT(t)

	original := &LBRegistrar{
		TypeMeta: metav1.TypeMeta{
			Kind:       "LBRegistrar",
			APIVersion: "nodes.peppy-ratio.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-lbregistrar",
			Namespace: "default",
		},
		Spec: LBRegistrarSpec{
			LoadBalancerId: "test-lb-id",
			BackendSetName: "test-backend-set",
			NodePort:       30080,
			Weight:         1,
			ApiKey: ApiKeySpec{
				User:        "test-user",
				Fingerprint: "test-fingerprint",
				Tenancy:     "test-tenancy",
				Region:      "us-ashburn-1",
				PrivateKey: PrivateKeySpec{
					Namespace: "default",
					SecretKeyRef: corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "test-secret",
						},
						Key: "private-key",
					},
				},
			},
		},
		Status: LBRegistrarStatus{
			Phase: PhaseReady,
		},
	}

	// Test DeepCopy
	copied := original.DeepCopy()
	g.Expect(copied).ToNot(gomega.BeNil())
	g.Expect(copied).ToNot(gomega.Equal(original)) // Different pointers
	g.Expect(copied.Spec).To(gomega.Equal(original.Spec))
	g.Expect(copied.Status).To(gomega.Equal(original.Status))
	g.Expect(copied.ObjectMeta).To(gomega.Equal(original.ObjectMeta))

	// Test DeepCopyInto
	var target LBRegistrar
	original.DeepCopyInto(&target)
	g.Expect(target).To(gomega.Equal(*original))
}

func TestLBRegistrarList_DeepCopy(t *testing.T) {
	g := gomega.NewWithT(t)

	original := &LBRegistrarList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "LBRegistrarList",
			APIVersion: "nodes.peppy-ratio.dev/v1alpha1",
		},
		Items: []LBRegistrar{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-1",
				},
				Spec: LBRegistrarSpec{
					LoadBalancerId: "lb-1",
					BackendSetName: "backend-1",
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-2",
				},
				Spec: LBRegistrarSpec{
					LoadBalancerId: "lb-2",
					BackendSetName: "backend-2",
				},
			},
		},
	}

	// Test DeepCopy
	copied := original.DeepCopy()
	g.Expect(copied).ToNot(gomega.BeNil())
	g.Expect(copied).ToNot(gomega.Equal(original)) // Different pointers
	g.Expect(copied.Items).To(gomega.HaveLen(2))
	g.Expect(copied.Items[0].Spec.LoadBalancerId).To(gomega.Equal("lb-1"))
	g.Expect(copied.Items[1].Spec.LoadBalancerId).To(gomega.Equal("lb-2"))

	// Test DeepCopyInto
	var target LBRegistrarList
	original.DeepCopyInto(&target)
	g.Expect(target).To(gomega.Equal(*original))
}

func TestPhaseConstants(t *testing.T) {
	g := gomega.NewWithT(t)

	g.Expect(PhaseNew).To(gomega.Equal(""))
	g.Expect(PhasePending).To(gomega.Equal("PENDING"))
	g.Expect(PhaseRegistering).To(gomega.Equal("REGISTERING"))
	g.Expect(PhaseReady).To(gomega.Equal("READY"))
}

func TestServiceSpec_PortTypes(t *testing.T) {
	g := gomega.NewWithT(t)

	// Test with integer port
	serviceSpecInt := ServiceSpec{
		Name:      "test-service",
		Namespace: "default",
		Port:      intstr.FromInt(8080),
	}
	g.Expect(serviceSpecInt.Port.IntValue()).To(gomega.Equal(8080))

	// Test with string port
	serviceSpecStr := ServiceSpec{
		Name:      "test-service",
		Namespace: "default",
		Port:      intstr.FromString("http"),
	}
	g.Expect(serviceSpecStr.Port.StrVal).To(gomega.Equal("http"))
}

func TestLBRegistrarSpec_ServiceFields(t *testing.T) {
	g := gomega.NewWithT(t)

	spec := LBRegistrarSpec{
		LoadBalancerId: "test-lb-id",
		BackendSetName: "test-backend-set",
		NodePort:       30080,
		Port:           8080, // Deprecated field
		Weight:         1,
		Service: &ServiceSpec{
			Name:      "test-service",
			Namespace: "default",
			Port:      intstr.FromInt(8080),
		},
		Services: []ServiceSpec{
			{
				Name:      "service-1",
				Namespace: "default",
				Port:      intstr.FromInt(8080),
				Weight:    1,
			},
			{
				Name:      "service-2",
				Namespace: "default",
				Port:      intstr.FromInt(8081),
				Weight:    2,
			},
		},
	}

	g.Expect(spec.LoadBalancerId).To(gomega.Equal("test-lb-id"))
	g.Expect(spec.BackendSetName).To(gomega.Equal("test-backend-set"))
	g.Expect(spec.NodePort).To(gomega.Equal(30080))
	g.Expect(spec.Port).To(gomega.Equal(8080))
	g.Expect(spec.Weight).To(gomega.Equal(1))
	g.Expect(spec.Service).ToNot(gomega.BeNil())
	g.Expect(spec.Service.Name).To(gomega.Equal("test-service"))
	g.Expect(spec.Services).To(gomega.HaveLen(2))
	g.Expect(spec.Services[0].Name).To(gomega.Equal("service-1"))
	g.Expect(spec.Services[1].Name).To(gomega.Equal("service-2"))
}
