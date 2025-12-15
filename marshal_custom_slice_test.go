package asn1

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"testing"
)

// PhoneNumber - simple custom type for testing slice elements.
type PhoneNumber struct {
	Digits string
}

func (p PhoneNumber) MarshalASN1() ([]byte, error) {
	// Custom encoding: prefix with 0xAA
	return append([]byte{0xAA}, []byte(p.Digits)...), nil
}

func (p *PhoneNumber) UnmarshalASN1(data []byte) error {
	if len(data) < 1 || data[0] != 0xAA {
		return fmt.Errorf("invalid prefix")
	}
	p.Digits = string(data[1:])
	return nil
}

type Message struct {
	Numbers []PhoneNumber `asn1:"sequence,tag:0"`
}

// TestSliceOfCustomMarshalers tests that custom marshalers are invoked for slice elements
func TestSliceOfCustomMarshalers(t *testing.T) {
	msg := &Message{
		Numbers: []PhoneNumber{
			{Digits: "123"},
			{Digits: "456"},
		},
	}
	
	encoded, err := Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	
	t.Logf("Encoded: %s", hex.EncodeToString(encoded))
	
	// Check that the custom marshaler was called
	// Each element should start with 0xAA prefix
	encodedStr := hex.EncodeToString(encoded)
	if !bytes.Contains(encoded, []byte{0xAA, '1', '2', '3'}) {
		t.Errorf("Custom marshaler not invoked: expected 0xAA prefix for '123', got %s", encodedStr)
	}
	if !bytes.Contains(encoded, []byte{0xAA, '4', '5', '6'}) {
		t.Errorf("Custom marshaler not invoked: expected 0xAA prefix for '456', got %s", encodedStr)
	}
	
	// Decode
	decoded := &Message{}
	err = Unmarshal(encoded, decoded)
	if err != nil {
		t.Fatalf("Decoding failed: %v", err)
	}
	
	// Verify round-trip
	if len(decoded.Numbers) != 2 {
		t.Errorf("Expected 2 numbers, got %d", len(decoded.Numbers))
	}
	if decoded.Numbers[0].Digits != "123" {
		t.Errorf("First number mismatch: got %q, want %q", decoded.Numbers[0].Digits, "123")
	}
	if decoded.Numbers[1].Digits != "456" {
		t.Errorf("Second number mismatch: got %q, want %q", decoded.Numbers[1].Digits, "456")
	}
}

// TestSliceOfCustomMarshaler_TBCDRealWorld tests the real telecom use case
type TelecomMessage struct {
	ID      int64               `asn1:"integer"`
	Numbers []ISDNAddressString `asn1:"sequence,tag:0"`
	Name    string              `asn1:"utf8string"`
}

func TestSliceOfCustomMarshalers_TBCDRealWorld(t *testing.T) {
	msg := &TelecomMessage{
		ID: 42,
		Numbers: []ISDNAddressString{
			{
				Nature:        NatureInternational,
				NumberingPlan: NumberingE164,
				Digits:        "467011111",
			},
			{
				Nature:        NatureInternational,
				NumberingPlan: NumberingE164,
				Digits:        "123456789",
			},
		},
		Name: "TestCall",
	}
	
	encoded, err := Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	
	t.Logf("Encoded %d bytes: %02X", len(encoded), encoded)
	
	// Check that TBCD encoding is used (not ASCII)
	// ASCII "467011111" would be: 34 36 37 30 31 31 31 31 31
	asciiBytes := []byte("467011111")
	if bytes.Contains(encoded, asciiBytes) {
		t.Errorf("Custom marshaler not invoked: found ASCII encoding %02X instead of TBCD", asciiBytes)
	}
	
	// Expected TBCD for "467011111": should start with 0x11 (nature/plan byte)
	// followed by TBCD-encoded digits
	expectedTBCD, _ := encodeTBCD("467011111")
	expectedFirstByte := (byte(NatureInternational) << 4) | byte(NumberingE164)
	expectedEncoding := append([]byte{expectedFirstByte}, expectedTBCD...)
	
	if !bytes.Contains(encoded, expectedEncoding) {
		t.Errorf("Expected TBCD encoding %02X not found in output", expectedEncoding)
	}
	
	// Decode
	decoded := &TelecomMessage{}
	err = Unmarshal(encoded, decoded)
	if err != nil {
		t.Fatalf("Decoding failed: %v", err)
	}
	
	// Verify round-trip
	if decoded.ID != msg.ID {
		t.Errorf("ID mismatch: got %d, want %d", decoded.ID, msg.ID)
	}
	if len(decoded.Numbers) != 2 {
		t.Fatalf("Expected 2 numbers, got %d", len(decoded.Numbers))
	}
	if decoded.Numbers[0].Digits != "467011111" {
		t.Errorf("First number mismatch: got %q, want %q", decoded.Numbers[0].Digits, "467011111")
	}
	if decoded.Numbers[1].Digits != "123456789" {
		t.Errorf("Second number mismatch: got %q, want %q", decoded.Numbers[1].Digits, "123456789")
	}
	if decoded.Name != msg.Name {
		t.Errorf("Name mismatch: got %q, want %q", decoded.Name, msg.Name)
	}
}

// TestSliceOfPointerToCustomMarshaler tests slice of pointers to custom types
type PointerMessage struct {
	Numbers []*PhoneNumber `asn1:"sequence"`
}

func TestSliceOfPointerToCustomMarshaler(t *testing.T) {
	msg := &PointerMessage{
		Numbers: []*PhoneNumber{
			{Digits: "111"},
			{Digits: "222"},
		},
	}
	
	encoded, err := Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	
	t.Logf("Encoded: %s", hex.EncodeToString(encoded))
	
	// Check that the custom marshaler was called
	if !bytes.Contains(encoded, []byte{0xAA, '1', '1', '1'}) {
		t.Errorf("Custom marshaler not invoked for pointer element")
	}
	
	// Decode
	decoded := &PointerMessage{}
	err = Unmarshal(encoded, decoded)
	if err != nil {
		t.Fatalf("Decoding failed: %v", err)
	}
	
	// Verify round-trip
	if len(decoded.Numbers) != 2 {
		t.Errorf("Expected 2 numbers, got %d", len(decoded.Numbers))
	}
	if decoded.Numbers[0].Digits != "111" {
		t.Errorf("First number mismatch: got %q, want %q", decoded.Numbers[0].Digits, "111")
	}
}

// TestEmptySliceOfCustomMarshaler tests edge case of empty slice
func TestEmptySliceOfCustomMarshaler(t *testing.T) {
	msg := &Message{
		Numbers: []PhoneNumber{},
	}
	
	encoded, err := Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	
	// Decode
	decoded := &Message{}
	err = Unmarshal(encoded, decoded)
	if err != nil {
		t.Fatalf("Decoding failed: %v", err)
	}
	
	if len(decoded.Numbers) != 0 {
		t.Errorf("Expected empty slice, got %d elements", len(decoded.Numbers))
	}
}
