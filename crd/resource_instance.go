package crd

import (
	"fmt"
	"strconv"
	"time"

	"github.com/dogmatiq/dissolve/dnssd"
	"github.com/dogmatiq/dyad"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// DNSSDServiceInstance is a resource that represents a DNS-SD service instance
// to be published.
type DNSSDServiceInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Stat Status `json:"status,omitempty"`
	Spec struct {
		Instance struct {
			Name        string          `json:"name"`
			ServiceType string          `json:"serviceType"`
			Domain      string          `json:"domain"`
			TTL         metav1.Duration `json:"ttl,omitempty"`
			Targets     [1]struct {
				Host     string `json:"host"`
				Port     uint16 `json:"port"`
				Priority uint16 `json:"priority,omitempty"`
				Weight   uint16 `json:"weight,omitempty"`
			} `json:"targets"`
			Attributes []map[string]any `json:"attributes,omitempty"`
		} `json:"instance"`
	} `json:"spec,omitempty"`
}

// DeepCopyObject returns a deep clone of i.
func (r *DNSSDServiceInstance) DeepCopyObject() runtime.Object {
	return dyad.Clone(r)
}

// DissolveName returns a Dissolve dnssd.ServiceInstanceName from a CRD service
// instance.
func (r *DNSSDServiceInstance) DissolveName() dnssd.ServiceInstanceName {
	return dnssd.ServiceInstanceName{
		Name:        r.Spec.Instance.Name,
		ServiceType: r.Spec.Instance.ServiceType,
		Domain:      r.Spec.Instance.Domain,
	}
}

// DissolveInstance returns a Dissolve dnssd.ServiceInstance from a CRD service
// instance.
func (r *DNSSDServiceInstance) DissolveInstance() dnssd.ServiceInstance {
	inst := dnssd.ServiceInstance{
		ServiceInstanceName: dnssd.ServiceInstanceName{
			Name:        r.Spec.Instance.Name,
			ServiceType: r.Spec.Instance.ServiceType,
			Domain:      r.Spec.Instance.Domain,
		},
		TargetHost: r.Spec.Instance.Targets[0].Host,
		TargetPort: r.Spec.Instance.Targets[0].Port,
		Priority:   r.Spec.Instance.Targets[0].Priority,
		Weight:     r.Spec.Instance.Targets[0].Weight,
		TTL:        r.Spec.Instance.TTL.Duration,
	}

	if inst.TTL == 0 {
		inst.TTL = 60 * time.Second
	}

	for _, src := range r.Spec.Instance.Attributes {
		var dst dnssd.Attributes

		for k, v := range src {
			switch v := v.(type) {
			case bool:
				if v {
					dst = dst.WithFlag(k)
				}
			case string:
				dst = dst.WithPair(k, []byte(v))
			case int64:
				s := strconv.FormatInt(v, 10)
				dst = dst.WithPair(k, []byte(s))
			case float64:
				s := strconv.FormatFloat(v, 'g', -1, 64)
				dst = dst.WithPair(k, []byte(s))
			case nil:
				// ignore
			default:
				// TODO: A validating web-hook will make it so this branch
				// cannot be reached.
				panic(fmt.Sprintf("unsupported attribute value: %s = %T", k, v))
			}
		}

		if !dst.IsEmpty() {
			inst.Attributes = append(inst.Attributes, dst)
		}
	}

	return inst
}

// Status returns the status of the resource.
func (r *DNSSDServiceInstance) Status() *Status {
	return &r.Stat
}

// DNSSDServiceInstanceList is a list of DNS-SD service instances.
type DNSSDServiceInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []DNSSDServiceInstance `json:"items"`
}

// DeepCopyObject returns a deep clone of l.
func (l *DNSSDServiceInstanceList) DeepCopyObject() runtime.Object {
	return dyad.Clone(l)
}
