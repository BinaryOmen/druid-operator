package druid

import (
	"context"
	"fmt"

	middlemanager "github.com/BinaryOmen/druid-operator/pkg/controller/nodes/middlemanagers"
	"github.com/BinaryOmen/druid-operator/pkg/controller/sync"

	binaryomenv1alpha1 "github.com/BinaryOmen/druid-operator/pkg/apis/binaryomen/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *ReconcileDruid) reconcileMM(instance *binaryomenv1alpha1.Druid) error {

	for _, fun := range []reconcileFun{
		r.reconcileStatefulSetMM,
		r.reconcileMMConfigMap,
	} {
		if err := fun(instance); err != nil {
			r.log.Error(err, "Reconciling DruidCluster MiddleManager Error", instance)
			return err
		}
	}
	return nil
}

func (r *ReconcileDruid) reconcileStatefulSetMM(instance *binaryomenv1alpha1.Druid) (err error) {
	ssCreate := middlemanager.MakeStatefulSetMM(instance)

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
			fmt.Printf("%s %s", instance.Namespace, ssCreate.GetName())
		}
	} else if err != nil {
		return err
	} else {
		if instance.Spec.MiddleManager.Replicas != *ssCur.Spec.Replicas {
			old := *ssCur.Spec.Replicas
			ssCur.Spec.Replicas = &instance.Spec.MiddleManager.Replicas
			if err = r.client.Update(context.TODO(), ssCur); err == nil {
				r.log.Info("Scale MiddleManager statefulSet success",
					"OldSize", old,
					"NewSize", instance.Spec.MiddleManager.Replicas)
			}

		}
		return r.updateStatefulSetMM(instance, ssCur, ssCreate)
	}

	r.log.Info("MiddleManager node num info",
		"Replicas", ssCur.Status.Replicas,
		"ReadyNum", ssCur.Status.ReadyReplicas,
		"CurrentNum", ssCur.Status.CurrentReplicas,
	)
	return
}

func (r *ReconcileDruid) reconcileMMConfigMap(instance *binaryomenv1alpha1.Druid) (err error) {
	cmCreate := middlemanager.MakeConfigMapMM(instance)

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
			r.log.Info("Create MiddleManager config map success",
				"ConfigMap.Namespace", instance.Namespace,
				"ConfigMap.Name", cmCreate.GetName())
		}
	} else if err != nil {
		return err
	} else {
		if err = r.client.Update(context.TODO(), cmCur); err == nil {
			r.log.Info("Update MiddleManager configmap success")
		}
		return r.updateMMCm(instance, cmCur, cmCreate)
	}
	return
}

func (r *ReconcileDruid) updateStatefulSetMM(instance *binaryomenv1alpha1.Druid, foundSts *appsv1.StatefulSet, sts *appsv1.StatefulSet) (err error) {
	r.log.Info("Updating StatefulSet",
		"StatefulSet.Namespace", foundSts.Namespace,
		"StatefulSet.Name", foundSts.Name)
	sync.SyncStatefulSet(foundSts, sts)
	err = r.client.Update(context.TODO(), foundSts)
	if err != nil {
		return err
	}

	return nil
}

func (r *ReconcileDruid) updateMMCm(instance *binaryomenv1alpha1.Druid, foundCm *v1.ConfigMap, cm *v1.ConfigMap) (err error) {
	r.log.Info("Updating CM",
		"ConfigMap.Namespace", foundCm.Namespace,
		"ConfigMap.Name", foundCm.Name)
	sync.SyncCm(foundCm, cm)
	err = r.client.Update(context.TODO(), foundCm)
	if err != nil {
		return err
	}

	return nil
}
