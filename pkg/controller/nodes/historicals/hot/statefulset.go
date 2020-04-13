package historicals

import (
	"fmt"

	binaryomenv1alpha1 "github.com/BinaryOmen/druid-operator/pkg/apis/binaryomen/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type keyAndNodeSpec struct {
	key  string
	spec binaryomenv1alpha1.NodeSpec
}

const historical = "historical"

func getAllNodeSpecsInDruidPrescribedOrder(c *binaryomenv1alpha1.Druid) ([]keyAndNodeSpec, error) {
	nodeSpecsByNodeType := map[string][]keyAndNodeSpec{
		historical: make([]keyAndNodeSpec, 0, 1),
	}

	for key, nodeSpec := range c.Spec.Nodes {
		nodeSpecs := nodeSpecsByNodeType[nodeSpec.NodeType]
		if nodeSpecs == nil {
			return nil, fmt.Errorf("druidSpec[%s:%s] has invalid NodeType[%s]. Deployment aborted", c.Kind, c.Name, nodeSpec.NodeType)
		} else {
			nodeSpecsByNodeType[nodeSpec.NodeType] = append(nodeSpecs, keyAndNodeSpec{key, nodeSpec})
		}
	}

	allNodeSpecs := make([]keyAndNodeSpec, 0, len(c.Spec.Nodes))

	allNodeSpecs = append(allNodeSpecs, nodeSpecsByNodeType[historical]...)

	return allNodeSpecs, nil
}

func Create(c *binaryomenv1alpha1.NodeSpec, cc *binaryomenv1alpha1.Druid) (sts *appsv1.StatefulSet, cm *v1.ConfigMap) {
	allNodeSpecs, _ := getAllNodeSpecsInDruidPrescribedOrder(cc)

	for _, elem := range allNodeSpecs {
		//key := elem.key
		nodeSpec := elem.spec

		if nodeSpec.NodeType == historical {
			sts := MakeStatefulSetHot(&nodeSpec, cc)
			cm := MakeConfigMapHot(&nodeSpec, cc)
			return sts, cm
		}
	}
	return
}

func MakeStatefulSetHot(c *binaryomenv1alpha1.NodeSpec, cc *binaryomenv1alpha1.Druid) *appsv1.StatefulSet {

	return &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "historical",
			Namespace: cc.Namespace,
		},
		Spec: makeStatefulSetSpec(c, cc),
	}
}

func makeStatefulSetSpec(c *binaryomenv1alpha1.NodeSpec, cc *binaryomenv1alpha1.Druid) appsv1.StatefulSetSpec {

	s := appsv1.StatefulSetSpec{
		ServiceName: "historical",
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app": "druid",
			},
		},
		Replicas:            &c.Replicas,
		Template:            makeStatefulSetPodTemplate(c, cc),
		PodManagementPolicy: appsv1.OrderedReadyPodManagement,
		UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
			Type: appsv1.RollingUpdateStatefulSetStrategyType,
		},
		VolumeClaimTemplates: GetVolumeClaimTemplates(c.VolumeClaimTemplates),
	}

	return s
}

func makeStatefulSetPodTemplate(c *binaryomenv1alpha1.NodeSpec, cc *binaryomenv1alpha1.Druid) v1.PodTemplateSpec {
	return v1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: MakeNodeName(c),
			Labels: map[string]string{
				"app": "druid",
			},
		},
		Spec: makePodSpec(c, cc),
	}
}

func makePodSpec(c *binaryomenv1alpha1.NodeSpec, cc *binaryomenv1alpha1.Druid) v1.PodSpec {

	spec := v1.PodSpec{
		Volumes: GetVolumes(c.Volumes),
		Containers: []v1.Container{
			{
				Name:    c.NodeType,
				Image:   cc.Spec.Image,
				Command: GetCommand(cc),
				Ports: []v1.ContainerPort{
					{
						Name:          c.NodeType,
						ContainerPort: c.Port,
						Protocol:      v1.Protocol("TCP"),
					},
				},
				VolumeMounts: GetVolumeMounts(c, c.VolumeMounts),
			},
		},
	}
	return spec
}

func MakeNodeName(c *binaryomenv1alpha1.NodeSpec) string {
	return fmt.Sprintf("%s", c.NodeType)
}

func GetCommand(c *binaryomenv1alpha1.Druid) []string {
	return []string{c.Spec.StartScript, "historical"}
}

func GetVolumeMounts(c *binaryomenv1alpha1.NodeSpec, vmM []v1.VolumeMount) []v1.VolumeMount {
	volumeMount := []v1.VolumeMount{
		{
			Name:      "runtime-properties-hot",
			MountPath: c.MountPath,
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
			Name: "runtime-properties-hot",
			VolumeSource: v1.VolumeSource{
				ConfigMap: &v1.ConfigMapVolumeSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: "runtime-properties-hot",
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
