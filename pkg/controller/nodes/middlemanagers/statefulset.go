package middlemanager

import (
	binaryomenv1alpha1 "github.com/BinaryOmen/druid-operator/pkg/apis/binaryomen/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func MakeStatefulSetMM(c *binaryomenv1alpha1.Druid) *appsv1.StatefulSet {
	return &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      MakeNodeName(c),
			Namespace: c.Namespace,
		},
		Spec: makeStatefulSetSpec(c),
	}
}

func makeStatefulSetSpec(c *binaryomenv1alpha1.Druid) appsv1.StatefulSetSpec {

	s := appsv1.StatefulSetSpec{
		ServiceName: "middlemanager",
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app": "druid",
			},
		},
		Replicas:            &c.Spec.MiddleManager.Replicas,
		Template:            makeStatefulSetPodTemplate(c),
		PodManagementPolicy: appsv1.OrderedReadyPodManagement,
		UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
			Type: appsv1.RollingUpdateStatefulSetStrategyType,
		},
		VolumeClaimTemplates: GetVolumeClaimTemplates(c.Spec.MiddleManager.VolumeClaimTemplates),
	}

	return s
}

func makeStatefulSetPodTemplate(c *binaryomenv1alpha1.Druid) v1.PodTemplateSpec {
	return v1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: MakeNodeName(c),
			Labels: map[string]string{
				"app": "druid",
			},
		},
		Spec: makePodSpec(c),
	}
}

func makePodSpec(c *binaryomenv1alpha1.Druid) v1.PodSpec {

	spec := v1.PodSpec{
		Volumes: GetVolumes(c.Spec.MiddleManager.Volumes),
		Containers: []v1.Container{
			{
				Name:    c.Spec.MiddleManager.NodeType,
				Image:   c.Spec.Image,
				Command: GetCommand(c),
				Ports: []v1.ContainerPort{
					{
						Name:          c.Spec.MiddleManager.NodeType,
						ContainerPort: c.Spec.MiddleManager.Port,
						Protocol:      v1.Protocol("TCP"),
					},
				},
				VolumeMounts: GetVolumeMounts(c, c.Spec.MiddleManager.VolumeMounts),
			},
		},
	}
	return spec
}
