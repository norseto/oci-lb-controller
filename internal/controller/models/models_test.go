package models

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestGetIPAddressReturnsInternalIP(t *testing.T) {
	node := &corev1.Node{
		Status: corev1.NodeStatus{
			Addresses: []corev1.NodeAddress{
				{Type: corev1.NodeExternalIP, Address: "1.1.1.1"},
				{Type: corev1.NodeInternalIP, Address: "10.0.0.1"},
			},
		},
	}
	ip := GetIPAddress(node)
	if ip != "10.0.0.1" {
		t.Errorf("expected internal IP 10.0.0.1, got %s", ip)
	}
}

func TestGetIPAddressNoInternalIP(t *testing.T) {
	node := &corev1.Node{
		Status: corev1.NodeStatus{
			Addresses: []corev1.NodeAddress{
				{Type: corev1.NodeExternalIP, Address: "1.1.1.1"},
			},
		},
	}
	ip := GetIPAddress(node)
	if ip != "" {
		t.Errorf("expected empty IP, got %s", ip)
	}
}
