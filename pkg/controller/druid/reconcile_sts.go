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
const overlord = "overlord"

func (r *ReconcileDruid) reconileDruid(cc *binaryomenv1alpha1.NodeSpec, c *binaryomenv1alpha1.Druid) error {

	for _, fun := range []reconcileFun{
		r.reconcileDruidNodes,
	} {
		if err := fun(cc, c); err != nil {
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
		overlord:      make([]keyAndNodeSpec, 0, 1),
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
	allNodeSpecs = append(allNodeSpecs, nodeSpecsByNodeType[overlord]...)

	return allNodeSpecs, nil
}

func (r *ReconcileDruid) reconcileDruidNodes(cc *binaryomenv1alpha1.NodeSpec, c *binaryomenv1alpha1.Druid) (err error) {
	allNodeSpecs, _ := getAllNodeSpecsInDruidPrescribedOrder(c)

	for _, elem := range allNodeSpecs {

		ns := elem.spec

		if ns.NodeType == historical || ns.NodeType == middlemanager {
			sts := nodes.MakeStatefulSet(&ns, c)
			r.reconcileSts(&ns, c, sts)
		}

		/*
			if ns.NodeType == overlord {
				d := nodes.Md(&ns, c)
				r.reconcileDeployment(&ns, c, d)
			}
		*/

		cmN := nodes.MakeConfigMapNode(&ns, c)
		cmC := nodes.MakeConfigMapCommon(&ns, c)
		r.reconcileConfigMap(&ns, c, cmN)
		r.reconcileConfigMap(&ns, c, cmC)

	}
	return
}

func (r *ReconcileDruid) reconcileSts(cc *binaryomenv1alpha1.NodeSpec, c *binaryomenv1alpha1.Druid, sts *appsv1.StatefulSet) (err error) {
	ssCur := &appsv1.StatefulSet{}
	err = r.client.Get(context.TODO(), types.NamespacedName{
		Name:      sts.Name,
		Namespace: sts.Namespace,
	}, ssCur)
	if err != nil && errors.IsNotFound(err) {
		if err = controllerutil.SetControllerReference(c, sts, r.scheme); err != nil {
			return err
		}

		if err = r.client.Create(context.TODO(), sts); err == nil {
			r.log.Info("Create statefulSet success",
				"StatefulSet.Namespace", c.Namespace,
				"StatefulSet.Name", sts.GetName())
		}
	} else if err != nil {
		return err
	} else {
		if cc.Replicas != *ssCur.Spec.Replicas {
			old := *ssCur.Spec.Replicas
			ssCur.Spec.Replicas = &cc.Replicas
			if err = r.client.Update(context.TODO(), ssCur); err == nil {
				r.log.Info("Scale  statefulSet success.",
					"OldSize", old,
					"NewSize", cc.Replicas)
			}

		}
		return r.updateStatefulSet(c, ssCur, sts)
	}

	r.log.Info("Node node num info",
		"Replicas", ssCur.Status.Replicas,
		"ReadyNum", ssCur.Status.ReadyReplicas,
		"CurrentNum", ssCur.Status.CurrentReplicas,
	)
	return
}

func (r *ReconcileDruid) reconcileDeployment(cc *binaryomenv1alpha1.NodeSpec, c *binaryomenv1alpha1.Druid, deploy *appsv1.Deployment) (err error) {
	deployCur := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), types.NamespacedName{
		Name:      deploy.Name,
		Namespace: deploy.Namespace,
	}, deployCur)
	if err != nil && errors.IsNotFound(err) {
		if err = controllerutil.SetControllerReference(c, deploy, r.scheme); err != nil {
			return err
		}

		if err = r.client.Create(context.TODO(), deploy); err == nil {
			r.log.Info("Create statefulSet success",
				"Deployment.Namespace", c.Namespace,
				"Deployment.Name", deploy.GetName())
		}
	} else if err != nil {
		r.log.Info("%s", err)
		return err
	} else {
		if cc.Replicas != *deployCur.Spec.Replicas {
			old := *deployCur.Spec.Replicas
			deployCur.Spec.Replicas = &cc.Replicas
			if err = r.client.Update(context.TODO(), deployCur); err == nil {
				r.log.Info("Scale  Deployment success.",
					"OldSize", old,
					"NewSize", cc.Replicas)
			}

		}
		return r.updateDeployment(c, deployCur, deploy)
	}

	r.log.Info("Node node num info",
		"Replicas", deployCur.Status.Replicas,
		"ReadyNum", deployCur.Status.ReadyReplicas,
		"CurrentNum", deployCur.Status.Replicas,
	)
	return
}

func (r *ReconcileDruid) reconcileConfigMap(cc *binaryomenv1alpha1.NodeSpec, c *binaryomenv1alpha1.Druid, cmCreate *v1.ConfigMap) (err error) {
	cmCur := &v1.ConfigMap{}
	err = r.client.Get(context.TODO(), types.NamespacedName{
		Name:      cmCreate.Name,
		Namespace: cmCreate.Namespace,
	}, cmCur)
	if err != nil && errors.IsNotFound(err) {
		if err = controllerutil.SetControllerReference(c, cmCreate, r.scheme); err != nil {
			return err
		}

		if err = r.client.Create(context.TODO(), cmCreate); err == nil {
			r.log.Info("Create  config map success",
				"ConfigMap.Namespace", c.Namespace,
				"ConfigMap.Name", cmCreate.GetName())
		}
	} else if err != nil {
		return err
	} else {
		if err = r.client.Update(context.TODO(), cmCur); err == nil {
			r.log.Info("Update configmap success")
		}
		return r.updateCm(c, cmCur, cmCreate)
	}
	return
}

func (r *ReconcileDruid) updateStatefulSet(c *binaryomenv1alpha1.Druid, foundSts *appsv1.StatefulSet, sts *appsv1.StatefulSet) (err error) {
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

func (r *ReconcileDruid) updateDeployment(c *binaryomenv1alpha1.Druid, foundDeploy *appsv1.Deployment, deploy *appsv1.Deployment) (err error) {
	r.log.Info("Updating StatefulSet",
		"StatefulSet.Namespace", foundDeploy.Namespace,
		"StatefulSet.Name", foundDeploy.Name)
	sync.SyncDeployment(foundDeploy, deploy)
	err = r.client.Update(context.TODO(), foundDeploy)
	if err != nil {
		return err
	}

	return nil
}

func (r *ReconcileDruid) updateCm(c *binaryomenv1alpha1.Druid, foundCm *v1.ConfigMap, cm *v1.ConfigMap) (err error) {
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
