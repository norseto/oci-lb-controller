package networkloadbalancer

import (
	"testing"

	api "github.com/norseto/oci-lb-controller/api/v1alpha1"
)

func TestDeterminePort(t *testing.T) {
	spec := api.LBRegistrarSpec{NodePort: 8080}
	if p := determinePort(spec); p != 8080 {
		t.Errorf("expected 8080 got %d", p)
	}
	spec.Port = 9090
	if p := determinePort(spec); p != 9090 {
		t.Errorf("expected 9090 got %d", p)
	}
}
