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

package oci

import (
	"testing"

	api "github.com/norseto/oci-lb-controller/api/v1alpha1"
)

func TestIsNetworkLoadBalancer(t *testing.T) {
	var testCases = []struct {
		name     string
		spec     api.LBRegistrarSpec
		expected bool
	}{
		{"Network Load Balancer", api.LBRegistrarSpec{LoadBalancerId: "ocid1.networkloadbalancer.oc1.phx.exampleuniqueID"}, true},
		{"Load Balancer", api.LBRegistrarSpec{LoadBalancerId: "ocid1.loadbalancer.oc1.phx.exampleuniqueID"}, false},
		{"Empty Load Balancer", api.LBRegistrarSpec{LoadBalancerId: ""}, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := isNetworkLoadBalancer(tc.spec)
			if actual != tc.expected {
				t.Errorf("Unexpected result: expected %v, actual %v", tc.expected, actual)
			}
		})
	}
}
