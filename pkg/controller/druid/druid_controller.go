package druid

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"

	binaryomenv1alpha1 "github.com/BinaryOmen/druid-operator/pkg/apis/binaryomen/v1alpha1"
	"github.com/BinaryOmen/druid-operator/pkg/controller/validation"
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

	// Watch for changes to secondary resource StatefulSet and requeue the owner Druid
	err = c.Watch(&source.Kind{Type: &appsv1.StatefulSet{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &binaryomenv1alpha1.Druid{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
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

	// Recreate any missing resources every 'ReconcileTime'
	return reconcile.Result{RequeueAfter: ReconcileTime}, nil
}
