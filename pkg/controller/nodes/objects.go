package nodes

import (
	"fmt"

	binaryomenv1alpha1 "github.com/BinaryOmen/druid-operator/pkg/apis/binaryomen/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
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

func MakeDeployment(cc *binaryomenv1alpha1.NodeSpec, c *binaryomenv1alpha1.Druid) *appsv1.Deployment {
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      makeNodeName(cc),
			Namespace: c.Namespace,
		},
		Spec: makeDeploymentSpec(cc, c),
	}
}

func makeStatefulSetSpec(cc *binaryomenv1alpha1.NodeSpec, c *binaryomenv1alpha1.Druid) appsv1.StatefulSetSpec {

	s := appsv1.StatefulSetSpec{
		ServiceName: cc.Name,
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app":  "druid",
				"type": cc.NodeType,
				"name": cc.Name,
			},
		},
		Replicas:            &cc.Replicas,
		Template:            makePodTemplate(cc, c),
		PodManagementPolicy: appsv1.OrderedReadyPodManagement,
		UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
			Type: appsv1.RollingUpdateStatefulSetStrategyType,
		},
		VolumeClaimTemplates: getVolumeClaimTemplates(cc.VolumeClaimTemplates),
	}

	return s
}
func makeDeploymentSpec(cc *binaryomenv1alpha1.NodeSpec, c *binaryomenv1alpha1.Druid) appsv1.DeploymentSpec {

	d := appsv1.DeploymentSpec{
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app":  "druid",
				"type": cc.NodeType,
				"name": cc.Name,
			},
		},
		Replicas: &cc.Replicas,
		Template: makePodTemplate(cc, c),
		Strategy: appsv1.DeploymentStrategy{
			Type:          "RollingUpdate",
			RollingUpdate: getRollingUpdateStrategy(),
		},
	}

	return d
}

func getRollingUpdateStrategy() *appsv1.RollingUpdateDeployment {
	var maxUnaval intstr.IntOrString = intstr.FromInt(25)
	var maxSurge intstr.IntOrString = intstr.FromInt(25)
	return &appsv1.RollingUpdateDeployment{
		MaxUnavailable: &maxUnaval,
		MaxSurge:       &maxSurge,
	}
}

func makePodTemplate(cc *binaryomenv1alpha1.NodeSpec, c *binaryomenv1alpha1.Druid) v1.PodTemplateSpec {
	return v1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name: makeNodeName(cc),
			Labels: map[string]string{
				"app":  "druid",
				"type": cc.NodeType,
				"name": cc.Name,
			},
			Annotations: getAnnotations(cc),
		},
		Spec: makePodSpec(cc, c),
	}
}

func makePodSpec(cc *binaryomenv1alpha1.NodeSpec, c *binaryomenv1alpha1.Druid) v1.PodSpec {

	spec := v1.PodSpec{
		NodeSelector:     cc.NodeSelector,
		Tolerations:      getTolerations(cc, c),
		Affinity:         getAffinity(cc, c),
		Volumes:          getVolumes(cc, cc.Volumes),
		ImagePullSecrets: c.Spec.ImagePullSecrets,
		SecurityContext:  cc.SecurityContext,
		Containers: []v1.Container{
			{
				Name:                     cc.Name,
				Image:                    c.Spec.Image,
				Command:                  getCommand(cc, c),
				Resources:                cc.Resources,
				Env:                      getEnv(cc, c),
				TerminationMessagePath:   "/dev/termination-log",
				TerminationMessagePolicy: "File",
				Ports: []v1.ContainerPort{
					{
						Name:          cc.Name,
						ContainerPort: cc.Service.TargetPort,
						Protocol:      v1.Protocol("TCP"),
					},
				},
				VolumeMounts: getVolumeMounts(cc, c, cc.VolumeMounts),
			},
		},
	}
	return spec
}

func makeNodeName(cc *binaryomenv1alpha1.NodeSpec) string {
	return fmt.Sprintf("druid-%s", cc.Name)
}

func getAnnotations(cc *binaryomenv1alpha1.NodeSpec) map[string]string {
	annotations := make(map[string]string)

	if cc.Annotations == nil {
		annotations["app"] = "druid"
		return annotations
	} else {
		return cc.Annotations
	}
}

func getCommand(cc *binaryomenv1alpha1.NodeSpec, c *binaryomenv1alpha1.Druid) []string {
	return []string{c.Spec.StartScript, cc.NodeType}
}

func getVolumeMounts(cc *binaryomenv1alpha1.NodeSpec, c *binaryomenv1alpha1.Druid, vmM []v1.VolumeMount) []v1.VolumeMount {
	volumeMount := []v1.VolumeMount{
		{
			Name:      makeConfigMapName(cc),
			MountPath: cc.MountPath,
		},
		{
			Name:      "common",
			MountPath: c.Spec.CommonConfigMountPath,
		},
	}
	for _, val := range vmM {
		volumeMount = append(volumeMount, val)
	}
	return volumeMount
}

func getVolumes(cc *binaryomenv1alpha1.NodeSpec, vm []v1.Volume) []v1.Volume {
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
		{
			Name: "common",
			VolumeSource: v1.VolumeSource{
				ConfigMap: &v1.ConfigMapVolumeSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: "common",
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

func getVolumeClaimTemplates(vcT []v1.PersistentVolumeClaim) []v1.PersistentVolumeClaim {
	pvc := []v1.PersistentVolumeClaim{}

	for _, val := range vcT {
		pvc = append(pvc, val)
	}
	return pvc
}

func getTolerations(cc *binaryomenv1alpha1.NodeSpec, c *binaryomenv1alpha1.Druid) []v1.Toleration {
	tolerations := []v1.Toleration{}
	for _, val := range c.Spec.Tolerations {
		tolerations = append(tolerations, val)
	}
	for _, val := range cc.Tolerations {
		tolerations = append(tolerations, val)
	}
	return tolerations
}

func getAffinity(cc *binaryomenv1alpha1.NodeSpec, c *binaryomenv1alpha1.Druid) *v1.Affinity {

	if cc.Affinity != nil {
		return cc.Affinity
	} else {
		return c.Spec.Affinity
	}

}

func getEnv(cc *binaryomenv1alpha1.NodeSpec, c *binaryomenv1alpha1.Druid) []v1.EnvVar {
	env := []v1.EnvVar{}
	for _, val := range c.Spec.Env {
		env = append(env, val)
	}
	for _, val := range cc.Env {
		env = append(env, val)
	}
	return env
}
