package reconciler

import (
	"context"
	"fmt"
	"time"

	"github.com/dogmatiq/dissolve/dnssd"
	"github.com/dogmatiq/proclaim/crd"
	"github.com/dogmatiq/proclaim/provider"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// InstanceReconciler manipulates DNS records to match the state of a
// crd.DNSSDServiceInstance.
type InstanceReconciler struct {
	Manager   manager.Manager
	Client    client.Client
	Resolver  *dnssd.UnicastResolver
	Providers []provider.Provider
}

// Reconcile performs a full reconciliation for the object referred to by the
// Request, which must be a crd.DNSSDServiceInstance.
func (r *InstanceReconciler) Reconcile(
	ctx context.Context,
	req reconcile.Request,
) (reconcile.Result, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Lookup the resource so we know whether to advertise or unadvertise.
	res := &crd.DNSSDServiceInstance{}
	if err := r.Client.Get(ctx, req.NamespacedName, res); err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	if requeue, err := r.initialize(ctx, res); requeue || err != nil {
		return reconcile.Result{Requeue: true}, err
	}

	// Advertise the service, unless its deletion timestamp is set, in which
	// case we unadvertise it.
	if res.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.advertise(ctx, res)
	}
	return r.unadvertise(ctx, res)
}

func (r *InstanceReconciler) initialize(
	ctx context.Context,
	res *crd.DNSSDServiceInstance,
) (bool, error) {
	types := []string{
		crd.ConditionTypeAdopted,
		crd.ConditionTypeAdvertised,
		crd.ConditionTypeDiscoverable,
	}

	var updates []crd.StatusUpdate

	for _, t := range types {
		c := res.Status().Condition(crd.ConditionTypeAdvertised)
		if c.LastTransitionTime.IsZero() {
			updates = append(
				updates,
				crd.MergeCondition(
					metav1.Condition{
						Type:   t,
						Status: metav1.ConditionUnknown,
					},
				),
			)
		}
	}

	return len(updates) > 0, r.update(res, updates...)
}

func (r *InstanceReconciler) update(
	res *crd.DNSSDServiceInstance,
	updates ...crd.StatusUpdate,
) error {
	// Build our own context with a timeout, so that we don't block forever, but
	// nor do we fail if we're updating the status while shutting down due to a
	// higher-level context cancelation.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := crd.UpdateStatus(ctx, r.Client, res, updates...); err != nil {
		return fmt.Errorf("unable to update status sub-resource: %w", err)
	}

	return nil
}
