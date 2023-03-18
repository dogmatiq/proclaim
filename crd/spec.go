package crd

// Instance is a DNS-SD service instance.
type Instance struct {
	Name        string              `json:"name"`
	ServiceType string              `json:"serviceType"`
	Domain      string              `json:"domain"`
	TargetHost  string              `json:"targetHost"`
	TargetPort  uint16              `json:"targetPort"`
	Priority    uint16              `json:"priority,omitempty"`
	Weight      uint16              `json:"weight,omitempty"`
	TTL         uint16              `json:"ttl,omitempty"`
	Attributes  []map[string]string `json:"attributes,omitempty"`
}

// DNSSDServiceInstanceSpec is the specification for a service instance.
type DNSSDServiceInstanceSpec struct {
	Instance Instance `json:"instance"`
}
