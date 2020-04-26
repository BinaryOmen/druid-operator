package nodes

import (
	"fmt"

	binaryomenv1alpha1 "github.com/BinaryOmen/druid-operator/pkg/apis/binaryomen/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MakeConfigMap for Historicals
func MakeConfigMapNode(cc *binaryomenv1alpha1.NodeSpec, c *binaryomenv1alpha1.Druid) *v1.ConfigMap {
	return &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      makeConfigMapName(cc),
			Namespace: c.Namespace,
		},
		Data: map[string]string{
			"runtime.properties": fmt.Sprintf("%s", cc.RuntimeProperties),
			"jvm.options":        fmt.Sprintf("%s", getJVM(cc, c)),
			"log4j2.xml":         fmt.Sprintf("%s", getLog4jConfig(cc, c)),
		},
	}
}

func MakeConfigMapCommon(cc *binaryomenv1alpha1.NodeSpec, c *binaryomenv1alpha1.Druid) *v1.ConfigMap {
	return &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "common",
			Namespace: c.Namespace,
		},
		Data: map[string]string{
			"common.runtime.properties": fmt.Sprintf("%s", c.Spec.CommonRuntimeProperties),
		},
	}
}

func makeConfigMapName(cc *binaryomenv1alpha1.NodeSpec) string {
	return fmt.Sprintf("%s", cc.Name)
}

func getJVM(cc *binaryomenv1alpha1.NodeSpec, c *binaryomenv1alpha1.Druid) string {
	if cc.JvmOptions != "" {
		return cc.JvmOptions
	} else {
		return c.Spec.JvmOptions
	}
}

func getLog4jConfig(cc *binaryomenv1alpha1.NodeSpec, c *binaryomenv1alpha1.Druid) string {
	if cc.Log4jConfig != "" {
		return cc.JvmOptions
	} else {
		return c.Spec.Log4jConfig
	}
}
