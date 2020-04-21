package sync

import (
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

// SyncStatefulSet synchronizes any updates to the stateful-set
// TODO: Change Sync to deepcopy
func SyncStatefulSet(curr *appsv1.StatefulSet, next *appsv1.StatefulSet) {
	curr.Spec.Replicas = next.Spec.Replicas
	curr.Spec.Template = next.Spec.Template
	curr.Spec.UpdateStrategy = next.Spec.UpdateStrategy

}

func SyncDeployment(curr *appsv1.Deployment, next *appsv1.Deployment) {
	curr.Spec.Replicas = next.Spec.Replicas
	curr.Spec.Template = next.Spec.Template
}

func SyncService(curr *v1.Service, next *v1.Service) {
	curr.Spec.Ports = next.Spec.Ports
	curr.Spec.Type = next.Spec.Type
}

func SyncCm(curr *v1.ConfigMap, next *v1.ConfigMap) {
	curr.Data = next.Data
	curr.BinaryData = next.BinaryData
}
