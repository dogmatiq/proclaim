package reconciler

import (
	"context"
	"fmt"

	"github.com/dogmatiq/proclaim/crd"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// unadvertise removes DNS records to ensure the given service instance is no
// longer advertised.
//
// It returns true on success. A non-nil error indicates context cancelation or
// a problem interacting with Kubernetes itself.
func (r *InstanceReconciler) unadvertise(
	ctx context.Context,
	res *crd.DNSSDServiceInstance,
) (reconcile.Result, error) {
	if res.Status().Provider != "" {
		a, ok, err := r.getAdvertiser(ctx, res)
		if !ok || err != nil {
			// The associated provider is not known to this reconciler.
			return reconcile.Result{}, err
		}

		advertised := res.Status().Condition(crd.ConditionTypeAdvertised)

		cs, err := a.UnadvertiseInstance(
			ctx,
			res.DissolveName(),
		)
		if err != nil {
			crd.ProviderError(
				r.Manager,
				res,
				res.Status().Provider,
				res.Status().ProviderDescription,
				err,
			)
			advertised = crd.UnadvertiseErrorCondition(err)
		} else if !cs.IsEmpty() {
			crd.DNSRecordsDeleted(r.Manager, res)
			advertised = crd.DNSRecordsDeletedCondition()
		}

		if err := r.update(
			res,
			crd.MergeCondition(advertised),
		); err != nil {
			return reconcile.Result{}, err
		}

		if advertised.Status != metav1.ConditionFalse {
			return reconcile.Result{Requeue: true}, nil
		}
	}

	controllerutil.RemoveFinalizer(res, crd.FinalizerName)
	if err := r.Client.Update(ctx, res); err != nil {
		return reconcile.Result{}, fmt.Errorf("unable to remove finalizer: %w", err)
	}

	return reconcile.Result{}, nil
}
