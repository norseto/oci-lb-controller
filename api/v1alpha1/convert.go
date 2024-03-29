package v1alpha1

import (
	"github.com/norseto/oci-lb-controller/api/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

// ConvertTo converts this LBRegistrar v2 to the Hub version (v1)
func (src *LBRegistrar) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha2.LBRegistrar)
	dst.ObjectMeta = src.ObjectMeta
	dst.TypeMeta = src.TypeMeta
	dst.Spec = v1alpha2.LBRegistrarSpec{
		LoadBalancerId: src.Spec.LoadBalancerId,
		NodePort:       src.Spec.Port,
		Weight:         src.Spec.Weight,
		BackendSetName: src.Spec.BackendSetName,
		ApiKey: v1alpha2.ApiKeySpec{
			User:        src.Spec.ApiKey.User,
			Fingerprint: src.Spec.ApiKey.Fingerprint,
			Region:      src.Spec.ApiKey.Region,
			PrivateKey: v1alpha2.PrivateKeySpec{
				Namespace:    src.Spec.ApiKey.PrivateKey.Namespace,
				SecretKeyRef: src.Spec.ApiKey.PrivateKey.SecretKeyRef,
			},
		},
	}
	dst.Status = v1alpha2.LBRegistrarStatus{
		Phase: src.Status.Phase,
	}
	return nil
}

// ConvertFrom converts from the Hub version (v1) to this LBRegistrar version
func (dst *LBRegistrar) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1alpha2.LBRegistrar)
	dst.ObjectMeta = src.ObjectMeta
	dst.TypeMeta = src.TypeMeta
	dst.Spec = LBRegistrarSpec{
		LoadBalancerId: src.Spec.LoadBalancerId,
		Port:           src.Spec.NodePort,
		Weight:         src.Spec.Weight,
		BackendSetName: src.Spec.BackendSetName,
		ApiKey: ApiKeySpec{
			User:        src.Spec.ApiKey.User,
			Fingerprint: src.Spec.ApiKey.Fingerprint,
			Region:      src.Spec.ApiKey.Region,
			PrivateKey: PrivateKeySpec{
				Namespace:    src.Spec.ApiKey.PrivateKey.Namespace,
				SecretKeyRef: src.Spec.ApiKey.PrivateKey.SecretKeyRef,
			},
		},
	}
	dst.Status = LBRegistrarStatus{
		Phase: src.Status.Phase,
	}
	return nil
}
