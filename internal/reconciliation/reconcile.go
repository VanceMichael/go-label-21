package reconciliation

import (
	"github.com/VanceMichael/go-base-airbridge/internal/domain"
	"github.com/VanceMichael/go-base-airbridge/internal/finance"
	"sort"
	"time"
)

type Entry struct {
	ID         string
	ExternalID string
	Kind       string
	Amount     int64
	OccurredAt time.Time
}
type Difference struct {
	ExternalID string
	Reason     string
	Local      *Entry
	Remote     *Entry
}

func ApplySettlement(ledger *finance.Ledger, tenantID, shipmentID, currency string, lines []Entry) error {
	if ledger == nil || tenantID == "" || shipmentID == "" || currency == "" || len(lines) == 0 {
		return domain.ErrInvalid
	}
	postings := make([]finance.Entry, 0, len(lines))
	for _, line := range lines {
		if err := Validate(line); err != nil || line.Amount <= 0 {
			return domain.ErrInvalid
		}
		posting := finance.Entry{ID: line.ExternalID, TenantID: tenantID, ShipmentID: shipmentID, Currency: currency, PostedAt: line.OccurredAt}
		switch line.Kind {
		case "debit":
			posting.Debit = line.Amount
		case "credit":
			posting.Credit = line.Amount
		default:
			return domain.ErrInvalid
		}
		postings = append(postings, posting)
	}
	return ledger.PostBatch(postings)
}

func Compare(local, remote []Entry) []Difference {
	lm := map[string]Entry{}
	rm := map[string]Entry{}
	for _, v := range local {
		lm[v.ExternalID] = v
	}
	for _, v := range remote {
		rm[v.ExternalID] = v
	}
	keys := make([]string, 0)
	for k := range lm {
		keys = append(keys, k)
	}
	for k := range rm {
		if _, ok := lm[k]; !ok {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	out := make([]Difference, 0)
	for _, k := range keys {
		l, lk := lm[k]
		r, rk := rm[k]
		switch {
		case !lk:
			out = append(out, Difference{ExternalID: k, Reason: "missing_local", Remote: &r})
		case !rk:
			out = append(out, Difference{ExternalID: k, Reason: "missing_remote", Local: &l})
		case l.Kind != r.Kind || l.Amount != r.Amount || !l.OccurredAt.Equal(r.OccurredAt):
			out = append(out, Difference{ExternalID: k, Reason: "different_values", Local: &l, Remote: &r})
		}
	}
	return out
}
func Validate(e Entry) error {
	if e.ExternalID == "" || e.Kind == "" || e.OccurredAt.IsZero() {
		return domain.ErrInvalid
	}
	return nil
}
