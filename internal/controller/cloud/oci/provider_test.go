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
