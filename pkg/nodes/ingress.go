package nodes

import (
	binaryomenv1alpha1 "github.com/BinaryOmen/druid-operator/pkg/apis/binaryomen/v1alpha1"
	extensions "k8s.io/api/extensions/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func MakeDruidIngress(cc *binaryomenv1alpha1.NodeSpec, c *binaryomenv1alpha1.Druid) *extensions.Ingress {
	return &extensions.Ingress{
		ObjectMeta: v1.ObjectMeta{
			Name:      cc.Name,
			Namespace: c.Namespace,
			Labels: map[string]string{
				"app":  "druid",
				"type": cc.NodeType,
				"name": cc.Name,
			},
			Annotations: getIngressAnnotations(cc),
		},
		Spec: getIngressSpec(cc),
	}
}

func getIngressTLS(cc *binaryomenv1alpha1.NodeSpec) []extensions.IngressTLS {
	if cc.Ingress.Enabled == false {
		return nil
	}

	if cc.Ingress.TLSEnabled {
		return []extensions.IngressTLS{
			{
				Hosts:      []string{cc.Ingress.Hostname},
				SecretName: cc.Ingress.TLSSecretName,
			},
		}
	}
	return nil
}

func getIngressSpec(cc *binaryomenv1alpha1.NodeSpec) extensions.IngressSpec {
	return extensions.IngressSpec{
		TLS: getIngressTLS(cc),
		Rules: []extensions.IngressRule{
			{
				Host: GetHost(cc),
				IngressRuleValue: extensions.IngressRuleValue{
					HTTP: &extensions.HTTPIngressRuleValue{
						Paths: []extensions.HTTPIngressPath{
							{
								Path: GetPath(cc),
								Backend: extensions.IngressBackend{
									ServiceName: cc.Name,
									ServicePort: intstr.FromInt(int(cc.Service.Port)),
								},
							},
						},
					},
				},
			},
		},
	}
}

func GetHost(cc *binaryomenv1alpha1.NodeSpec) string {
	if cc.Ingress.Enabled == false {
		return ""
	}
	return cc.Ingress.Hostname
}

func GetPath(cc *binaryomenv1alpha1.NodeSpec) string {
	if cc.Ingress.Enabled == false {
		return "/"
	}
	return cc.Ingress.Path
}

func getIngressAnnotations(cc *binaryomenv1alpha1.NodeSpec) map[string]string {
	annotations := make(map[string]string)

	if cc.Ingress.Annotations == nil {
		annotations["app"] = cc.Name
		return annotations
	} else {
		return cc.Ingress.Annotations
	}
}
