package asn1

import (
	"bytes"
	"testing"
)

// Test for issue: Decoding fails for []byte fields with octetstring tag
// https://github.com/limecodeswe/asn1/issues/XXX

func TestOctetStringWithTag(t *testing.T) {
	type TestStruct struct {
		Data []byte `asn1:"octetstring,tag:0"`
	}

	testData := []byte{0x01, 0x02, 0x03, 0x04}
	original := TestStruct{
		Data: testData,
	}

	// Marshal to ASN.1
	encoded, err := Marshal(&original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	t.Logf("Encoded %d bytes: %02X", len(encoded), encoded)

	// Unmarshal back
	var decoded TestStruct
	if err := Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Verify round-trip
	if !bytes.Equal(decoded.Data, original.Data) {
		t.Errorf("Data mismatch: got %v, want %v", decoded.Data, original.Data)
	}
}

func TestOctetStringWithMultipleTags(t *testing.T) {
	type TestStruct struct {
		Data1 []byte `asn1:"octetstring,tag:0"`
		Data2 []byte `asn1:"octetstring,tag:1"`
		Data3 []byte `asn1:"octetstring,tag:2"`
	}

	original := TestStruct{
		Data1: []byte{0x01, 0x02},
		Data2: []byte{0x03, 0x04, 0x05},
		Data3: []byte{0x06},
	}

	// Marshal to ASN.1
	encoded, err := Marshal(&original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	t.Logf("Encoded %d bytes: %02X", len(encoded), encoded)

	// Unmarshal back
	var decoded TestStruct
	if err := Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Verify round-trip
	if !bytes.Equal(decoded.Data1, original.Data1) {
		t.Errorf("Data1 mismatch: got %v, want %v", decoded.Data1, original.Data1)
	}
	if !bytes.Equal(decoded.Data2, original.Data2) {
		t.Errorf("Data2 mismatch: got %v, want %v", decoded.Data2, original.Data2)
	}
	if !bytes.Equal(decoded.Data3, original.Data3) {
		t.Errorf("Data3 mismatch: got %v, want %v", decoded.Data3, original.Data3)
	}
}

func TestOctetStringWithOptionalTag(t *testing.T) {
	type TestStruct struct {
		Required []byte  `asn1:"octetstring"`
		Optional *[]byte `asn1:"octetstring,optional,tag:0"`
	}

	testData := []byte{0x01, 0x02, 0x03}
	optionalData := []byte{0x04, 0x05}

	// Test with optional field present
	t.Run("with_optional", func(t *testing.T) {
		original := TestStruct{
			Required: testData,
			Optional: &optionalData,
		}

		encoded, err := Marshal(&original)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		t.Logf("Encoded %d bytes: %02X", len(encoded), encoded)

		var decoded TestStruct
		if err := Unmarshal(encoded, &decoded); err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		if !bytes.Equal(decoded.Required, original.Required) {
			t.Errorf("Required mismatch: got %v, want %v", decoded.Required, original.Required)
		}
		if decoded.Optional == nil {
			t.Errorf("Optional is nil, expected %v", *original.Optional)
		} else if !bytes.Equal(*decoded.Optional, *original.Optional) {
			t.Errorf("Optional mismatch: got %v, want %v", *decoded.Optional, *original.Optional)
		}
	})

	// Test without optional field
	t.Run("without_optional", func(t *testing.T) {
		original := TestStruct{
			Required: testData,
			Optional: nil,
		}

		encoded, err := Marshal(&original)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		t.Logf("Encoded %d bytes: %02X", len(encoded), encoded)

		var decoded TestStruct
		if err := Unmarshal(encoded, &decoded); err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		if !bytes.Equal(decoded.Required, original.Required) {
			t.Errorf("Required mismatch: got %v, want %v", decoded.Required, original.Required)
		}
		if decoded.Optional != nil {
			t.Errorf("Optional is not nil, got %v", *decoded.Optional)
		}
	})
}

func TestOctetStringWithExplicitTag(t *testing.T) {
	type TestStruct struct {
		Data []byte `asn1:"octetstring,tag:0,explicit"`
	}

	testData := []byte{0x01, 0x02, 0x03, 0x04}
	original := TestStruct{
		Data: testData,
	}

	// Marshal to ASN.1
	encoded, err := Marshal(&original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	t.Logf("Encoded %d bytes: %02X", len(encoded), encoded)

	// Unmarshal back
	var decoded TestStruct
	if err := Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Verify round-trip
	if !bytes.Equal(decoded.Data, original.Data) {
		t.Errorf("Data mismatch: got %v, want %v", decoded.Data, original.Data)
	}
}

func TestEmptyOctetStringWithTag(t *testing.T) {
	type TestStruct struct {
		Data []byte `asn1:"octetstring,tag:0"`
	}

	original := TestStruct{
		Data: []byte{},
	}

	// Marshal to ASN.1
	encoded, err := Marshal(&original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	t.Logf("Encoded %d bytes: %02X", len(encoded), encoded)

	// Unmarshal back
	var decoded TestStruct
	if err := Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Verify round-trip
	if !bytes.Equal(decoded.Data, original.Data) {
		t.Errorf("Data mismatch: got %v, want %v", decoded.Data, original.Data)
	}
}
