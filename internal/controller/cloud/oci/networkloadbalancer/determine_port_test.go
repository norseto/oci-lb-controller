package networkloadbalancer

import (
	"testing"

	api "github.com/norseto/oci-lb-controller/api/v1alpha1"
)

func TestDeterminePortPrefersPort(t *testing.T) {
	spec := api.LBRegistrarSpec{Port: 80, NodePort: 30000}
	if p := determinePort(spec); p != 80 {
		t.Errorf("expected 80, got %d", p)
	}
}

func TestDeterminePortFallsBackToNodePort(t *testing.T) {
	spec := api.LBRegistrarSpec{NodePort: 30000}
	if p := determinePort(spec); p != 30000 {
		t.Errorf("expected 30000, got %d", p)
	}
}
