package middlemanager

import (
	"fmt"

	binaryomenv1alpha1 "github.com/BinaryOmen/druid-operator/pkg/apis/binaryomen/v1alpha1"
	v1 "k8s.io/api/core/v1"
)

func MakeNodeName(c *binaryomenv1alpha1.Druid) string {
	return fmt.Sprintf("%s-%s", c.Name, c.Spec.MiddleManager.NodeType)
}

func GetCommand(c *binaryomenv1alpha1.Druid) []string {
	return []string{c.Spec.StartScript, "middlemanager"}
}

func GetVolumeMounts(c *binaryomenv1alpha1.Druid, vmM []v1.VolumeMount) []v1.VolumeMount {
	volumeMount := []v1.VolumeMount{
		{
			Name:      "runtime-properties-mm",
			MountPath: c.Spec.MiddleManager.MountPath,
		},
	}
	for _, val := range vmM {
		volumeMount = append(volumeMount, val)
	}
	return volumeMount
}

func GetVolumes(vm []v1.Volume) []v1.Volume {
	volumes := []v1.Volume{
		{
			Name: "runtime-properties-mm",
			VolumeSource: v1.VolumeSource{
				ConfigMap: &v1.ConfigMapVolumeSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: "runtime-properties-mm",
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
