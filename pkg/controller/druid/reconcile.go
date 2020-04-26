package druid

import (
	"context"
	"fmt"

	nodes "github.com/BinaryOmen/druid-operator/pkg/nodes"
	"github.com/BinaryOmen/druid-operator/pkg/sync"
	extensions "k8s.io/api/extensions/v1beta1"

	binaryomenv1alpha1 "github.com/BinaryOmen/druid-operator/pkg/apis/binaryomen/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/policy/v1beta1"
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

// TODO: Add running status
func (r *ReconcileDruid) reconcileDruidNodes(cc *binaryomenv1alpha1.NodeSpec, c *binaryomenv1alpha1.Druid) (err error) {
	allNodeSpecs, _ := getAllNodeSpecsInDruidPrescribedOrder(c)

	for _, elem := range allNodeSpecs {

		ns := elem.spec

		// create common properties configmap
		driuidCmRuntime := nodes.MakeConfigMapNode(&ns, c)
		err = r.reconcileConfigMap(&ns, c, driuidCmRuntime)
		if err != nil {
			r.log.Error(err, "Reconciling CM Runtime Properties Error", cc)
		}
		// create node runtime properties configmap
		druidCmCommon := nodes.MakeConfigMapCommon(&ns, c)
		err = r.reconcileConfigMap(&ns, c, druidCmCommon)
		if err != nil {
			r.log.Error(err, "Reconciling CM Common Properties Error", cc)
		}
		// create statefulsets for historicals and middlemanagers
		if ns.NodeType == historical || ns.NodeType == middleManager {
			sts := nodes.MakeStatefulSet(&ns, c)
			err = r.reconcileSts(&ns, c, sts)
			if err != nil {
				r.log.Error(err, "Reconciling Statefull Nodes Error", cc)

			}
		}
		// create deployments for overlord, router, broker and coordinator
		if ns.NodeType == overlord || ns.NodeType == router || ns.NodeType == broker || ns.NodeType == coordinator {
			d := nodes.MakeDeployment(&ns, c)
			err = r.reconcileDeployment(&ns, c, d)
			if err != nil {
				r.log.Error(err, "Reconciling Stateless Nodes Error", cc)

			}
		}
		// create druid service
		druidSvc := nodes.MakeService(&ns, c)
		err = r.reconcileService(&ns, c, druidSvc)
		if err != nil {
			r.log.Error(err, "Reconciling  Druid Service Error", cc)
		}

		// create ingress
		if ns.Ingress.Enabled == true {
			ing := nodes.MakeDruidIngress(&ns, c)
			err = r.reconcileIngress(&ns, c, ing)
			if err != nil {
				r.log.Error(err, "Reconcile Ingress Error", ing)
			}

		}
		// create poddisruptionbudget
		if ns.PodDisruptionBudget == true {
			pdb, err := nodes.MakePodDisruptionBudget(&ns, c)
			if err != nil {
				r.log.Error(err, "Making PDB Error", cc)
			}
			err = r.reconcilePdb(&ns, c, pdb)
			if err != nil {
				r.log.Error(err, "Reconciling Druid PDB Error", cc)
			}
		}

	}
	return
}

// reconcileSts will reconcile statefulsets
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
		return r.updateStatefulSet(ssCur, sts)
	}

	r.log.Info("Node node num info",
		"Replicas", ssCur.Status.Replicas,
		"ReadyNum", ssCur.Status.ReadyReplicas,
		"CurrentNum", ssCur.Status.CurrentReplicas,
	)
	return
}

// reconcileDeployment shall reconcile deployments
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
		return r.updateDeployment(c, dmCur, dmCreate)
	}
	return
}

// reconcileConfigMap shall reconcile all the common & runtime properties
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

// reconcileService shall reconcile druid svc's
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

// reconcilePdb shall reconcile pdb
func (r *ReconcileDruid) reconcilePdb(cc *binaryomenv1alpha1.NodeSpec, c *binaryomenv1alpha1.Druid, pdbCreate *v1beta1.PodDisruptionBudget) (err error) {
	pdbCur := &v1beta1.PodDisruptionBudget{}
	err = r.client.Get(context.TODO(), types.NamespacedName{
		Name:      pdbCreate.Name,
		Namespace: pdbCreate.Namespace,
	}, pdbCur)
	if err != nil && errors.IsNotFound(err) {
		if err = controllerutil.SetControllerReference(c, pdbCreate, r.scheme); err != nil {
			return err
		}

		if err = r.client.Create(context.TODO(), pdbCreate); err == nil {
			r.log.Info("Create  Pod Disruption Budget success",
				"Pdb.Namespace", c.Namespace,
				"Pdb.Name", pdbCreate.GetName())
		}
	} else if err != nil {
		return err
	} else {
		if err = r.client.Update(context.TODO(), pdbCur); err == nil {
			r.log.Info("Update Service success")
		}
	}
	return
}

// reconcileIngress shall reconcile ingress spec
func (r *ReconcileDruid) reconcileIngress(cc *binaryomenv1alpha1.NodeSpec, c *binaryomenv1alpha1.Druid, ingCreate *extensions.Ingress) (err error) {
	ingCur := &extensions.Ingress{}
	err = r.client.Get(context.TODO(), types.NamespacedName{
		Name:      ingCreate.Name,
		Namespace: ingCreate.Namespace,
	}, ingCur)
	if err != nil && errors.IsNotFound(err) {
		if err = controllerutil.SetControllerReference(c, ingCreate, r.scheme); err != nil {
			return err
		}

		if err = r.client.Create(context.TODO(), ingCreate); err == nil {
			r.log.Info("Create  Ingress success",
				"Ingress.Namespace", c.Namespace,
				"Ingress.Name", ingCreate.GetName())
		}
	} else if err != nil {
		return err
	} else {
		if err = r.client.Update(context.TODO(), ingCur); err == nil {
			r.log.Info("Update Ingress success")
		}
		return r.updateIng(c, ingCur, ingCreate)
	}
	return
}

// upateStatefulset shall sync fountsts with curr sts state
func (r *ReconcileDruid) updateStatefulSet(foundSts *appsv1.StatefulSet, sts *appsv1.StatefulSet) (err error) {
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

// updateDeployment shall sync foundedeploy with curr deployment state
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

// updateCm shall sync the common and runtime properties configmap
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

// updateService shall sync the service
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

// updateService shall sync the service
func (r *ReconcileDruid) updateIng(c *binaryomenv1alpha1.Druid, foundIng *extensions.Ingress, ing *extensions.Ingress) (err error) {
	r.log.Info("Updating Ingress",
		"Ingress.Namespace", foundIng.Namespace,
		"Ingress.Name", foundIng.Name)
	sync.SyncIngress(foundIng, ing)
	err = r.client.Update(context.TODO(), foundIng)
	if err != nil {
		return err
	}

	return nil
}

// https://github.com/druid-io/druid-operator/blob/0d843a4cd3b4aebfa13c2144ebdab2998f6de9e2/pkg/controller/druid/handler.go#L957
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
