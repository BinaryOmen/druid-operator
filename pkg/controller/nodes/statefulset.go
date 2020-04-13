package nodes

import (
	"fmt"

	binaryomenv1alpha1 "github.com/BinaryOmen/druid-operator/pkg/apis/binaryomen/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func MakeStatefulSet(cc *binaryomenv1alpha1.NodeSpec, c *binaryomenv1alpha1.Druid) *appsv1.StatefulSet {

	return &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      makeNodeName(cc),
			Namespace: c.Namespace,
		},
		Spec: makeStatefulSetSpec(cc, c),
	}
}

func makeStatefulSetSpec(cc *binaryomenv1alpha1.NodeSpec, c *binaryomenv1alpha1.Druid) appsv1.StatefulSetSpec {

	s := appsv1.StatefulSetSpec{
		ServiceName: cc.NodeType,
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app": "druid",
			},
		},
		Replicas:            &cc.Replicas,
		Template:            makeStatefulSetPodTemplate(cc, c),
		PodManagementPolicy: appsv1.OrderedReadyPodManagement,
		UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
			Type: appsv1.RollingUpdateStatefulSetStrategyType,
		},
		VolumeClaimTemplates: GetVolumeClaimTemplates(cc.VolumeClaimTemplates),
	}

	return s
}

func makeStatefulSetPodTemplate(c *binaryomenv1alpha1.NodeSpec, cc *binaryomenv1alpha1.Druid) v1.PodTemplateSpec {
	return v1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: makeNodeName(c),
			Labels: map[string]string{
				"app": "druid",
			},
		},
		Spec: makePodSpec(c, cc),
	}
}

func makePodSpec(cc *binaryomenv1alpha1.NodeSpec, c *binaryomenv1alpha1.Druid) v1.PodSpec {

	spec := v1.PodSpec{
		Volumes: GetVolumes(cc, cc.Volumes),
		Containers: []v1.Container{
			{
				Name:    cc.NodeType,
				Image:   c.Spec.Image,
				Command: GetCommand(cc, c),
				Ports: []v1.ContainerPort{
					{
						Name:          cc.NodeType,
						ContainerPort: cc.Port,
						Protocol:      v1.Protocol("TCP"),
					},
				},
				VolumeMounts: GetVolumeMounts(cc, cc.VolumeMounts),
			},
		},
	}
	return spec
}

func makeNodeName(cc *binaryomenv1alpha1.NodeSpec) string {
	return fmt.Sprintf("druid-%s", cc.NodeType)
}

func GetCommand(cc *binaryomenv1alpha1.NodeSpec, c *binaryomenv1alpha1.Druid) []string {
	return []string{c.Spec.StartScript, cc.NodeType}
}

func GetVolumeMounts(cc *binaryomenv1alpha1.NodeSpec, vmM []v1.VolumeMount) []v1.VolumeMount {
	volumeMount := []v1.VolumeMount{
		{
			Name:      makeConfigMapName(cc),
			MountPath: cc.MountPath,
		},
	}
	for _, val := range vmM {
		volumeMount = append(volumeMount, val)
	}
	return volumeMount
}

func GetVolumes(cc *binaryomenv1alpha1.NodeSpec, vm []v1.Volume) []v1.Volume {
	volumes := []v1.Volume{
		{
			Name: makeConfigMapName(cc),
			VolumeSource: v1.VolumeSource{
				ConfigMap: &v1.ConfigMapVolumeSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: makeConfigMapName(cc),
					},
				},
			},
		},
	}

	for _, val := range vm {
		volumes = append(volumes, val)
	}
	return volumes
}

func GetVolumeClaimTemplates(vcT []v1.PersistentVolumeClaim) []v1.PersistentVolumeClaim {
	pvc := []v1.PersistentVolumeClaim{}

	for _, val := range vcT {
		pvc = append(pvc, val)
	}
	return pvc
}
