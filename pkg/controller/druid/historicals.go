package druid

import (
	"context"

	"github.com/BinaryOmen/druid-operator/pkg/controller/nodes"

	binaryomenv1alpha1 "github.com/BinaryOmen/druid-operator/pkg/apis/binaryomen/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *ReconcileDruid) reconcileHistorical(instance *binaryomenv1alpha1.Druid) error {

	for _, fun := range []reconcileFun{
		r.reconcileStatefulSet,
		r.reconcileHistoricalConfigMap,
	} {
		if err := fun(instance); err != nil {
			r.log.Error(err, "Reconciling DruidCluster Historical Error", instance)
			return err
		}
	}
	return nil
}

func (r *ReconcileDruid) reconcileStatefulSet(instance *binaryomenv1alpha1.Druid) (err error) {
	ssCreate := nodes.MakeStatefulSet(instance)

	ssCur := &appsv1.StatefulSet{}
	err = r.client.Get(context.TODO(), types.NamespacedName{
		Name:      ssCreate.Name,
		Namespace: ssCreate.Namespace,
	}, ssCur)
	if err != nil && errors.IsNotFound(err) {
		if err = controllerutil.SetControllerReference(instance, ssCreate, r.scheme); err != nil {
			return err
		}

		if err = r.client.Create(context.TODO(), ssCreate); err == nil {
			r.log.Info("Create historical statefulSet success",
				"StatefulSet.Namespace", instance.Namespace,
				"StatefulSet.Name", ssCreate.GetName())
		}
	} else if err != nil {
		return err
	} else {
		if instance.Spec.Historicals.Replicas != *ssCur.Spec.Replicas {
			old := *ssCur.Spec.Replicas
			ssCur.Spec.Replicas = &instance.Spec.Historicals.Replicas
			if err = r.client.Update(context.TODO(), ssCur); err == nil {
				r.log.Info("Scale Historical statefulSet success",
					"OldSize", old,
					"NewSize", instance.Spec.Historicals.Replicas)
			}
		}
	}

	r.log.Info("Historical node num info",
		"Replicas", ssCur.Status.Replicas,
		"ReadyNum", ssCur.Status.ReadyReplicas,
		"CurrentNum", ssCur.Status.CurrentReplicas,
	)
	return
}

func (r *ReconcileDruid) reconcileHistoricalConfigMap(instance *binaryomenv1alpha1.Druid) (err error) {
	cmCreate := nodes.MakeConfigMap(instance)

	cmCur := &v1.ConfigMap{}
	err = r.client.Get(context.TODO(), types.NamespacedName{
		Name:      cmCreate.Name,
		Namespace: cmCreate.Namespace,
	}, cmCur)
	if err != nil && errors.IsNotFound(err) {
		if err = controllerutil.SetControllerReference(instance, cmCreate, r.scheme); err != nil {
			return err
		}

		if err = r.client.Create(context.TODO(), cmCreate); err == nil {
			r.log.Info("Create historical config map success",
				"ConfigMap.Namespace", instance.Namespace,
				"ConfigMap.Name", cmCreate.GetName())
		}
	}
	return
}

func (r *ReconcileDruid) updateStatefulSet(instance *binaryomenv1alpha1.Druid, foundSts *appsv1.StatefulSet, sts *appsv1.StatefulSet) (err error) {
	r.log.Info("Updating StatefulSet",
		"StatefulSet.Namespace", foundSts.Namespace,
		"StatefulSet.Name", foundSts.Name)
	SyncStatefulSet(foundSts, sts)
	err = r.client.Update(context.TODO(), foundSts)
	if err != nil {
		return err
	}

	return nil
}

// SyncStatefulSet synchronizes any updates to the stateful-set
func SyncStatefulSet(curr *appsv1.StatefulSet, next *appsv1.StatefulSet) {
	curr.Spec.Replicas = next.Spec.Replicas
	curr.Spec.Template = next.Spec.Template
	curr.Spec.UpdateStrategy = next.Spec.UpdateStrategy
}
