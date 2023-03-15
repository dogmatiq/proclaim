package dnsimpleprovider

import (
	"github.com/dnsimple/dnsimple-go/dnsimple"
	"github.com/dogmatiq/proclaim/provider/dnsimpleprovider/internal/dnsimplex"
)

// changeSet encapsulates a set of DNS record changes that must be applied to
// reconcile the DNS zone with the desired state.
type changeSet struct {
	creates []dnsimple.ZoneRecordAttributes
	updates []struct {
		Before dnsimple.ZoneRecord
		After  dnsimple.ZoneRecordAttributes
	}
	deletes []dnsimple.ZoneRecord
}

func (cs *changeSet) Create(attr dnsimple.ZoneRecordAttributes) {
	cs.creates = append(cs.creates, attr)
}

func (cs *changeSet) Update(rec dnsimple.ZoneRecord, attr dnsimple.ZoneRecordAttributes) {
	if !dnsimplex.RecordHasAttributes(rec, attr) {
		cs.updates = append(
			cs.updates,
			struct {
				Before dnsimple.ZoneRecord
				After  dnsimple.ZoneRecordAttributes
			}{
				rec,
				attr,
			},
		)
	}
}

func (cs *changeSet) Delete(rec dnsimple.ZoneRecord) {
	cs.deletes = append(cs.deletes, rec)
}
