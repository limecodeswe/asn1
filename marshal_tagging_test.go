package asn1

import (
	"testing"
)

// TestImplicitTagging tests that context-specific tags use IMPLICIT tagging by default
func TestImplicitTagging(t *testing.T) {
	type Message struct {
		ID   int64  `asn1:"integer,tag:0"` // IMPLICIT by default
		Name string `asn1:"utf8string,tag:1"` // IMPLICIT by default
	}

	input := &Message{
		ID:   42,
		Name: "test",
	}

	// Marshal
	encoded, err := Marshal(input)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// Decode to check the structure
	decoded, _, err := DecodeTLV(encoded)
	if err != nil {
		t.Fatalf("DecodeTLV() error = %v", err)
	}

	// Should be a SEQUENCE
	if decoded.Tag().Class != 0 || decoded.Tag().Number != TagSequence {
		t.Fatalf("Expected SEQUENCE, got class=%d number=%d", decoded.Tag().Class, decoded.Tag().Number)
	}

	// Parse sequence elements
	content := decoded.Value()
	offset := 0

	// First element should be [0] IMPLICIT INTEGER (context-specific primitive)
	elem1, consumed, err := DecodeTLV(content[offset:])
	if err != nil {
		t.Fatalf("DecodeTLV(elem1) error = %v", err)
	}
	if elem1.Tag().Class != 2 { // Context-specific
		t.Errorf("elem1 class = %d, want 2 (context-specific)", elem1.Tag().Class)
	}
	if elem1.Tag().Number != 0 {
		t.Errorf("elem1 tag number = %d, want 0", elem1.Tag().Number)
	}
	if elem1.Tag().Constructed {
		t.Error("elem1 should be primitive (implicit tagging)")
	}
	offset += consumed

	// Second element should be [1] IMPLICIT UTF8String (context-specific primitive)
	elem2, _, err := DecodeTLV(content[offset:])
	if err != nil {
		t.Fatalf("DecodeTLV(elem2) error = %v", err)
	}
	if elem2.Tag().Class != 2 {
		t.Errorf("elem2 class = %d, want 2 (context-specific)", elem2.Tag().Class)
	}
	if elem2.Tag().Number != 1 {
		t.Errorf("elem2 tag number = %d, want 1", elem2.Tag().Number)
	}
	if elem2.Tag().Constructed {
		t.Error("elem2 should be primitive (implicit tagging)")
	}

	// Unmarshal and verify round-trip
	var output Message
	err = Unmarshal(encoded, &output)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if output.ID != input.ID {
		t.Errorf("ID = %d, want %d", output.ID, input.ID)
	}
	if output.Name != input.Name {
		t.Errorf("Name = %q, want %q", output.Name, input.Name)
	}
}

// TestExplicitTagging tests that the 'explicit' option enables EXPLICIT tagging
func TestExplicitTagging(t *testing.T) {
	type Message struct {
		ID   int64  `asn1:"integer,tag:0,explicit"` // EXPLICIT tagging
		Name string `asn1:"utf8string,tag:1,explicit"` // EXPLICIT tagging
	}

	input := &Message{
		ID:   42,
		Name: "test",
	}

	// Marshal
	encoded, err := Marshal(input)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// Decode to check the structure
	decoded, _, err := DecodeTLV(encoded)
	if err != nil {
		t.Fatalf("DecodeTLV() error = %v", err)
	}

	// Parse sequence elements
	content := decoded.Value()
	offset := 0

	// First element should be [0] EXPLICIT (context-specific constructed wrapper)
	elem1, consumed, err := DecodeTLV(content[offset:])
	if err != nil {
		t.Fatalf("DecodeTLV(elem1) error = %v", err)
	}
	if elem1.Tag().Class != 2 {
		t.Errorf("elem1 class = %d, want 2 (context-specific)", elem1.Tag().Class)
	}
	if elem1.Tag().Number != 0 {
		t.Errorf("elem1 tag number = %d, want 0", elem1.Tag().Number)
	}
	if !elem1.Tag().Constructed {
		t.Error("elem1 should be constructed (explicit tagging)")
	}

	// Inside should be the INTEGER
	inner1, _, err := DecodeTLV(elem1.Value())
	if err != nil {
		t.Fatalf("DecodeTLV(inner1) error = %v", err)
	}
	if inner1.Tag().Class != 0 || inner1.Tag().Number != TagInteger {
		t.Errorf("inner1 should be universal INTEGER, got class=%d number=%d", inner1.Tag().Class, inner1.Tag().Number)
	}
	offset += consumed

	// Second element should be [1] EXPLICIT
	elem2, _, err := DecodeTLV(content[offset:])
	if err != nil {
		t.Fatalf("DecodeTLV(elem2) error = %v", err)
	}
	if elem2.Tag().Class != 2 {
		t.Errorf("elem2 class = %d, want 2 (context-specific)", elem2.Tag().Class)
	}
	if elem2.Tag().Number != 1 {
		t.Errorf("elem2 tag number = %d, want 1", elem2.Tag().Number)
	}
	if !elem2.Tag().Constructed {
		t.Error("elem2 should be constructed (explicit tagging)")
	}

	// Unmarshal and verify round-trip
	var output Message
	err = Unmarshal(encoded, &output)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if output.ID != input.ID {
		t.Errorf("ID = %d, want %d", output.ID, input.ID)
	}
	if output.Name != input.Name {
		t.Errorf("Name = %q, want %q", output.Name, input.Name)
	}
}

// TestMixedTagging tests using both implicit and explicit tagging in the same struct
func TestMixedTagging(t *testing.T) {
	type Message struct {
		ImplicitInt  int64  `asn1:"integer,tag:0"`          // IMPLICIT (default)
		ExplicitInt  int64  `asn1:"integer,tag:1,explicit"` // EXPLICIT
		ImplicitStr  string `asn1:"utf8string,tag:2"`       // IMPLICIT (default)
		ExplicitStr  string `asn1:"utf8string,tag:3,explicit"` // EXPLICIT
	}

	input := &Message{
		ImplicitInt: 10,
		ExplicitInt: 20,
		ImplicitStr: "implicit",
		ExplicitStr: "explicit",
	}

	// Marshal
	encoded, err := Marshal(input)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// Unmarshal
	var output Message
	err = Unmarshal(encoded, &output)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	// Verify all fields
	if output.ImplicitInt != input.ImplicitInt {
		t.Errorf("ImplicitInt = %d, want %d", output.ImplicitInt, input.ImplicitInt)
	}
	if output.ExplicitInt != input.ExplicitInt {
		t.Errorf("ExplicitInt = %d, want %d", output.ExplicitInt, input.ExplicitInt)
	}
	if output.ImplicitStr != input.ImplicitStr {
		t.Errorf("ImplicitStr = %q, want %q", output.ImplicitStr, input.ImplicitStr)
	}
	if output.ExplicitStr != input.ExplicitStr {
		t.Errorf("ExplicitStr = %q, want %q", output.ExplicitStr, input.ExplicitStr)
	}
}

// TestImplicitSequence tests implicit tagging with SEQUENCE (should use implicit for constructed too)
func TestImplicitSequence(t *testing.T) {
	type Address struct {
		Street string `asn1:"utf8string"`
		City   string `asn1:"utf8string"`
	}

	type Person struct {
		Name    string   `asn1:"utf8string"`
		Address *Address `asn1:"sequence,tag:0"` // IMPLICIT SEQUENCE (context-specific constructed)
	}

	input := &Person{
		Name: "John",
		Address: &Address{
			Street: "123 Main St",
			City:   "Springfield",
		},
	}

	// Marshal
	encoded, err := Marshal(input)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// Decode to check structure
	decoded, _, err := DecodeTLV(encoded)
	if err != nil {
		t.Fatalf("DecodeTLV() error = %v", err)
	}

	// Parse sequence
	content := decoded.Value()
	offset := 0

	// Skip name field
	_, consumed, _ := DecodeTLV(content[offset:])
	offset += consumed

	// Address should be [0] IMPLICIT SEQUENCE (context-specific constructed with tag 0)
	addrElem, _, err := DecodeTLV(content[offset:])
	if err != nil {
		t.Fatalf("DecodeTLV(addr) error = %v", err)
	}
	if addrElem.Tag().Class != 2 {
		t.Errorf("address class = %d, want 2 (context-specific)", addrElem.Tag().Class)
	}
	if addrElem.Tag().Number != 0 {
		t.Errorf("address tag number = %d, want 0", addrElem.Tag().Number)
	}
	if !addrElem.Tag().Constructed {
		t.Error("address should be constructed")
	}

	// Unmarshal
	var output Person
	err = Unmarshal(encoded, &output)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	// Verify
	if output.Name != input.Name {
		t.Errorf("Name = %q, want %q", output.Name, input.Name)
	}
	if output.Address == nil {
		t.Fatal("Address is nil")
	}
	if output.Address.Street != input.Address.Street {
		t.Errorf("Street = %q, want %q", output.Address.Street, input.Address.Street)
	}
	if output.Address.City != input.Address.City {
		t.Errorf("City = %q, want %q", output.Address.City, input.Address.City)
	}
}

// TestExplicitSequence tests explicit tagging with SEQUENCE
func TestExplicitSequence(t *testing.T) {
	type Address struct {
		Street string `asn1:"utf8string"`
		City   string `asn1:"utf8string"`
	}

	type Person struct {
		Name    string   `asn1:"utf8string"`
		Address *Address `asn1:"sequence,tag:0,explicit"` // EXPLICIT SEQUENCE
	}

	input := &Person{
		Name: "John",
		Address: &Address{
			Street: "123 Main St",
			City:   "Springfield",
		},
	}

	// Marshal
	encoded, err := Marshal(input)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// Decode to check structure
	decoded, _, err := DecodeTLV(encoded)
	if err != nil {
		t.Fatalf("DecodeTLV() error = %v", err)
	}

	// Parse sequence
	content := decoded.Value()
	offset := 0

	// Skip name field
	_, consumed, _ := DecodeTLV(content[offset:])
	offset += consumed

	// Address should be [0] EXPLICIT wrapper
	wrapper, _, err := DecodeTLV(content[offset:])
	if err != nil {
		t.Fatalf("DecodeTLV(wrapper) error = %v", err)
	}
	if wrapper.Tag().Class != 2 {
		t.Errorf("wrapper class = %d, want 2", wrapper.Tag().Class)
	}
	if wrapper.Tag().Number != 0 {
		t.Errorf("wrapper tag = %d, want 0", wrapper.Tag().Number)
	}
	if !wrapper.Tag().Constructed {
		t.Error("wrapper should be constructed")
	}

	// Inside should be a universal SEQUENCE
	innerSeq, _, err := DecodeTLV(wrapper.Value())
	if err != nil {
		t.Fatalf("DecodeTLV(innerSeq) error = %v", err)
	}
	if innerSeq.Tag().Class != 0 || innerSeq.Tag().Number != TagSequence {
		t.Errorf("inner should be universal SEQUENCE, got class=%d number=%d", innerSeq.Tag().Class, innerSeq.Tag().Number)
	}

	// Unmarshal
	var output Person
	err = Unmarshal(encoded, &output)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	// Verify
	if output.Name != input.Name {
		t.Errorf("Name = %q, want %q", output.Name, input.Name)
	}
	if output.Address == nil {
		t.Fatal("Address is nil")
	}
	if output.Address.Street != input.Address.Street {
		t.Errorf("Street = %q, want %q", output.Address.Street, input.Address.Street)
	}
	if output.Address.City != input.Address.City {
		t.Errorf("City = %q, want %q", output.Address.City, input.Address.City)
	}
}
