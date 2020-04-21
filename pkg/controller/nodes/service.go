package nodes

import (
	binaryomenv1alpha1 "github.com/BinaryOmen/druid-operator/pkg/apis/binaryomen/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func MakeService(cc *binaryomenv1alpha1.NodeSpec, c *binaryomenv1alpha1.Druid) *v1.Service {

	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cc.Name,
			Namespace: c.Namespace,
			Labels: map[string]string{
				"app":  "druid",
				"type": cc.NodeType,
				"name": cc.Name,
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Port: cc.Service.Port,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Type(0),
						IntVal: cc.Service.TargetPort,
					},
					NodePort: 0,
				},
			},
			Selector: map[string]string{
				"name": cc.Name,
			},
			ClusterIP: "",
			Type:      getServiceType(cc),
		},
	}
}

func getServiceType(cc *binaryomenv1alpha1.NodeSpec) v1.ServiceType {
	if cc.Service.Type == "" {
		return v1.ServiceTypeClusterIP
	}
	return cc.Service.Type
}
