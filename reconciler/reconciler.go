package reconciler

import (
	"context"
	"fmt"
	"time"

	"github.com/dogmatiq/dissolve/dnssd"
	"github.com/dogmatiq/proclaim/crd"
	"github.com/dogmatiq/proclaim/provider"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Reconciler manipulates DNS records to match the state of a
// crd.DNSSDServiceInstance.
type Reconciler struct {
	Client        client.Client
	EventRecorder record.EventRecorder
	Providers     []provider.Provider
}

// Reconcile performs a full reconciliation for the object referred to by the
// Request, which must be a crd.DNSSDServiceInstance.
func (r *Reconciler) Reconcile(
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

	// Advertise the service, unless its deletion timestamp is set, in which
	// case we unadvertise it.
	op := r.advertise
	if !res.ObjectMeta.DeletionTimestamp.IsZero() {
		op = r.unadvertise
	}

	ok, err := op(
		ctx,
		res,
		instanceFromSpec(res.Spec),
	)

	// Requeue on failure. Note that we only return an error if there's a
	// problem interacting with kubernetes itself. If err == nil and ok == false
	// then there was an issue with at least one of the providers so we need to
	// requeue.
	return reconcile.Result{
		Requeue: !ok,
	}, err
}

func (r *Reconciler) setStatus(
	ctx context.Context,
	res *crd.DNSSDServiceInstance,
	status crd.Status,
) error {
	if res.Status.Status == status {
		return nil
	}

	res.Status.Status = status
	if err := r.Client.Status().Update(ctx, res); err != nil {
		return fmt.Errorf("unable to update resource status: %w", err)
	}

	return nil
}

// instanceFromSpec returns a dnssd.Instance from a CRD service instance
// specification.
func instanceFromSpec(spec crd.DNSSDServiceInstanceSpec) dnssd.ServiceInstance {
	result := dnssd.ServiceInstance{
		Instance:    spec.InstanceName,
		ServiceType: spec.ServiceType,
		Domain:      spec.Domain,
		TargetHost:  spec.TargetHost,
		TargetPort:  spec.TargetPort,
		Priority:    spec.Priority,
		Weight:      spec.Weight,
		TTL:         time.Duration(spec.TTL) * time.Second,
	}

	if result.TTL == 0 {
		result.TTL = 60 * time.Second
	}

	for _, src := range spec.Attributes {
		var dst dnssd.Attributes

		for k, v := range src {
			if v == "" {
				dst.SetFlag(k)
			} else {
				dst.Set(k, []byte(v))
			}
		}

		result.Attributes = append(result.Attributes, dst)
	}

	return result
}
