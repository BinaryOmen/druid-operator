package validation

import (
	binaryomenv1alpha1 "github.com/BinaryOmen/druid-operator/pkg/apis/binaryomen/v1alpha1"
)

type Validator struct {
	Validated    bool
	ErrorMessage string
}

// Validate Druid
func (v *Validator) Validate(c *binaryomenv1alpha1.Druid) {
	v.Validated = true

	if c.Spec.CommonRuntimeProperties == "" {
		v.ErrorMessage = v.ErrorMessage + "CommonRuntimeProperties missing from Druid Cluster Spec\n"
		v.Validated = false
	}

	if c.Spec.CommonConfigMountPath == "" {
		v.ErrorMessage = v.ErrorMessage + "CommonConfigMountPath missing from Druid Cluster Spec\n"
		v.Validated = false
	}

	if c.Spec.StartScript == "" {
		v.ErrorMessage = v.ErrorMessage + "StartScript missing from Druid Cluster Spec\n"
		v.Validated = false
	}

	if c.Spec.Image == "" {
		v.ErrorMessage = v.ErrorMessage + "Image missing from Druid Cluster Spec\n"
		v.Validated = false
	}

	for _, n := range c.Spec.Nodes {
		if n.NodeType == "" {
			v.ErrorMessage = v.ErrorMessage + "NodeType missing from Druid Node Spec\n"
			v.Validated = false
		}

		if n.Replicas < 1 {
			v.ErrorMessage = v.ErrorMessage + "Minimum of one Replicas needed in Druid Node Spec\n"
			v.Validated = false
		}

		if n.RuntimeProperties == "" {
			v.ErrorMessage = v.ErrorMessage + "RuntimeProperties missing in Druid Node Spec\n"
			v.Validated = false
		}

		if n.MountPath == "" {
			v.ErrorMessage = v.ErrorMessage + "MountPath missing in Druid Node Spec\n"
			v.Validated = false
		}

		if n.Service.Port == 0 || n.Service.TargetPort == 0 {
			v.ErrorMessage = v.ErrorMessage + "Service is missing in Druid Node Spec\n"
			v.Validated = false
		}

		if n.Name == "" {
			v.ErrorMessage = v.ErrorMessage + "Node Name missing in Druid Node Spec\n"
			v.Validated = false
		}
	}
}
