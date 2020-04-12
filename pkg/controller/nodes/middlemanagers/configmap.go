package middlemanager

import (
	binaryomenv1alpha1 "github.com/BinaryOmen/druid-operator/pkg/apis/binaryomen/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MakeConfigMap for Historicals
func MakeConfigMapMM(c *binaryomenv1alpha1.Druid) *v1.ConfigMap {
	return &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "runtime-properties-mm",
			Namespace: c.Namespace,
		},
		Data: map[string]string{
			"runtime.properties": c.Spec.MiddleManager.RuntimeProperties,
		},
	}
}
