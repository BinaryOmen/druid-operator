package nodes

import (
	"fmt"

	binaryomenv1alpha1 "github.com/BinaryOmen/druid-operator/pkg/apis/binaryomen/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func MakeStatefulSet(c *binaryomenv1alpha1.Druid) *appsv1.StatefulSet {
	return &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      MakeStatefulSetName(c),
			Namespace: c.Namespace,
		},
		Spec: makeStatefulSetSpec(c),
	}
}

func MakeStatefulSetName(c *binaryomenv1alpha1.Druid) string {
	return fmt.Sprintf("%s-historical", c.GetName())
}

func makeStatefulSetSpec(c *binaryomenv1alpha1.Druid) appsv1.StatefulSetSpec {
	s := appsv1.StatefulSetSpec{
		ServiceName: "historical",
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app": "druid",
			},
		},
		Replicas:            &c.Spec.Historicals.Replicas,
		Template:            makeStatefulSetPodTemplate(c),
		PodManagementPolicy: appsv1.OrderedReadyPodManagement,
		UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
			Type: appsv1.RollingUpdateStatefulSetStrategyType,
		},
		VolumeClaimTemplates: getVolumeClaimTemplates(c),
	}

	return s
}

func makeStatefulSetPodTemplate(c *binaryomenv1alpha1.Druid) v1.PodTemplateSpec {
	return v1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: c.GetName(),
			Labels: map[string]string{
				"app": "druid",
			},
		},
		Spec: makePodSpec(c),
	}
}

func makePodSpec(c *binaryomenv1alpha1.Druid) v1.PodSpec {
	spec := v1.PodSpec{
		Volumes: []v1.Volume{
			{
				Name: "runtime-properties",
				VolumeSource: v1.VolumeSource{
					ConfigMap: &v1.ConfigMapVolumeSource{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "runtime-properties",
						},
					},
				},
			},
		},
		Containers: []v1.Container{
			{
				Name:    "historicals",
				Image:   c.Spec.Image,
				Command: []string{c.Spec.StartScript, "historical"},
				Ports: []v1.ContainerPort{
					v1.ContainerPort{
						Name:          "http",
						ContainerPort: c.Spec.Historicals.Port,
						Protocol:      v1.Protocol("TCP"),
					},
				},
				VolumeMounts: []v1.VolumeMount{
					v1.VolumeMount{
						Name:      "runtime-properties",
						MountPath: c.Spec.Historicals.MountPath,
					},
				},
			},
		},
	}
	return spec
}

func getVolumeClaimTemplates(c *binaryomenv1alpha1.Druid) []v1.PersistentVolumeClaim {
	pvc := []v1.PersistentVolumeClaim{}

	for _, val := range c.Spec.Historicals.CommonNode.VolumeClaimTemplates {
		pvc = append(pvc, val)
	}
	return pvc
}

func MakeConfigMap(c *binaryomenv1alpha1.Druid) *v1.ConfigMap {
	return &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "runtime-properties",
			Namespace: c.Namespace,
		},
		Data: map[string]string{
			"runtime.properties": c.Spec.Historicals.RuntimeProperties,
		},
	}
}
