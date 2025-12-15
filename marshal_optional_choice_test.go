package asn1

import (
	"testing"
)

type LegID struct {
	LegType uint8 `asn1:"integer,tag:0"`
}

type BCSMEvent struct {
	EventTypeBCSM uint8  `asn1:"integer,tag:0"`
	MonitorMode   uint8  `asn1:"integer,tag:1"`
	LegID         *LegID `asn1:"sequence,optional,tag:2"` // Optional field
}

// BCSMEventWithChoice uses choice tag
type BCSMEventWithChoice struct {
	EventTypeBCSM uint8  `asn1:"integer,tag:0"`
	MonitorMode   uint8  `asn1:"integer,tag:1"`
	LegID         *LegID `asn1:"choice,optional,tag:2"` // Optional choice field
}

func TestOptionalChoice(t *testing.T) {
	// Encode with LegID = nil (not present)
	orig := &BCSMEvent{
		EventTypeBCSM: 1,
		MonitorMode:   2,
		LegID:         nil, // Optional field not present
	}

	encoded, err := Marshal(orig)
	if err != nil {
		t.Fatalf("Encoding failed: %v", err)
	}

	// Decode fails here
	decoded := &BCSMEvent{}
	err = Unmarshal(encoded, decoded)
	if err != nil {
		t.Fatalf("Decoding failed: %v", err) // Fails with type mismatch
	}

	if decoded.LegID != nil {
		t.Error("Expected LegID to be nil")
	}
}

func TestOptionalChoiceWithValue(t *testing.T) {
	// Encode with LegID present (using sequence tag since LegID is not a choice struct)
	orig := &BCSMEvent{
		EventTypeBCSM: 1,
		MonitorMode:   2,
		LegID:         &LegID{LegType: 5},
	}

	encoded, err := Marshal(orig)
	if err != nil {
		t.Fatalf("Encoding failed: %v", err)
	}

	// Decode
	decoded := &BCSMEvent{}
	err = Unmarshal(encoded, decoded)
	if err != nil {
		t.Fatalf("Decoding failed: %v", err)
	}

	if decoded.LegID == nil {
		t.Fatal("Expected LegID to be non-nil")
	}

	if decoded.LegID.LegType != 5 {
		t.Errorf("Expected LegType to be 5, got %d", decoded.LegID.LegType)
	}
}

// TestDecodeOptionalChoiceFromExternalData tests decoding data from an external source
// where the optional choice field is not present
func TestDecodeOptionalChoiceFromExternalData(t *testing.T) {
	// First encode without the optional field using the regular sequence tag
	orig := &BCSMEvent{
		EventTypeBCSM: 1,
		MonitorMode:   2,
		LegID:         nil,
	}

	encoded, err := Marshal(orig)
	if err != nil {
		t.Fatalf("Encoding failed: %v", err)
	}

	// Now try to decode using the choice variant
	decoded := &BCSMEventWithChoice{}
	err = Unmarshal(encoded, decoded)
	if err != nil {
		t.Fatalf("Decoding with choice tag failed: %v", err)
	}

	if decoded.LegID != nil {
		t.Error("Expected LegID to be nil")
	}
}

// Add a third field after the optional field to test tag skipping
type BCSMEventWithChoiceAndExtra struct {
	EventTypeBCSM uint8  `asn1:"integer,tag:0"`
	MonitorMode   uint8  `asn1:"integer,tag:1"`
	LegID         *LegID `asn1:"choice,optional,tag:2"` // Optional choice field
	ExtraField    uint8  `asn1:"integer,tag:3"`
}

// TestOptionalChoiceSkippingBug verifies that when an optional field is not present,
// the decoder correctly skips it and processes the next field
func TestOptionalChoiceSkippingBug(t *testing.T) {
	// Manually create encoded data with tags 0, 1, 3 (skipping tag 2)
	seq := NewSequence()

	// EventTypeBCSM - tag 0
	event := NewInteger(1)
	eventTagged := replaceTag(event, 0)
	seq.Add(eventTagged)

	// MonitorMode - tag 1
	monitor := NewInteger(2)
	monitorTagged := replaceTag(monitor, 1)
	seq.Add(monitorTagged)

	// ExtraField - tag 3 (skipping tag 2)
	extra := NewInteger(99)
	extraTagged := replaceTag(extra, 3)
	seq.Add(extraTagged)

	encoded, err := seq.Encode()
	if err != nil {
		t.Fatalf("Encoding failed: %v", err)
	}

	// Decode - this should correctly skip the optional field and decode ExtraField
	decoded := &BCSMEventWithChoiceAndExtra{}
	err = Unmarshal(encoded, decoded)
	if err != nil {
		t.Fatalf("Decoding failed: %v", err)
	}
	
	if decoded.EventTypeBCSM != 1 {
		t.Errorf("Expected EventTypeBCSM to be 1, got %d", decoded.EventTypeBCSM)
	}
	if decoded.MonitorMode != 2 {
		t.Errorf("Expected MonitorMode to be 2, got %d", decoded.MonitorMode)
	}
	if decoded.LegID != nil {
		t.Error("Expected LegID to be nil")
	}
	if decoded.ExtraField != 99 {
		t.Errorf("Expected ExtraField to be 99, got %d", decoded.ExtraField)
	}
}

// TestOptionalChoiceRoundTrip tests the exact reproduction case from issue #6
// This validates that optional choice fields work correctly for encoding and decoding
func TestOptionalChoiceRoundTrip(t *testing.T) {
	t.Run("nil optional choice field", func(t *testing.T) {
		// Encode with LegID = nil (not present)
		// Using BCSMEventWithChoice to test the choice tag specifically
		orig := &BCSMEventWithChoice{
			EventTypeBCSM: 1,
			MonitorMode:   2,
			LegID:         nil, // Optional field not present
		}

		encoded, err := Marshal(orig)
		if err != nil {
			t.Fatalf("Encoding failed: %v", err)
		}

		// Decode - this was failing with "expected ASN1Structured for struct, got *asn1.ASN1Value"
		decoded := &BCSMEventWithChoice{}
		err = Unmarshal(encoded, decoded)
		if err != nil {
			t.Fatalf("Decoding failed: %v", err)
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
	})

	t.Run("present optional field with sequence tag", func(t *testing.T) {
		// Using BCSMEvent with sequence tag since LegID is a regular struct
		orig := &BCSMEvent{
			EventTypeBCSM: 3,
			MonitorMode:   4,
			LegID:         &LegID{LegType: 5},
		}

		encoded, err := Marshal(orig)
		if err != nil {
			t.Fatalf("Encoding failed: %v", err)
		}

		// Decode
		decoded := &BCSMEvent{}
		err = Unmarshal(encoded, decoded)
		if err != nil {
			t.Fatalf("Decoding failed: %v", err)
		}

		if decoded.LegID == nil {
			t.Fatal("Expected LegID to be non-nil")
		}

		if decoded.LegID.LegType != 5 {
			t.Errorf("Expected LegType to be 5, got %d", decoded.LegID.LegType)
		}

		if decoded.EventTypeBCSM != 3 {
			t.Errorf("Expected EventTypeBCSM to be 3, got %d", decoded.EventTypeBCSM)
		}
		if decoded.MonitorMode != 4 {
			t.Errorf("Expected MonitorMode to be 4, got %d", decoded.MonitorMode)
		}
	})
}
