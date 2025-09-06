package asn1

import (
	"bytes"
	"testing"
)

func TestBoolean(t *testing.T) {
	tests := []struct {
		name  string
		value bool
		want  []byte
	}{
		{"true", true, []byte{0x01, 0x01, 0xFF}},
		{"false", false, []byte{0x01, 0x01, 0x00}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBoolean(tt.value)
			
			// Test encoding
			encoded, err := b.Encode()
			if err != nil {
				t.Fatalf("Encode() error = %v", err)
			}
			if !bytes.Equal(encoded, tt.want) {
				t.Errorf("Encode() = %x, want %x", encoded, tt.want)
			}

			// Test decoding
			decoded, consumed, err := DecodeBoolean(encoded)
			if err != nil {
				t.Fatalf("DecodeBoolean() error = %v", err)
			}
			if consumed != len(encoded) {
				t.Errorf("DecodeBoolean() consumed = %d, want %d", consumed, len(encoded))
			}
			if decoded.Value() != tt.value {
				t.Errorf("DecodeBoolean() value = %t, want %t", decoded.Value(), tt.value)
			}
		})
	}
}

func TestInteger(t *testing.T) {
	tests := []struct {
		name  string
		value int64
		want  []byte
	}{
		{"zero", 0, []byte{0x02, 0x01, 0x00}},
		{"positive", 127, []byte{0x02, 0x01, 0x7F}},
		{"large positive", 128, []byte{0x02, 0x02, 0x00, 0x80}},
		{"negative", -1, []byte{0x02, 0x01, 0xFF}},
		{"large negative", -128, []byte{0x02, 0x01, 0x80}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := NewInteger(tt.value)
			
			// Test encoding
			encoded, err := i.Encode()
			if err != nil {
				t.Fatalf("Encode() error = %v", err)
			}
			if !bytes.Equal(encoded, tt.want) {
				t.Errorf("Encode() = %x, want %x", encoded, tt.want)
			}

			// Test decoding
			decoded, consumed, err := DecodeInteger(encoded)
			if err != nil {
				t.Fatalf("DecodeInteger() error = %v", err)
			}
			if consumed != len(encoded) {
				t.Errorf("DecodeInteger() consumed = %d, want %d", consumed, len(encoded))
			}
			val, err := decoded.Int64()
			if err != nil {
				t.Fatalf("Int64() error = %v", err)
			}
			if val != tt.value {
				t.Errorf("DecodeInteger() value = %d, want %d", val, tt.value)
			}
		})
	}
}

func TestOctetString(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  []byte
	}{
		{"empty", "", []byte{0x04, 0x00}},
		{"hello", "hello", []byte{0x04, 0x05, 0x68, 0x65, 0x6C, 0x6C, 0x6F}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := NewOctetStringFromString(tt.value)
			
			// Test encoding
			encoded, err := o.Encode()
			if err != nil {
				t.Fatalf("Encode() error = %v", err)
			}
			if !bytes.Equal(encoded, tt.want) {
				t.Errorf("Encode() = %x, want %x", encoded, tt.want)
			}

			// Test decoding
			decoded, consumed, err := DecodeOctetString(encoded)
			if err != nil {
				t.Fatalf("DecodeOctetString() error = %v", err)
			}
			if consumed != len(encoded) {
				t.Errorf("DecodeOctetString() consumed = %d, want %d", consumed, len(encoded))
			}
			if decoded.StringValue() != tt.value {
				t.Errorf("DecodeOctetString() value = %q, want %q", decoded.StringValue(), tt.value)
			}
		})
	}
}

func TestNull(t *testing.T) {
	want := []byte{0x05, 0x00}
	
	n := NewNull()
	
	// Test encoding
	encoded, err := n.Encode()
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}
	if !bytes.Equal(encoded, want) {
		t.Errorf("Encode() = %x, want %x", encoded, want)
	}

	// Test decoding
	decoded, consumed, err := DecodeNull(encoded)
	if err != nil {
		t.Fatalf("DecodeNull() error = %v", err)
	}
	if consumed != len(encoded) {
		t.Errorf("DecodeNull() consumed = %d, want %d", consumed, len(encoded))
	}
	if decoded == nil {
		t.Error("DecodeNull() returned nil")
	}
}

func TestSequence(t *testing.T) {
	// Create a sequence with INTEGER(42) and BOOLEAN(true)
	seq := NewSequence()
	seq.Add(NewInteger(42))
	seq.Add(NewBoolean(true))
	
	// Test encoding
	encoded, err := seq.Encode()
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	// Expected: SEQUENCE { INTEGER 42, BOOLEAN true }
	// 30 06 02 01 2A 01 01 FF
	expected := []byte{0x30, 0x06, 0x02, 0x01, 0x2A, 0x01, 0x01, 0xFF}
	if !bytes.Equal(encoded, expected) {
		t.Errorf("Encode() = %x, want %x", encoded, expected)
	}

	// Test that we can decode it back to basic TLV structures
	decoded, consumed, err := DecodeTLV(encoded)
	if err != nil {
		t.Fatalf("DecodeTLV() error = %v", err)
	}
	if consumed != len(encoded) {
		t.Errorf("DecodeTLV() consumed = %d, want %d", consumed, len(encoded))
	}
	if decoded.tag.Number != TagSequence || !decoded.tag.Constructed {
		t.Errorf("DecodeTLV() tag = %v, want SEQUENCE", decoded.tag)
	}
}

func TestObjectIdentifier(t *testing.T) {
	tests := []struct {
		name       string
		components []int
		oidString  string
		want       []byte
	}{
		{
			name:       "1.2.840.113549",
			components: []int{1, 2, 840, 113549},
			oidString:  "1.2.840.113549",
			want:       []byte{0x06, 0x06, 0x2A, 0x86, 0x48, 0x86, 0xF7, 0x0D},
		},
		{
			name:       "2.5.4.3",
			components: []int{2, 5, 4, 3},
			oidString:  "2.5.4.3",
			want:       []byte{0x06, 0x03, 0x55, 0x04, 0x03},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test creation from components
			oid := NewObjectIdentifier(tt.components)
			
			// Test encoding
			encoded, err := oid.Encode()
			if err != nil {
				t.Fatalf("Encode() error = %v", err)
			}
			if !bytes.Equal(encoded, tt.want) {
				t.Errorf("Encode() = %x, want %x", encoded, tt.want)
			}

			// Test string representation
			if oid.DotNotation() != tt.oidString {
				t.Errorf("DotNotation() = %q, want %q", oid.DotNotation(), tt.oidString)
			}

			// Test creation from string
			oidFromString, err := NewObjectIdentifierFromString(tt.oidString)
			if err != nil {
				t.Fatalf("NewObjectIdentifierFromString() error = %v", err)
			}
			
			encodedFromString, err := oidFromString.Encode()
			if err != nil {
				t.Fatalf("Encode() from string error = %v", err)
			}
			if !bytes.Equal(encodedFromString, tt.want) {
				t.Errorf("Encode() from string = %x, want %x", encodedFromString, tt.want)
			}

			// Test decoding
			decoded, consumed, err := DecodeObjectIdentifier(encoded)
			if err != nil {
				t.Fatalf("DecodeObjectIdentifier() error = %v", err)
			}
			if consumed != len(encoded) {
				t.Errorf("DecodeObjectIdentifier() consumed = %d, want %d", consumed, len(encoded))
			}
			if decoded.DotNotation() != tt.oidString {
				t.Errorf("DecodeObjectIdentifier() OID = %q, want %q", decoded.DotNotation(), tt.oidString)
			}
		})
	}
}

func TestBitString(t *testing.T) {
	tests := []struct {
		name       string
		bits       string
		unusedBits int
		want       []byte
	}{
		{
			name:       "empty",
			bits:       "",
			unusedBits: 0,
			want:       []byte{0x03, 0x01, 0x00},
		},
		{
			name:       "1010",
			bits:       "1010",
			unusedBits: 4,
			want:       []byte{0x03, 0x02, 0x04, 0xA0},
		},
		{
			name:       "10101010",
			bits:       "10101010",
			unusedBits: 0,
			want:       []byte{0x03, 0x02, 0x00, 0xAA},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := NewBitStringFromBits(tt.bits)
			
			// Test unused bits calculation
			if bs.UnusedBits() != tt.unusedBits {
				t.Errorf("UnusedBits() = %d, want %d", bs.UnusedBits(), tt.unusedBits)
			}

			// Test bit string conversion
			if bs.ToBitString() != tt.bits {
				t.Errorf("ToBitString() = %q, want %q", bs.ToBitString(), tt.bits)
			}

			// Test encoding
			encoded, err := bs.Encode()
			if err != nil {
				t.Fatalf("Encode() error = %v", err)
			}
			if !bytes.Equal(encoded, tt.want) {
				t.Errorf("Encode() = %x, want %x", encoded, tt.want)
			}

			// Test decoding
			decoded, consumed, err := DecodeBitString(encoded)
			if err != nil {
				t.Fatalf("DecodeBitString() error = %v", err)
			}
			if consumed != len(encoded) {
				t.Errorf("DecodeBitString() consumed = %d, want %d", consumed, len(encoded))
			}
			if decoded.ToBitString() != tt.bits {
				t.Errorf("DecodeBitString() bits = %q, want %q", decoded.ToBitString(), tt.bits)
			}
		})
	}
}

func TestPerson(t *testing.T) {
	// Create a test person
	person := NewPerson(12345, "John Doe", "john.doe@example.com", true).
		SetDepartment("Engineering").
		SetPhoneNumber("+1-555-0123").
		SetSalary(75000).
		SetPermissions([]string{"read", "write", "admin"}).
		SetEmployeeType(0) // full-time

	// Test encoding
	encoded, err := person.Encode()
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	// Test that encoding produces some reasonable output
	if len(encoded) == 0 {
		t.Error("Encode() produced empty output")
	}

	// Test string representation
	str := person.String()
	if str == "" {
		t.Error("String() produced empty output")
	}
	
	// Test compact string representation
	compact := person.CompactString()
	if compact == "" {
		t.Error("CompactString() produced empty output")
	}

	t.Logf("Person encoded to %d bytes", len(encoded))
	t.Logf("Person string: %s", compact)
}

func TestPersonDirectory(t *testing.T) {
	// Create a test directory
	dir := NewPersonDirectory()
	
	person1 := NewPerson(1, "Alice", "alice@example.com", true).SetDepartment("Engineering")
	person2 := NewPerson(2, "Bob", "bob@example.com", false).SetDepartment("Sales")
	
	dir.AddPerson(person1)
	dir.AddPerson(person2)

	// Test encoding
	encoded, err := dir.Encode()
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	// Test that encoding produces some reasonable output
	if len(encoded) == 0 {
		t.Error("Encode() produced empty output")
	}

	// Test that we have the right number of persons
	persons := dir.Persons()
	if len(persons) != 2 {
		t.Errorf("Persons() returned %d persons, want 2", len(persons))
	}

	t.Logf("PersonDirectory encoded to %d bytes", len(encoded))
	t.Logf("PersonDirectory: %s", dir.CompactString())
}