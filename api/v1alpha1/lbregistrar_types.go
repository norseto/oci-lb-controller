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
