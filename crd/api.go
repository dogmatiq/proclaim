package crd

const (
	// GroupName is the API group name used by Proclaim.
	GroupName = "proclaim.dogmatiq.io"

	// FinalizerName is the name of the finalizer used by Proclaim to ensure
	// that DNS-SD services are unadvertised when they're underlying resource
	// is deleted.
	FinalizerName = GroupName + "/unadvertise"

	// Version is the version of the API/CRDs.
	Version = "v1"
)
