package reconciliation

import (
	"errors"
	"testing"
	"time"

	"github.com/VanceMichael/go-base-airbridge/internal/domain"
	"github.com/VanceMichael/go-base-airbridge/internal/finance"
)

func TestFailedSettlementImportDoesNotLeavePartialLedgerEntries(t *testing.T) {
	now := time.Now().UTC()
	ledger := finance.New()
	if err := ledger.Post(finance.Entry{ID: "existing-credit", TenantID: "tenant-1", ShipmentID: "shipment-old", Currency: "CNY", Credit: 100, PostedAt: now.Add(-time.Hour)}); err != nil {
		t.Fatal(err)
	}
	before := ledger.Entries("tenant-1")
	beforeBalance := ledger.Balance("tenant-1")
	failed := []Entry{
		{ExternalID: "new-debit", Kind: "debit", Amount: 40, OccurredAt: now},
		{ExternalID: "existing-credit", Kind: "credit", Amount: 40, OccurredAt: now},
	}
	if err := ApplySettlement(ledger, "tenant-1", "shipment-1", "CNY", failed); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("settlement error = %v", err)
	}
	if entries := ledger.Entries("tenant-1"); len(entries) != len(before) || ledger.Balance("tenant-1") != beforeBalance {
		t.Fatalf("failed settlement changed ledger: entries=%+v balance=%d", entries, ledger.Balance("tenant-1"))
	}

	validLedger := finance.New()
	valid := []Entry{
		{ExternalID: "debit-1", Kind: "debit", Amount: 75, OccurredAt: now},
		{ExternalID: "credit-1", Kind: "credit", Amount: 75, OccurredAt: now.Add(time.Second)},
	}
	if err := ApplySettlement(validLedger, "tenant-1", "shipment-1", "CNY", valid); err != nil {
		t.Fatalf("valid settlement error = %v", err)
	}
	if len(validLedger.Entries("tenant-1")) != 2 || validLedger.Balance("tenant-1") != 0 {
		t.Fatalf("valid settlement entries=%+v balance=%d", validLedger.Entries("tenant-1"), validLedger.Balance("tenant-1"))
	}
}
