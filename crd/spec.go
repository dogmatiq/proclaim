package crd

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// Instance is a DNS-SD service instance.
type Instance struct {
	Name        string              `json:"name"`
	ServiceType string              `json:"serviceType"`
	Domain      string              `json:"domain"`
	TTL         metav1.Duration     `json:"ttl,omitempty"`
	Targets     [1]Target           `json:"targets"`
	Attributes  []map[string]string `json:"attributes,omitempty"`
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
