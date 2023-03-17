package reconciler

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/dogmatiq/dissolve/dnssd"
	"github.com/dogmatiq/dyad"
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
	Resolver      *dnssd.UnicastResolver
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
	if res.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.advertise(ctx, res)
	}
	return r.unadvertise(ctx, res)
}

func (r *Reconciler) updateStatus(
	ctx context.Context,
	res *crd.DNSSDServiceInstance,
	fn func(*crd.DNSSDServiceInstanceStatus),
) error {
	updated := dyad.Clone(res.Status)
	fn(&updated)

	if reflect.DeepEqual(updated, res.Status) {
		return nil
	}

	res.Status = updated

	if err := r.Client.Status().Update(ctx, res); err != nil {
		return fmt.Errorf("unable to update status sub-resource: %w", err)
	}

	return nil
}

// instanceFromSpec returns a dnssd.Instance from a CRD service instance
// specification.
func instanceFromSpec(spec crd.DNSSDServiceInstanceSpec) dnssd.ServiceInstance {
	result := dnssd.ServiceInstance{
		Name:        spec.Instance.Name,
		ServiceType: spec.Instance.ServiceType,
		Domain:      spec.Instance.Domain,
		TargetHost:  spec.Instance.TargetHost,
		TargetPort:  spec.Instance.TargetPort,
		Priority:    spec.Instance.Priority,
		Weight:      spec.Instance.Weight,
		TTL:         time.Duration(spec.Instance.TTL) * time.Second,
	}

	if result.TTL == 0 {
		result.TTL = 60 * time.Second
	}

	for _, src := range spec.Instance.Attributes {
		var dst dnssd.Attributes

		for k, v := range src {
			if v == "" {
				dst = dst.WithFlag(k)
			} else {
				dst = dst.WithPair(k, []byte(v))
			}
		}

		result.Attributes = append(result.Attributes, dst)
	}

	return result
}
