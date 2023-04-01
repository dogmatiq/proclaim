package reconciler

import (
	"context"
	"fmt"
	"time"

	"github.com/dogmatiq/dissolve/dnssd"
	"github.com/dogmatiq/proclaim/crd"
	"github.com/dogmatiq/proclaim/provider"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Reconciler manipulates DNS records to match the state of a Proclaim CRD.
type Reconciler[T crd.Resource] struct {
	Manager   manager.Manager
	Client    client.Client
	Resolver  *dnssd.UnicastResolver
	Providers []provider.Provider
}

// Reconcile performs a full reconciliation for the object referred to by the
// Request, which must be a crd.DNSSDServiceInstance.
func (r *Reconciler[T]) Reconcile(
	ctx context.Context,
	req reconcile.Request,
) (reconcile.Result, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Lookup the resource so we know whether to advertise or unadvertise.
	ptr := new(T)
	res := any(ptr).(crd.Resource)
	if err := r.Client.Get(ctx, req.NamespacedName, res); err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	if requeue, err := crd.InitializeConditions(ctx, r.Client, res); requeue || err != nil {
		return reconcile.Result{Requeue: true}, err
	}

	// Advertise the service, unless its deletion timestamp is set, in which
	// case we unadvertise it.
	if res.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.advertise(ctx, res)
	}
	return r.unadvertise(ctx, res)
}

func (r *SubTypeReconciler) advertise(
	ctx context.Context,
	res *crd.DNSSDServiceInstanceSubType,
) (reconcile.Result, error) {
	if controllerutil.AddFinalizer(res, crd.FinalizerName) {
		if err := r.Client.Update(ctx, res); err != nil {
			return reconcile.Result{}, fmt.Errorf("unable to add finalizer: %w", err)
		}
	}

	if shouldAdvertise(res) {
		// if err := r.doAdvertise(ctx, res); err != nil {
		// 	return reconcile.Result{}, err
		// }
	}

	if shouldDiscover(res) {
		// ttl, err := r.doDiscover(ctx, res)
		// return shouldRequeue(res, ttl), err
	}

	return shouldRequeue(res, 0), nil
}

func (r *SubTypeReconciler) unadvertise(
	ctx context.Context,
	res *crd.DNSSDServiceInstanceSubType,
) (reconcile.Result, error) {
	if controllerutil.AddFinalizer(res, crd.FinalizerName) {
		if err := r.Client.Update(ctx, res); err != nil {
			return reconcile.Result{}, fmt.Errorf("unable to add finalizer: %w", err)
		}
	}

	if shouldAdvertise(res) {
		// if err := r.doAdvertise(ctx, res); err != nil {
		// 	return reconcile.Result{}, err
		// }
	}

	if shouldDiscover(res) {
		// ttl, err := r.doDiscover(ctx, res)
		// return shouldRequeue(res, ttl), err
	}

	return shouldRequeue(res, 0), nil
}
