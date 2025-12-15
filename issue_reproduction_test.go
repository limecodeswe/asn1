package asn1

import (
	"testing"
)

// Exact reproduction from the issue report
type LegIDFromIssue struct {
	LegType uint8 `asn1:"integer,tag:0"`
}

type BCSMEventFromIssue struct {
	EventTypeBCSM uint8           `asn1:"integer,tag:0"`
	MonitorMode   uint8           `asn1:"integer,tag:1"`
	LegID         *LegIDFromIssue `asn1:"choice,optional,tag:2"` // Optional choice field
}

func TestIssueReproduction(t *testing.T) {
	// Encode with LegID = nil (not present)
	orig := &BCSMEventFromIssue{
		EventTypeBCSM: 1,
		MonitorMode:   2,
		LegID:         nil, // Optional field not present
	}

	encoded, err := Marshal(orig)
	if err != nil {
		t.Fatalf("Encoding failed: %v", err)
	}

	// Decode - this was failing before the fix
	decoded := &BCSMEventFromIssue{}
	err = Unmarshal(encoded, decoded)
	if err != nil {
		t.Fatalf("Decoding failed: %v", err) // This was the error before fix
	}

	if decoded.LegID != nil {
		t.Error("Expected LegID to be nil")
	}

	// Verify other fields
	if decoded.EventTypeBCSM != 1 {
		t.Errorf("Expected EventTypeBCSM to be 1, got %d", decoded.EventTypeBCSM)
	}
	if decoded.MonitorMode != 2 {
		t.Errorf("Expected MonitorMode to be 2, got %d", decoded.MonitorMode)
	}

	t.Log("✅ Issue reproduction test passed - optional choice fields now work correctly!")
}
