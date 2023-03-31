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

// DNSSDServiceInstance is a resource that represents a DNS-SD service instance.
type DNSSDServiceInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec struct {
		Instance Instance `json:"instance"`
	} `json:"spec,omitempty"`
	Status Status `json:"status,omitempty"`
}

// DeepCopyObject returns a deep clone of i.
func (i *DNSSDServiceInstance) DeepCopyObject() runtime.Object {
	return dyad.Clone(i)
}

func (i *DNSSDServiceInstance) domain() string {
	return i.Spec.Instance.Domain
}

func (i *DNSSDServiceInstance) status() *Status {
	return &i.Status
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

// Instance is a DNS-SD service instance.
type Instance struct {
	Name        string           `json:"name"`
	ServiceType string           `json:"serviceType"`
	Domain      string           `json:"domain"`
	TTL         metav1.Duration  `json:"ttl,omitempty"`
	Targets     [1]Target        `json:"targets"`
	Attributes  []map[string]any `json:"attributes,omitempty"`
}

// Target describes a single target address for a DNS service instance.
type Target struct {
	Host     string `json:"host"`
	Port     uint16 `json:"port"`
	Priority uint16 `json:"priority,omitempty"`
	Weight   uint16 `json:"weight,omitempty"`
}

// ToDissolve returns a Dissolve dnssd.ServiceInstance from a CRD service
// instance.
func ToDissolve(i Instance) dnssd.ServiceInstance {
	inst := dnssd.ServiceInstance{
		Name:        i.Name,
		ServiceType: i.ServiceType,
		Domain:      i.Domain,
		TargetHost:  i.Targets[0].Host,
		TargetPort:  i.Targets[0].Port,
		Priority:    i.Targets[0].Priority,
		Weight:      i.Targets[0].Weight,
		TTL:         i.TTL.Duration,
	}

	if inst.TTL == 0 {
		inst.TTL = 60 * time.Second
	}

	for _, src := range i.Attributes {
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
