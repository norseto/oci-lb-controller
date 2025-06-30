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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// LBRegistrarSpec defines the desired state of LBRegistrar
type LBRegistrarSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	LoadBalancerId string `json:"loadBalancerId,omitempty"`

	// NodePort is the target port on the node.
	// If Service is specified, this field is ignored and the nodePort from the service is used.
	// +optional
	NodePort int `json:"nodePort,omitempty"`

	// Port is a deprecated alias for NodePort. Use NodePort instead.
	// +optional
	Port int `json:"port,omitempty"`

	// Service provides the information to fetch the NodePort from a Service.
	// If this is specified, the NodePort field is ignored.
	// Deprecated: Use Services field for multi-service support.
	// +optional
	Service *ServiceSpec `json:"service,omitempty"`

	// Services provides the information to fetch NodePorts from multiple Services.
	// If this is specified, the Service and NodePort fields are ignored.
	// +optional
	Services []ServiceSpec `json:"services,omitempty"`

	// +kubebuilder:default:=1
	Weight int `json:"weight,omitempty"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	BackendSetName string `json:"backendSetName,omitempty"`

	// +kubebuilder:validation:Required
	ApiKey ApiKeySpec `json:"apiKey,omitempty"`
}

// ServiceSpec defines the target service to get NodePort from.
type ServiceSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Namespace string `json:"namespace"`

	// Port is the port of the service.
	// It can be a port name or a port number.
	// +kubebuilder:validation:Required
	Port intstr.IntOrString `json:"port"`

	// FilterByEndpoints enables filtering nodes based on service endpoints.
	// When true, only nodes running pods for this service are registered to the load balancer.
	// +optional
	FilterByEndpoints bool `json:"filterByEndpoints,omitempty"`

	// Weight is the weight for this service's backends in the load balancer.
	// +kubebuilder:default:=1
	// +optional
	Weight int `json:"weight,omitempty"`

	// BackendSetName is the name of the backend set for this service.
	// If not specified, uses the LBRegistrarSpec.BackendSetName.
	// +optional
	BackendSetName string `json:"backendSetName,omitempty"`
}

type ApiKeySpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	User string `json:"user,omitempty"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Fingerprint string `json:"fingerprint,omitempty"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Tenancy string `json:"tenancy,omitempty"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Region string `json:"region,omitempty"`

	// +kubebuilder:validation:Required
	PrivateKey PrivateKeySpec `json:"privateKey"`
}

type PrivateKeySpec struct {
	Namespace    string                   `json:"namespace,omitempty"`
	SecretKeyRef corev1.SecretKeySelector `json:"secretKeyRef"`
}

// LBRegistrarStatus defines the observed state of LBRegistrar
type LBRegistrarStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Phase string `json:"phase,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster
//+kubebuilder:printcolumn:JSONPath=".status.phase", name=Phase, type=string

// LBRegistrar is the Schema for the lbregistrars API
type LBRegistrar struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LBRegistrarSpec   `json:"spec,omitempty"`
	Status LBRegistrarStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// LBRegistrarList contains a list of LBRegistrar
type LBRegistrarList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LBRegistrar `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LBRegistrar{}, &LBRegistrarList{})
}

const (
	// PhaseNew represents the NEW, not initialized phase.
	PhaseNew = ""
	// PhasePending represents the "PENDING" phase. The controller is reconciling load balancer.
	PhasePending = "PENDING"
	// PhaseRegistering represents the "REGISTERING" phase. The controller is registering to load balancer.
	PhaseRegistering = "REGISTERING"
	// PhaseReady represents the "READY" phase. The controller registered to load balancer.
	PhaseReady = "READY"
)
