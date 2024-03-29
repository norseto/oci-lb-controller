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

package models

import corev1 "k8s.io/api/core/v1"

// LoadBalanceTarget represents a target for load balancing.
type LoadBalanceTarget struct {
	IpAddress string
	Port      int
	Weight    int
}

// TargetGroup represents a group of targets for load balancing.
type TargetGroup struct {
	Name           string
	LoadBalancerId string
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

// MakeNodeTargets creates an array of LoadBalanceTarget objects based on the provided parameters and
// the list of corev1.Node objects.
func MakeNodeTargets(port, weight int, nodes *corev1.NodeList) []*LoadBalanceTarget {
	targets := make([]*LoadBalanceTarget, 0, len(nodes.Items))
	for _, node := range nodes.Items {
		ip := GetIPAddress(&node)
		if ip == "" {
			continue
		}
		targets = append(targets, &LoadBalanceTarget{
			IpAddress: ip,
			Port:      port,
			Weight:    weight,
		})
	}
	return targets
}
