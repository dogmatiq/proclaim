package reconciler

import (
	"context"
	"fmt"

	"github.com/dogmatiq/proclaim/crd"
	"github.com/dogmatiq/proclaim/provider"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	advertised := res.Condition(crd.ConditionTypeAdvertised)

	if advertised.Status != metav1.ConditionFalse {
		a, ok, err := r.getAdvertiser(ctx, res)
		if !ok || err != nil {
			// The associated provider is not known to this reconciler.
			return reconcile.Result{}, err
		}

		result, err := a.Unadvertise(ctx, instanceFromSpec(res.Spec))

		switch result {
		case provider.UnadvertiseError:
			advertised = crd.AdvertisedConditionUnadvertiseError(err)
			r.EventRecorder.Eventf(
				res,
				"Warning",
				"Error",
				"%s: %s",
				res.Status.ProviderDescription,
				err.Error(),
			)
		case provider.InstanceNotAdvertised:
			// The service instance was not advertised, so we don't need to
			// push another event.

		case provider.UnadvertisedExistingInstance:
			advertised = crd.AdvertisedConditionRecordsRemoved()
			r.EventRecorder.Eventf(
				res,
				"Normal",
				"Unadvertised",
				"service instance unadvertised successfully",
			)
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
