package crd

import (
	"fmt"
	"strconv"
	"time"

	"github.com/dogmatiq/dissolve/dnssd"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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

// DNSSDServiceInstanceSpec is the specification for a service instance.
type DNSSDServiceInstanceSpec struct {
	Instance Instance `json:"instance"`
}

// ToDissolve returns a Dissolve dnssd.Instance from a CRD service instance
// specification.
func (s DNSSDServiceInstanceSpec) ToDissolve() dnssd.ServiceInstance {
	inst := dnssd.ServiceInstance{
		Name:        s.Instance.Name,
		ServiceType: s.Instance.ServiceType,
		Domain:      s.Instance.Domain,
		TargetHost:  s.Instance.Targets[0].Host,
		TargetPort:  s.Instance.Targets[0].Port,
		Priority:    s.Instance.Targets[0].Priority,
		Weight:      s.Instance.Targets[0].Weight,
		TTL:         s.Instance.TTL.Duration,
	}

	if inst.TTL == 0 {
		inst.TTL = 60 * time.Second
	}

	for _, src := range s.Instance.Attributes {
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
