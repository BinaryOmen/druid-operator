package druid

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"

	binaryomenv1alpha1 "github.com/BinaryOmen/druid-operator/pkg/apis/binaryomen/v1alpha1"
	"github.com/BinaryOmen/druid-operator/pkg/validation"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_druid")

const ReconcileTime = 30 * time.Second
const druidFinalizer = "finalizer.druid.binaryomen.org"

// Add creates a new Druid Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileDruid{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("druid-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Druid
	err = c.Watch(&source.Kind{Type: &binaryomenv1alpha1.Druid{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource StatefulSet
	err = c.Watch(&source.Kind{Type: &appsv1.StatefulSet{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &binaryomenv1alpha1.Druid{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Deployment
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &binaryomenv1alpha1.Druid{},
	})

	// Watch for changes to secondary resource Service
	err = c.Watch(&source.Kind{Type: &v1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &binaryomenv1alpha1.Druid{},
	})

	// Watch for change to secondary resource configmap
	err = c.Watch(&source.Kind{Type: &v1.ConfigMap{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &binaryomenv1alpha1.Druid{},
	})

	if err != nil {
		return err
	}
	return nil
}

// blank assignment to verify that ReconcileDruid implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileDruid{}

// ReconcileDruid reconciles a Druid object
type ReconcileDruid struct {
	client client.Client
	scheme *runtime.Scheme
	log    logr.Logger
}

type reconcileFun func(cc *binaryomenv1alpha1.NodeSpec, c *binaryomenv1alpha1.Druid) error

func (r *ReconcileDruid) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	r.log = log.WithValues(
		"Request.Namespace", request.Namespace,
		"Request.Name", request.Name)
	r.log.Info("Reconciling DruidCluster")

	c := &binaryomenv1alpha1.Druid{}
	cc := &binaryomenv1alpha1.NodeSpec{}

	err := r.client.Get(context.TODO(), request.NamespacedName, c)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	// Validate Spec
	validator := validation.Validator{}
	validator.Validate(c)

	if !validator.Validated {
		e := fmt.Errorf("Failed to create Druid CR due to [%s]", validator.ErrorMessage)
		r.log.Error(e, e.Error(), "name", c.Name, "namespace", c.Namespace)
		return reconcile.Result{}, nil
	}

	// Reconcile
	for _, fun := range []reconcileFun{
		r.reconileDruid,
	} {
		if err = fun(cc, c); err != nil {
			return reconcile.Result{}, err
		}
	}

	isDruidMarkedToBeDeleted := c.GetDeletionTimestamp() != nil
	if isDruidMarkedToBeDeleted {
		if contains(c.GetFinalizers(), druidFinalizer) {
			// Run finalization logic for memcachedFinalizer. If the
			// finalization logic fails, don't remove the finalizer so
			// that we can retry during the next reconciliation.
			if err := r.finalizeDruid(c); err != nil {
				return reconcile.Result{}, err
			}

			// Remove memcachedFinalizer. Once all finalizers have been
			// removed, the object will be deleted.
			c.SetFinalizers(remove(c.GetFinalizers(), druidFinalizer))
			err := r.client.Update(context.TODO(), c)
			if err != nil {
				return reconcile.Result{}, err
			}
		}
		return reconcile.Result{}, nil
	}

	// Add finalizer for this CR
	if !contains(c.GetFinalizers(), druidFinalizer) {
		if err := r.addFinalizer(c); err != nil {
			return reconcile.Result{}, err
		}
	}

	// Recreate any missing resources every 'ReconcileTime'
	return reconcile.Result{RequeueAfter: ReconcileTime}, nil
}
func (r *ReconcileDruid) finalizeDruid(c *binaryomenv1alpha1.Druid) error {
	// TODO(user): Add the cleanup steps that the operator
	// needs to do before the CR can be deleted. Examples
	// of finalizers include performing backups and deleting
	// resources that are not owned by this CR, like a PVC.
	r.log.Info("Successfully finalized memcached")
	return nil
}

func (r *ReconcileDruid) addFinalizer(c *binaryomenv1alpha1.Druid) error {
	r.log.Info("Adding Finalizer for the Memcached")
	c.SetFinalizers(append(c.GetFinalizers(), druidFinalizer))

	// Update CR
	err := r.client.Update(context.TODO(), c)
	if err != nil {
		r.log.Error(err, "Failed to update Memcached with finalizer")
		return err
	}
	return nil
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func remove(list []string, s string) []string {
	for i, v := range list {
		if v == s {
			list = append(list[:i], list[i+1:]...)
		}
	}
	return list
}
