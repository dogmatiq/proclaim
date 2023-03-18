package reconciler

import (
	"context"
	"fmt"

	"github.com/dogmatiq/proclaim/crd"
	"github.com/dogmatiq/proclaim/provider"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// unadvertise removes DNS records to ensure the given service instance is no
// longer advertised.
//
// It returns true on success. A non-nil error indicates context cancelation or
// a problem interacting with Kubernetes itself.
func (r *Reconciler) unadvertise(
	ctx context.Context,
	res *crd.DNSSDServiceInstance,
) (reconcile.Result, error) {
	if !controllerutil.ContainsFinalizer(res, crd.FinalizerName) {
		return reconcile.Result{}, nil
	}

	advertised := res.Condition(crd.ConditionTypeAdvertised)

	if res.Status.ProviderID != "" {
		a, ok, err := r.getAdvertiser(ctx, res)
		if !ok || err != nil {
			// The associated provider is not known to this reconciler.
			return reconcile.Result{}, err
		}

		result, err := a.Unadvertise(ctx, instanceFromSpec(res.Spec))

		switch result {
		case provider.UnadvertisedExistingInstance:
			crd.DNSRecordsDeleted(r.Manager, res)
			advertised = crd.DNSRecordsDeletedCondition()

		case provider.InstanceNotAdvertised:
			// no change

		case provider.UnadvertiseError:
			crd.ProviderError(
				r.Manager,
				res,
				res.Status.ProviderID,
				res.Status.ProviderDescription,
				err,
			)
			advertised = crd.UnadvertiseErrorCondition(err)
		}

		if err := r.update(
			res,
			crd.MergeCondition(advertised),
		); err != nil {
			return reconcile.Result{}, err
		}
	}

	controllerutil.RemoveFinalizer(res, crd.FinalizerName)
	if err := r.Client.Update(ctx, res); err != nil {
		return reconcile.Result{}, fmt.Errorf("unable to remove finalizer: %w", err)
	}

	return reconcile.Result{}, nil
}
