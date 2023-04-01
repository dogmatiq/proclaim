package crd

import (
	"github.com/dogmatiq/dissolve/dnssd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Resource is a DNS-SD resource that has a status.
type Resource interface {
	client.Object

	DissolveName() dnssd.ServiceInstanceName
	Status() *Status
}
