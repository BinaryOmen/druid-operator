package druid

import (
	"context"
	"fmt"

	nodes "github.com/BinaryOmen/druid-operator/pkg/controller/nodes"
	"github.com/BinaryOmen/druid-operator/pkg/controller/sync"

	binaryomenv1alpha1 "github.com/BinaryOmen/druid-operator/pkg/apis/binaryomen/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const historical = "historical"
const middlemanager = "middlemanager"

func (r *ReconcileDruid) reconileDruid(c *binaryomenv1alpha1.NodeSpec, cc *binaryomenv1alpha1.Druid) error {

	for _, fun := range []reconcileFun{
		r.reconcileDruidNodes,
	} {
		if err := fun(c, cc); err != nil {
			r.log.Error(err, "Reconciling DruidCluster  Error", cc)
			return err
		}
	}
	//	time.Sleep(30 * time.Second)
	return nil
}

type keyAndNodeSpec struct {
	key  string
	spec binaryomenv1alpha1.NodeSpec
}

func getAllNodeSpecsInDruidPrescribedOrder(c *binaryomenv1alpha1.Druid) ([]keyAndNodeSpec, error) {
	nodeSpecsByNodeType := map[string][]keyAndNodeSpec{
		historical:    make([]keyAndNodeSpec, 0, 1),
		middlemanager: make([]keyAndNodeSpec, 0, 1),
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
	allNodeSpecs = append(allNodeSpecs, nodeSpecsByNodeType[middlemanager]...)

	return allNodeSpecs, nil
}

func (r *ReconcileDruid) reconcileDruidNodes(c *binaryomenv1alpha1.NodeSpec, cc *binaryomenv1alpha1.Druid) (err error) {
	allNodeSpecs, _ := getAllNodeSpecsInDruidPrescribedOrder(cc)

	for _, elem := range allNodeSpecs {

		ns := elem.spec

		if ns.NodeType == middlemanager || ns.NodeType == historical {
			sts := nodes.MakeStatefulSet(&ns, cc)
			r.reconcileSts(&ns, cc, sts)
		}
		cm := nodes.MakeConfigMap(&ns, cc)
		r.reconcileConfigMap(&ns, cc, cm)

	}
	return
}

func (r *ReconcileDruid) reconcileSts(c *binaryomenv1alpha1.NodeSpec, cc *binaryomenv1alpha1.Druid, sts *appsv1.StatefulSet) (err error) {
	ssCur := &appsv1.StatefulSet{}
	err = r.client.Get(context.TODO(), types.NamespacedName{
		Name:      sts.Name,
		Namespace: sts.Namespace,
	}, ssCur)
	if err != nil && errors.IsNotFound(err) {
		if err = controllerutil.SetControllerReference(cc, sts, r.scheme); err != nil {
			return err
		}

		if err = r.client.Create(context.TODO(), sts); err == nil {
			r.log.Info("Create statefulSet success",
				"StatefulSet.Namespace", cc.Namespace,
				"StatefulSet.Name", sts.GetName())
		}
	} else if err != nil {
		return err
	} else {
		if c.Replicas != *ssCur.Spec.Replicas {
			old := *ssCur.Spec.Replicas
			ssCur.Spec.Replicas = &c.Replicas
			if err = r.client.Update(context.TODO(), ssCur); err == nil {
				r.log.Info("Scale  statefulSet success.",
					"OldSize", old,
					"NewSize", c.Replicas)
			}

		}
		return r.updateStatefulSet(cc, ssCur, sts)
	}

	r.log.Info("Node node num info",
		"Replicas", ssCur.Status.Replicas,
		"ReadyNum", ssCur.Status.ReadyReplicas,
		"CurrentNum", ssCur.Status.CurrentReplicas,
	)
	return
}

func (r *ReconcileDruid) reconcileConfigMap(c *binaryomenv1alpha1.NodeSpec, cc *binaryomenv1alpha1.Druid, cmCreate *v1.ConfigMap) (err error) {
	cmCur := &v1.ConfigMap{}
	err = r.client.Get(context.TODO(), types.NamespacedName{
		Name:      cmCreate.Name,
		Namespace: cmCreate.Namespace,
	}, cmCur)
	if err != nil && errors.IsNotFound(err) {
		if err = controllerutil.SetControllerReference(cc, cmCreate, r.scheme); err != nil {
			return err
		}

		if err = r.client.Create(context.TODO(), cmCreate); err == nil {
			r.log.Info("Create  config map success",
				"ConfigMap.Namespace", cc.Namespace,
				"ConfigMap.Name", cmCreate.GetName())
		}
	} else if err != nil {
		return err
	} else {
		if err = r.client.Update(context.TODO(), cmCur); err == nil {
			r.log.Info("Update configmap success")
		}
		return r.updateCm(cc, cmCur, cmCreate)
	}
	return
}

func (r *ReconcileDruid) updateStatefulSet(instance *binaryomenv1alpha1.Druid, foundSts *appsv1.StatefulSet, sts *appsv1.StatefulSet) (err error) {
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

func (r *ReconcileDruid) updateCm(instance *binaryomenv1alpha1.Druid, foundCm *v1.ConfigMap, cm *v1.ConfigMap) (err error) {
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
