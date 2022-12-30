package proclaim

import (
	"context"

	"github.com/dogmatiq/dapper"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type Reconciler struct {
	Client client.Client
	Scheme *runtime.Scheme
}

func (r *Reconciler) Reconcile(
	ctx context.Context,
	req reconcile.Request,
) (reconcile.Result, error) {
	dapper.Print(req)
	log := log.FromContext(ctx).WithValues("proclaim", req.NamespacedName)
	log.V(1).Info("reconciling DNS-SD instance")

	// inst := &Instance{}

	// if err := r.Client.Get(ctx, req.NamespacedName, inst); err != nil {
	// 	log.Error(err, "unable to get instance")
	// 	return ctrl.Result{}, err
	// }

	return ctrl.Result{}, nil
}
