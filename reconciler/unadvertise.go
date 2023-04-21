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
	a, ok, err := r.shouldUnadvertise(ctx, res)
	if err != nil {
		return reconcile.Result{}, err
	}

	if ok {
		advertised := res.Condition(crd.ConditionTypeAdvertised)

		cs, err := a.Unadvertise(ctx, res.Spec.ToDissolve())
		if err != nil {
			crd.ProviderError(
				r.Manager,
				res,
				res.Status.Provider,
				res.Status.ProviderDescription,
				err,
			)
			advertised = crd.UnadvertiseErrorCondition(err)
		} else if cs.IsEmpty() {
			advertised = crd.DNSRecordsDoNotExistCondition()
		} else {
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
			r.Logger.Info(
				"re-queueing",
				"namespace", res.Namespace,
				"name", res.Name,
				"reason", "potentially still advertised",
			)
			return reconcile.Result{Requeue: true}, nil
		}
	}

	controllerutil.RemoveFinalizer(res, crd.FinalizerName)
	if err := r.Client.Update(ctx, res); err != nil {
		return reconcile.Result{}, fmt.Errorf("unable to remove finalizer: %w", err)
	}

	r.Logger.Info(
		"removed finalizer",
		"namespace", res.Namespace,
		"name", res.Name,
		"finalizer", crd.FinalizerName,
	)

	return reconcile.Result{}, nil
}

func (r *Reconciler) shouldUnadvertise(
	ctx context.Context,
	res *crd.DNSSDServiceInstance,
) (adv provider.Advertiser, should bool, err error) {
	a := res.Condition(crd.ConditionTypeAdvertised)

	reason := ""

	if a.Status == metav1.ConditionFalse {
		should = false
		reason = "not advertised"
	} else {
		adv, should, err = r.getAdvertiser(ctx, res)
		if err != nil {
			return nil, false, err
		}
		if !should {
			reason = "unrecognized provider"
		} else if a.Status == metav1.ConditionTrue {
			should = true
			reason = "still advertised"
		} else if a.Status == metav1.ConditionUnknown {
			should = true
			reason = "potentially still advertised"
		}
	}

	message := "not unadvertising"
	if should {
		message = "unadvertising"
	}

	r.Logger.Info(
		message,
		"namespace", res.Namespace,
		"name", res.Name,
		"reason", reason,
	)

	return adv, should, nil
}
