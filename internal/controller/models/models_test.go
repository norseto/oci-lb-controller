package models

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestGetIPAddress(t *testing.T) {
	node := &corev1.Node{}
	node.Status.Addresses = []corev1.NodeAddress{
		{Type: corev1.NodeInternalIP, Address: "10.0.0.1"},
		{Type: corev1.NodeHostName, Address: "node1"},
	}
	if ip := GetIPAddress(node); ip != "10.0.0.1" {
		t.Errorf("expected IP 10.0.0.1 got %s", ip)
	}

	node.Status.Addresses = []corev1.NodeAddress{
		{Type: corev1.NodeHostName, Address: "node1"},
	}
	if ip := GetIPAddress(node); ip != "" {
		t.Errorf("expected empty IP got %s", ip)
	}
}
