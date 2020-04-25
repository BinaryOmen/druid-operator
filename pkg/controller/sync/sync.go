package sync

import (
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/policy/v1beta1"
)

// SyncStatefulSet synchronizes any updates to the stateful-set
// Use Deepcopy funcs
func SyncStatefulSet(curr *appsv1.StatefulSet, next *appsv1.StatefulSet) {
	currSts := curr.DeepCopy()
	currSts.Spec.Replicas = next.Spec.Replicas
	currSts.Spec.Template = next.Spec.Template
	currSts.Spec.UpdateStrategy = next.Spec.UpdateStrategy

}

func SyncDeployment(curr *appsv1.Deployment, next *appsv1.Deployment) {
	currDeployment := curr.DeepCopy()
	currDeployment.Spec.Replicas = next.Spec.Replicas
	currDeployment.Spec.Template = next.Spec.Template
}

func SyncService(curr *v1.Service, next *v1.Service) {
	currSvc := curr.DeepCopy()
	currSvc.Spec.Ports = next.Spec.Ports
	currSvc.Spec.Type = next.Spec.Type
}

func SyncCm(curr *v1.ConfigMap, next *v1.ConfigMap) {
	currCm := curr.DeepCopy()
	currCm.Data = next.Data
	currCm.BinaryData = next.BinaryData
}

func SyncPdb(curr *v1beta1.PodDisruptionBudget, next *v1beta1.PodDisruptionBudget) {
	currPdb := curr.DeepCopy()
	currPdb.Spec.MaxUnavailable = next.Spec.MaxUnavailable
}
