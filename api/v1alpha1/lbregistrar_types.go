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

	// +kubebuilder:default:=80
	NodePort int `json:"nodePort,omitempty"`

	// Port is a deprecated alias for NodePort. Use NodePort instead.
	// +optional
	Port int `json:"port,omitempty"`

	// +kubebuilder:default:=1
	Weight int `json:"weight,omitempty"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	BackendSetName string `json:"backendSetName,omitempty"`

	// +kubebuilder:validation:Required
	ApiKey ApiKeySpec `json:"apiKey,omitempty"`
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
