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
		// The proclaim finalizer has already been removed so we've already
		// unadvertised this service instance (if it was even necessary).
		return reconcile.Result{}, nil
	}

	if res.Status.ProviderID != "" {
		a, ok, err := r.getAdvertiser(ctx, res)
		if !ok || err != nil {
			// The associated provider is not known to this reconciler.
			return reconcile.Result{}, err
		}

		if res.Status.Status != crd.StatusUnadvertiseError {
			if err := r.updateStatus(
				ctx,
				res,
				func(s *crd.DNSSDServiceInstanceStatus) {
					s.Status = crd.StatusUnadvertising
				},
			); err != nil {
				return reconcile.Result{}, err
			}
		}

		result, err := a.Unadvertise(ctx, instanceFromSpec(res.Spec))

		switch result {
		case provider.UnadvertiseError:
			r.EventRecorder.Eventf(
				res,
				"Warning",
				"Error",
				"%s: %s",
				res.Status.ProviderDescription,
				err.Error(),
			)

			if err := r.updateStatus(
				ctx,
				res,
				func(s *crd.DNSSDServiceInstanceStatus) {
					s.Status = crd.StatusUnadvertiseError
				},
			); err != nil {
				return reconcile.Result{}, err
			}

			// Requeue for retry.
			return reconcile.Result{Requeue: true}, ctx.Err()

		case provider.InstanceNotAdvertised:
			// The service instance was not advertised, so we don't need to
			// push another event.

		case provider.UnadvertisedExistingInstance:
			r.EventRecorder.Eventf(
				res,
				"Normal",
				"Unadvertised",
				"service instance unadvertised successfully",
			)
		}
	}

	if err := r.updateStatus(
		ctx,
		res,
		func(s *crd.DNSSDServiceInstanceStatus) {
			s.Status = crd.StatusUnadvertised
		},
	); err != nil {
		return reconcile.Result{}, err
	}

	controllerutil.RemoveFinalizer(res, crd.FinalizerName)
	if err := r.Client.Update(ctx, res); err != nil {
		return reconcile.Result{}, fmt.Errorf("unable to remove finalizer: %w", err)
	}

	return reconcile.Result{}, nil
}
