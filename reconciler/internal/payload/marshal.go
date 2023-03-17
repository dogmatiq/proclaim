package payload

import (
	"fmt"

	"github.com/dogmatiq/proclaim/provider"
	"google.golang.org/protobuf/proto"
)

const version = 1

// Marshal returns a payload that can be used to identify the given provider
// and advertiser.
func Marshal(p provider.Provider, a provider.Advertiser) []byte {
	data, err := proto.Marshal(
		&Payload{
			Version:      version,
			ProviderId:   p.ID(),
			AdvertiserId: a.ID(),
		},
	)
	if err != nil {
		panic(err)
	}

	return data
}

// Unmarshal returns the payload represented by the given data.
func Unmarshal(data []byte) (*Payload, error) {
	var p Payload
	if err := proto.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("unable to unmarshal payload: %w", err)
	}

	if p.GetVersion() != version {
		return nil, fmt.Errorf("unsupported payload version %d", p.GetVersion())
	}

	return &p, nil
}
