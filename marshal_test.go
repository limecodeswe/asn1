package asn1

import (
	"bytes"
	"testing"
	"time"
)

// Test struct for basic marshaling
type TestStruct struct {
	ID       int64  `asn1:"integer"`
	Name     string `asn1:"utf8string"`
	IsActive bool   `asn1:"boolean"`
	Data     []byte `asn1:"octetstring"`
}

// Test struct with optional fields and context tags
type TestStructWithTags struct {
	Required  string     `asn1:"utf8string"`
	Optional1 *string    `asn1:"printablestring,optional,tag:0"`
	Optional2 *int64     `asn1:"integer,optional,tag:1"`
	Optional3 *time.Time `asn1:"utctime,optional,tag:2"`
}

// Test struct with slices
type TestStructWithSlices struct {
	Name    string   `asn1:"utf8string"`
	Numbers []int64  `asn1:"sequence"`
	Tags    []string `asn1:"sequence"`
}

func TestBasicMarshaling(t *testing.T) {
	original := &TestStruct{
		ID:       42,
		Name:     "Test Name",
		IsActive: true,
		Data:     []byte("test data"),
	}

	// Marshal to ASN.1
	encoded, err := Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	t.Logf("Encoded %d bytes", len(encoded))

	// Unmarshal back
	var decoded TestStruct
	if err := Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Verify round-trip
	if decoded.ID != original.ID {
		t.Errorf("ID mismatch: got %d, want %d", decoded.ID, original.ID)
	}
	if decoded.Name != original.Name {
		t.Errorf("Name mismatch: got %q, want %q", decoded.Name, original.Name)
	}
	if decoded.IsActive != original.IsActive {
		t.Errorf("IsActive mismatch: got %t, want %t", decoded.IsActive, original.IsActive)
	}
	if !bytes.Equal(decoded.Data, original.Data) {
		t.Errorf("Data mismatch: got %v, want %v", decoded.Data, original.Data)
	}
}

func TestOptionalFieldsMarshaling(t *testing.T) {
	optional1 := "optional value"
	optional2 := int64(123)
	optional3 := time.Date(2023, 12, 25, 14, 30, 0, 0, time.UTC)

	original := &TestStructWithTags{
		Required:  "required value",
		Optional1: &optional1,
		Optional2: &optional2,
		Optional3: &optional3,
	}

	// Marshal to ASN.1
	encoded, err := Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	t.Logf("Encoded %d bytes", len(encoded))

	// Unmarshal back
	var decoded TestStructWithTags
	if err := Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Verify round-trip
	if decoded.Required != original.Required {
		t.Errorf("Required mismatch: got %q, want %q", decoded.Required, original.Required)
	}
	if decoded.Optional1 == nil || *decoded.Optional1 != *original.Optional1 {
		t.Errorf("Optional1 mismatch: got %v, want %v", decoded.Optional1, original.Optional1)
	}
	if decoded.Optional2 == nil || *decoded.Optional2 != *original.Optional2 {
		t.Errorf("Optional2 mismatch: got %v, want %v", decoded.Optional2, original.Optional2)
	}
	if decoded.Optional3 == nil || !decoded.Optional3.Equal(*original.Optional3) {
		t.Errorf("Optional3 mismatch: got %v, want %v", decoded.Optional3, original.Optional3)
	}
}

func TestSliceMarshaling(t *testing.T) {
	original := &TestStructWithSlices{
		Name:    "Test with slices",
		Numbers: []int64{1, 2, 3, 42},
		Tags:    []string{"tag1", "tag2", "tag3"},
	}

	// Marshal to ASN.1
	encoded, err := Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	t.Logf("Encoded %d bytes", len(encoded))

	// Unmarshal back
	var decoded TestStructWithSlices
	if err := Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Verify round-trip
	if decoded.Name != original.Name {
		t.Errorf("Name mismatch: got %q, want %q", decoded.Name, original.Name)
	}
	if len(decoded.Numbers) != len(original.Numbers) {
		t.Fatalf("Numbers length mismatch: got %d, want %d", len(decoded.Numbers), len(original.Numbers))
	}
	for i, num := range original.Numbers {
		if decoded.Numbers[i] != num {
			t.Errorf("Numbers[%d] mismatch: got %d, want %d", i, decoded.Numbers[i], num)
		}
	}
	if len(decoded.Tags) != len(original.Tags) {
		t.Fatalf("Tags length mismatch: got %d, want %d", len(decoded.Tags), len(original.Tags))
	}
	for i, tag := range original.Tags {
		if decoded.Tags[i] != tag {
			t.Errorf("Tags[%d] mismatch: got %q, want %q", i, decoded.Tags[i], tag)
		}
	}
}

func TestEmptyOptionalFields(t *testing.T) {
	original := &TestStructWithTags{
		Required: "only required field",
		// All optional fields are nil
	}

	// Marshal to ASN.1
	encoded, err := Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	t.Logf("Encoded %d bytes", len(encoded))

	// Unmarshal back
	var decoded TestStructWithTags
	if err := Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Verify round-trip
	if decoded.Required != original.Required {
		t.Errorf("Required mismatch: got %q, want %q", decoded.Required, original.Required)
	}
	if decoded.Optional1 != nil {
		t.Errorf("Optional1 should be nil, got %v", decoded.Optional1)
	}
	if decoded.Optional2 != nil {
		t.Errorf("Optional2 should be nil, got %v", decoded.Optional2)
	}
	if decoded.Optional3 != nil {
		t.Errorf("Optional3 should be nil, got %v", decoded.Optional3)
	}
}

func TestTagParsing(t *testing.T) {
	tests := []struct {
		tag      string
		expected fieldInfo
		hasError bool
	}{
		{
			tag: "integer",
			expected: fieldInfo{
				Type:     "integer",
				Optional: false,
				HasTag:   false,
			},
		},
		{
			tag: "utf8string,optional,tag:5",
			expected: fieldInfo{
				Type:     "utf8string",
				Optional: true,
				HasTag:   true,
				Tag:      5,
			},
		},
		{
			tag: "boolean,omitempty",
			expected: fieldInfo{
				Type:      "boolean",
				Optional:  false,
				HasTag:    false,
				Omitempty: true,
			},
		},
		{
			tag:      "",
			hasError: true,
		},
	}

	for i, test := range tests {
		info, err := parseASN1Tag(test.tag)
		if test.hasError {
			if err == nil {
				t.Errorf("Test %d: expected error for tag %q", i, test.tag)
			}
			continue
		}
		if err != nil {
			t.Errorf("Test %d: unexpected error for tag %q: %v", i, test.tag, err)
			continue
		}

		if info.Type != test.expected.Type {
			t.Errorf("Test %d: Type mismatch: got %q, want %q", i, info.Type, test.expected.Type)
		}
		if info.Optional != test.expected.Optional {
			t.Errorf("Test %d: Optional mismatch: got %t, want %t", i, info.Optional, test.expected.Optional)
		}
		if info.HasTag != test.expected.HasTag {
			t.Errorf("Test %d: HasTag mismatch: got %t, want %t", i, info.HasTag, test.expected.HasTag)
		}
		if info.HasTag && info.Tag != test.expected.Tag {
			t.Errorf("Test %d: Tag mismatch: got %d, want %d", i, info.Tag, test.expected.Tag)
		}
		if info.Omitempty != test.expected.Omitempty {
			t.Errorf("Test %d: Omitempty mismatch: got %t, want %t", i, info.Omitempty, test.expected.Omitempty)
		}
	}
}

func TestComparisonWithManualEncoding(t *testing.T) {
	// Test that struct tag encoding produces the same result as manual encoding

	// Create a simple structure manually
	manual := NewSequence()
	manual.Add(NewInteger(42))
	manual.Add(NewUTF8String("test"))
	manual.Add(NewBoolean(true))
	manual.Add(NewOctetString([]byte("data")))

	manualEncoded, err := manual.Encode()
	if err != nil {
		t.Fatalf("Manual encoding failed: %v", err)
	}

	// Create the same structure using struct tags
	tagged := &TestStruct{
		ID:       42,
		Name:     "test",
		IsActive: true,
		Data:     []byte("data"),
	}

	taggedEncoded, err := Marshal(tagged)
	if err != nil {
		t.Fatalf("Struct tag encoding failed: %v", err)
	}

	// They should produce identical results
	if !bytes.Equal(manualEncoded, taggedEncoded) {
		t.Errorf("Manual and struct tag encoding produced different results")
		t.Logf("Manual:    %02X", manualEncoded)
		t.Logf("StructTag: %02X", taggedEncoded)
	}
}
