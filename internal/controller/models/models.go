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
