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

const (
	broker        = "broker"
	coordinator   = "coordinator"
	overlord      = "overlord"
	middleManager = "middleManager"
	indexer       = "indexer"
	historical    = "historical"
	router        = "router"
)

type keyAndNodeSpec struct {
	key  string
	spec binaryomenv1alpha1.NodeSpec
}

func (r *ReconcileDruid) reconileDruid(cc *binaryomenv1alpha1.NodeSpec, c *binaryomenv1alpha1.Druid) error {

	for _, fun := range []reconcileFun{
		r.reconcileDruidNodes,
	} {
		if err := fun(cc, c); err != nil {
			r.log.Error(err, "Reconciling DruidCluster  Error", cc)
			return err
		}
	}

	return nil
}

// getAllNodeSpecsInDruidPrescribedOrder func shall initializes nodes[string]nodeSpec
func getAllNodeSpecsInDruidPrescribedOrder(c *binaryomenv1alpha1.Druid) ([]keyAndNodeSpec, error) {
	nodeSpecsByNodeType := map[string][]keyAndNodeSpec{
		historical:    make([]keyAndNodeSpec, 0, 1),
		overlord:      make([]keyAndNodeSpec, 0, 1),
		middleManager: make([]keyAndNodeSpec, 0, 1),
		indexer:       make([]keyAndNodeSpec, 0, 1),
		broker:        make([]keyAndNodeSpec, 0, 1),
		coordinator:   make([]keyAndNodeSpec, 0, 1),
		router:        make([]keyAndNodeSpec, 0, 1),
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
	allNodeSpecs = append(allNodeSpecs, nodeSpecsByNodeType[overlord]...)
	allNodeSpecs = append(allNodeSpecs, nodeSpecsByNodeType[middleManager]...)
	allNodeSpecs = append(allNodeSpecs, nodeSpecsByNodeType[indexer]...)
	allNodeSpecs = append(allNodeSpecs, nodeSpecsByNodeType[broker]...)
	allNodeSpecs = append(allNodeSpecs, nodeSpecsByNodeType[coordinator]...)
	allNodeSpecs = append(allNodeSpecs, nodeSpecsByNodeType[router]...)

	return allNodeSpecs, nil
}

// TODO: Deploy Cm, deployments, sts and Svc in order.
// TODO: Add running status
func (r *ReconcileDruid) reconcileDruidNodes(cc *binaryomenv1alpha1.NodeSpec, c *binaryomenv1alpha1.Druid) (err error) {
	allNodeSpecs, _ := getAllNodeSpecsInDruidPrescribedOrder(c)

	for _, elem := range allNodeSpecs {

		ns := elem.spec

		if ns.NodeType == historical || ns.NodeType == middleManager {
			sts := nodes.MakeStatefulSet(&ns, c)
			err = r.reconcileSts(&ns, c, sts)
			if err != nil {
				r.log.Error(err, "Reconciling Statefull Nodes  Error", cc)

			}
		}
		if ns.NodeType == overlord || ns.NodeType == router || ns.NodeType == broker || ns.NodeType == coordinator {
			d := nodes.MakeDeployment(&ns, c)
			err = r.reconcileDeployment(&ns, c, d)
			if err != nil {
				r.log.Error(err, "Reconciling Stateless Nodes  Error", cc)

			}
		}

		// TODO: Handle error
		cmN := nodes.MakeConfigMapNode(&ns, c)
		cmC := nodes.MakeConfigMapCommon(&ns, c)

		druidSvc := nodes.MakeService(&ns, c)
		r.reconcileService(&ns, c, druidSvc)

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

func (r *ReconcileDruid) reconcileDeployment(cc *binaryomenv1alpha1.NodeSpec, c *binaryomenv1alpha1.Druid, dmCreate *appsv1.Deployment) (err error) {
	dmCur := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), types.NamespacedName{
		Name:      dmCreate.Name,
		Namespace: dmCreate.Namespace,
	}, dmCur)
	if err != nil && errors.IsNotFound(err) {
		if err = controllerutil.SetControllerReference(c, dmCreate, r.scheme); err != nil {
			return err
		}

		if err = r.client.Create(context.TODO(), dmCreate); err == nil {
			r.log.Info("Create druid deployment success",
				"Deployment.Namespace", c.Namespace,
				"Deployment.Name", dmCreate.GetName())
		}
	} else if err != nil {
		return err
	} else {
		if cc.Replicas != *dmCur.Spec.Replicas {
			old := *dmCur.Spec.Replicas
			dmCur.Spec.Replicas = &cc.Replicas
			if err = r.client.Update(context.TODO(), dmCur); err == nil {
				r.log.Info("Scale druid deployment success",
					"OldSize", old,
					"NewSize", cc.Replicas)
			}
		}
	}
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

func (r *ReconcileDruid) reconcileService(cc *binaryomenv1alpha1.NodeSpec, c *binaryomenv1alpha1.Druid, svcCreate *v1.Service) (err error) {
	svcCur := &v1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{
		Name:      svcCreate.Name,
		Namespace: svcCreate.Namespace,
	}, svcCur)
	if err != nil && errors.IsNotFound(err) {
		if err = controllerutil.SetControllerReference(c, svcCreate, r.scheme); err != nil {
			return err
		}

		if err = r.client.Create(context.TODO(), svcCreate); err == nil {
			r.log.Info("Create  service success",
				"Service.Namespace", c.Namespace,
				"Service.Name", svcCreate.GetName())
		}
	} else if err != nil {
		return err
	} else {
		if err = r.client.Update(context.TODO(), svcCur); err == nil {
			r.log.Info("Update Service success")
		}
		return r.updateService(c, svcCur, svcCreate)
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

func (r *ReconcileDruid) updateService(c *binaryomenv1alpha1.Druid, foundSvc *v1.Service, svc *v1.Service) (err error) {
	r.log.Info("Updating Service",
		"Service.Namespace", foundSvc.Namespace,
		"Service.Name", foundSvc.Name)
	sync.SyncService(foundSvc, svc)
	err = r.client.Update(context.TODO(), foundSvc)
	if err != nil {
		return err
	}

	return nil
}
