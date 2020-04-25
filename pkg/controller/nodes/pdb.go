package nodes

import (
	binaryomenv1alpha1 "github.com/BinaryOmen/druid-operator/pkg/apis/binaryomen/v1alpha1"
	"k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func MakePodDisruptionBudget(cc *binaryomenv1alpha1.NodeSpec, c *binaryomenv1alpha1.Druid) (*v1beta1.PodDisruptionBudget, error) {
	pdb := &v1beta1.PodDisruptionBudget{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "policy/v1beta1",
			Kind:       "PodDisruptionBudget",
		},

		ObjectMeta: metav1.ObjectMeta{
			Name:      cc.Name,
			Namespace: c.Namespace,
			Labels: map[string]string{
				"app":  "druid",
				"type": cc.NodeType,
				"name": cc.Name,
			},
		},

		Spec: v1beta1.PodDisruptionBudgetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":  "druid",
					"type": cc.NodeType,
					"name": cc.Name,
				},
			},
			MaxUnavailable: &intstr.IntOrString{
				Type:   intstr.Type(0),
				IntVal: int32(cc.MaxUnavailable),
			},
		},
	}

	return pdb, nil
}
