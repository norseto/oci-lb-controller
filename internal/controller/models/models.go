package models

import corev1 "k8s.io/api/core/v1"

// LoadBalanceTarget represents a target for load balancing.
type LoadBalanceTarget struct {
	Name      string
	IpAddress string
	Port      int
	Weight    int
}

// GetIPAddress is a function that retrieves the internal IP address of a corev1.Node object.
// It iterates through the addresses in the node's status and returns the first address
// of type corev1.NodeInternalIP. If no address of that type is found, it returns an empty string.
func GetIPAddress(node *corev1.Node) string {
	for _, addr := range node.Status.Addresses {
		if addr.Type == corev1.NodeInternalIP {
			return addr.Address
		}
	}
	return ""
}
