package asn1

import (
	"bytes"
	"fmt"
	"testing"
)

// TestISDNAddressString simulates the telecom use case from the issue
// TBCD (Telephony Binary Coded Decimal) encoding for phone numbers

type NatureOfAddress byte
type NumberingPlan byte

const (
	NatureInternational NatureOfAddress = 0x01
	NumberingE164       NumberingPlan   = 0x01
)

// ISDNAddressString implements custom ASN.1 marshaling for TBCD-encoded phone numbers
type ISDNAddressString struct {
	Nature        NatureOfAddress
	NumberingPlan NumberingPlan
	Digits        string
}

// MarshalASN1 implements ASN1Marshaler for TBCD encoding
func (a *ISDNAddressString) MarshalASN1() ([]byte, error) {
	// Encode digits as TBCD (nibble-swapped BCD)
	tbcdDigits, err := encodeTBCD(a.Digits)
	if err != nil {
		return nil, err
	}
	
	// First byte contains nature and numbering plan
	firstByte := (byte(a.Nature) << 4) | byte(a.NumberingPlan)
	
	// Combine first byte with TBCD digits
	result := append([]byte{firstByte}, tbcdDigits...)
	return result, nil
}

// UnmarshalASN1 implements ASN1Unmarshaler for TBCD decoding
func (a *ISDNAddressString) UnmarshalASN1(data []byte) error {
	if len(data) < 1 {
		return fmt.Errorf("ISDN address too short")
	}
	
	// Extract nature and numbering plan from first byte
	a.Nature = NatureOfAddress((data[0] >> 4) & 0x07)
	a.NumberingPlan = NumberingPlan(data[0] & 0x0F)
	
	// Decode TBCD digits
	var err error
	a.Digits, err = decodeTBCD(data[1:])
	return err
}

// encodeTBCD converts decimal digits to TBCD format (nibble-swapped BCD)
func encodeTBCD(digits string) ([]byte, error) {
	// Calculate bytes needed: each byte stores 2 digits, so (len+1)/2 handles odd lengths
	result := make([]byte, (len(digits)+1)/2)
	
	for i := 0; i < len(digits); i++ {
		if digits[i] < '0' || digits[i] > '9' {
			return nil, fmt.Errorf("invalid digit: %c", digits[i])
		}
		
		digit := digits[i] - '0'
		byteIdx := i / 2
		
		if i%2 == 0 {
			// First digit goes in lower nibble
			result[byteIdx] = digit
		} else {
			// Second digit goes in upper nibble
			result[byteIdx] |= digit << 4
		}
	}
	
	// If odd number of digits, pad with 0xF
	if len(digits)%2 == 1 {
		result[len(result)-1] |= 0xF0
	}
	
	return result, nil
}

// decodeTBCD converts TBCD format to decimal digits
func decodeTBCD(data []byte) (string, error) {
	var result []byte
	
	for _, b := range data {
		// Lower nibble
		lowNibble := b & 0x0F
		if lowNibble <= 9 {
			result = append(result, '0'+lowNibble)
		} else if lowNibble == 0xF {
			// Padding, stop here
			break
		} else {
			return "", fmt.Errorf("invalid TBCD nibble: 0x%X", lowNibble)
		}
		
		// Upper nibble
		highNibble := (b >> 4) & 0x0F
		if highNibble <= 9 {
			result = append(result, '0'+highNibble)
		} else if highNibble == 0xF {
			// Padding, stop here
			break
		} else {
			return "", fmt.Errorf("invalid TBCD nibble: 0x%X", highNibble)
		}
	}
	
	return string(result), nil
}

// InitialDPArg is a simplified telecom struct using custom marshaled types
type InitialDPArg struct {
	ServiceKey         uint32            `asn1:"integer,tag:0"`
	CalledPartyNumber  ISDNAddressString `asn1:"octetstring,tag:2"`
	CallingPartyNumber ISDNAddressString `asn1:"octetstring,tag:3"`
	EventTypeBCSM      uint32            `asn1:"integer,tag:9"`
}

func TestCustomMarshalerISDN(t *testing.T) {
	// Test the telecom use case from the issue
	arg := InitialDPArg{
		ServiceKey: 123,
		CalledPartyNumber: ISDNAddressString{
			Nature:        NatureInternational,
			NumberingPlan: NumberingE164,
			Digits:        "12345678",
		},
		CallingPartyNumber: ISDNAddressString{
			Nature:        NatureInternational,
			NumberingPlan: NumberingE164,
			Digits:        "87654321",
		},
		EventTypeBCSM: 456,
	}
	
	// Marshal to ASN.1
	encoded, err := Marshal(&arg)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	
	t.Logf("Encoded %d bytes: %02X", len(encoded), encoded)
	
	// Unmarshal back
	var decoded InitialDPArg
	if err := Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	
	// Verify round-trip
	if decoded.ServiceKey != arg.ServiceKey {
		t.Errorf("ServiceKey mismatch: got %d, want %d", decoded.ServiceKey, arg.ServiceKey)
	}
	if decoded.CalledPartyNumber.Nature != arg.CalledPartyNumber.Nature {
		t.Errorf("CalledPartyNumber.Nature mismatch: got %v, want %v", 
			decoded.CalledPartyNumber.Nature, arg.CalledPartyNumber.Nature)
	}
	if decoded.CalledPartyNumber.NumberingPlan != arg.CalledPartyNumber.NumberingPlan {
		t.Errorf("CalledPartyNumber.NumberingPlan mismatch: got %v, want %v", 
			decoded.CalledPartyNumber.NumberingPlan, arg.CalledPartyNumber.NumberingPlan)
	}
	if decoded.CalledPartyNumber.Digits != arg.CalledPartyNumber.Digits {
		t.Errorf("CalledPartyNumber.Digits mismatch: got %q, want %q", 
			decoded.CalledPartyNumber.Digits, arg.CalledPartyNumber.Digits)
	}
	if decoded.CallingPartyNumber.Digits != arg.CallingPartyNumber.Digits {
		t.Errorf("CallingPartyNumber.Digits mismatch: got %q, want %q", 
			decoded.CallingPartyNumber.Digits, arg.CallingPartyNumber.Digits)
	}
	if decoded.EventTypeBCSM != arg.EventTypeBCSM {
		t.Errorf("EventTypeBCSM mismatch: got %d, want %d", decoded.EventTypeBCSM, arg.EventTypeBCSM)
	}
}

// Test simple custom marshaler with octetstring
type CustomBytes struct {
	data []byte
}

func (c *CustomBytes) MarshalASN1() ([]byte, error) {
	// Simple encoding: prepend length byte
	result := make([]byte, 1+len(c.data))
	result[0] = byte(len(c.data))
	copy(result[1:], c.data)
	return result, nil
}

func (c *CustomBytes) UnmarshalASN1(data []byte) error {
	if len(data) < 1 {
		return fmt.Errorf("data too short")
	}
	length := int(data[0])
	if len(data) < 1+length {
		return fmt.Errorf("data too short for declared length")
	}
	c.data = make([]byte, length)
	copy(c.data, data[1:1+length])
	return nil
}

type TestStructWithCustomBytes struct {
	Name   string      `asn1:"utf8string"`
	Custom CustomBytes `asn1:"octetstring"`
}

func TestCustomMarshalerOctetString(t *testing.T) {
	original := &TestStructWithCustomBytes{
		Name: "test",
		Custom: CustomBytes{
			data: []byte("hello world"),
		},
	}
	
	// Marshal
	encoded, err := Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	
	t.Logf("Encoded %d bytes", len(encoded))
	
	// Unmarshal
	var decoded TestStructWithCustomBytes
	if err := Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	
	// Verify
	if decoded.Name != original.Name {
		t.Errorf("Name mismatch: got %q, want %q", decoded.Name, original.Name)
	}
	if !bytes.Equal(decoded.Custom.data, original.Custom.data) {
		t.Errorf("Custom data mismatch: got %v, want %v", decoded.Custom.data, original.Custom.data)
	}
}

// Test custom marshaler with pointer receiver
type CustomInteger struct {
	value int64
}

func (c *CustomInteger) MarshalASN1() ([]byte, error) {
	// Encode as big-endian int64
	result := make([]byte, 8)
	val := c.value
	for i := 7; i >= 0; i-- {
		result[i] = byte(val & 0xFF)
		val >>= 8
	}
	// Trim leading zeros
	start := 0
	for start < len(result)-1 && result[start] == 0 {
		start++
	}
	return result[start:], nil
}

func (c *CustomInteger) UnmarshalASN1(data []byte) error {
	var val int64
	for _, b := range data {
		val = (val << 8) | int64(b)
	}
	c.value = val
	return nil
}

type TestStructWithCustomInt struct {
	ID     CustomInteger `asn1:"octetstring"`
	Name   string        `asn1:"utf8string"`
}

func TestCustomMarshalerPointerReceiver(t *testing.T) {
	original := &TestStructWithCustomInt{
		ID: CustomInteger{
			value: 123456789,
		},
		Name: "test",
	}
	
	// Marshal
	encoded, err := Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	
	t.Logf("Encoded %d bytes", len(encoded))
	
	// Unmarshal
	var decoded TestStructWithCustomInt
	if err := Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	
	// Verify
	if decoded.ID.value != original.ID.value {
		t.Errorf("ID mismatch: got %d, want %d", decoded.ID.value, original.ID.value)
	}
	if decoded.Name != original.Name {
		t.Errorf("Name mismatch: got %q, want %q", decoded.Name, original.Name)
	}
}

// Test mixed standard and custom types
type MixedStruct struct {
	StandardInt  int64             `asn1:"integer"`
	CustomField  ISDNAddressString `asn1:"octetstring"`
	StandardStr  string            `asn1:"utf8string"`
}

func TestMixedStandardAndCustom(t *testing.T) {
	original := &MixedStruct{
		StandardInt: 42,
		CustomField: ISDNAddressString{
			Nature:        NatureInternational,
			NumberingPlan: NumberingE164,
			Digits:        "555123456",
		},
		StandardStr: "test string",
	}
	
	// Marshal
	encoded, err := Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	
	t.Logf("Encoded %d bytes", len(encoded))
	
	// Unmarshal
	var decoded MixedStruct
	if err := Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	
	// Verify
	if decoded.StandardInt != original.StandardInt {
		t.Errorf("StandardInt mismatch: got %d, want %d", decoded.StandardInt, original.StandardInt)
	}
	if decoded.CustomField.Digits != original.CustomField.Digits {
		t.Errorf("CustomField.Digits mismatch: got %q, want %q", 
			decoded.CustomField.Digits, original.CustomField.Digits)
	}
	if decoded.StandardStr != original.StandardStr {
		t.Errorf("StandardStr mismatch: got %q, want %q", decoded.StandardStr, original.StandardStr)
	}
}

// Test TBCD encoding/decoding functions directly
func TestTBCDEncoding(t *testing.T) {
	tests := []struct {
		digits   string
		expected []byte
	}{
		{"12345678", []byte{0x21, 0x43, 0x65, 0x87}},
		{"1234567", []byte{0x21, 0x43, 0x65, 0xF7}},
		{"1", []byte{0xF1}},
		{"", []byte{}},
	}
	
	for _, test := range tests {
		t.Run(test.digits, func(t *testing.T) {
			encoded, err := encodeTBCD(test.digits)
			if err != nil {
				t.Fatalf("encodeTBCD failed: %v", err)
			}
			if !bytes.Equal(encoded, test.expected) {
				t.Errorf("Encoded mismatch: got %02X, want %02X", encoded, test.expected)
			}
			
			// Test round-trip
			decoded, err := decodeTBCD(encoded)
			if err != nil {
				t.Fatalf("decodeTBCD failed: %v", err)
			}
			if decoded != test.digits {
				t.Errorf("Decoded mismatch: got %q, want %q", decoded, test.digits)
			}
		})
	}
}

// Test error cases for custom marshaling
func TestCustomMarshalerErrors(t *testing.T) {
	// Test with invalid TBCD data
	invalid := ISDNAddressString{}
	err := invalid.UnmarshalASN1([]byte{})
	if err == nil {
		t.Error("Expected error for empty data, got nil")
	}
	
	// Test with invalid digits in encoding
	badDigits := ISDNAddressString{
		Nature:        NatureInternational,
		NumberingPlan: NumberingE164,
		Digits:        "123ABC",
	}
	_, err = badDigits.MarshalASN1()
	if err == nil {
		t.Error("Expected error for invalid digits, got nil")
	}
}

// Test custom types with optional fields
type StructWithOptionalCustom struct {
	Required       string             `asn1:"utf8string"`
	OptionalCustom *ISDNAddressString `asn1:"octetstring,optional,tag:0"`
	OptionalStd    *int64             `asn1:"integer,optional,tag:1"`
}

func TestOptionalCustomFields(t *testing.T) {
	// Test with both fields present
	phone := ISDNAddressString{
		Nature:        NatureInternational,
		NumberingPlan: NumberingE164,
		Digits:        "123456",
	}
	opt := int64(99)
	original := &StructWithOptionalCustom{
		Required:       "test",
		OptionalCustom: &phone,
		OptionalStd:    &opt,
	}
	
	encoded, err := Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	
	t.Logf("Encoded with both optional fields: %d bytes", len(encoded))
	
	var decoded StructWithOptionalCustom
	if err := Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	
	if decoded.Required != original.Required {
		t.Errorf("Required mismatch: got %q, want %q", decoded.Required, original.Required)
	}
	if decoded.OptionalCustom == nil {
		t.Error("OptionalCustom should not be nil")
	} else if decoded.OptionalCustom.Digits != original.OptionalCustom.Digits {
		t.Errorf("OptionalCustom.Digits mismatch: got %q, want %q", 
			decoded.OptionalCustom.Digits, original.OptionalCustom.Digits)
	}
	if decoded.OptionalStd == nil || *decoded.OptionalStd != *original.OptionalStd {
		t.Errorf("OptionalStd mismatch: got %v, want %v", decoded.OptionalStd, original.OptionalStd)
	}
	
	// Test with optional fields missing
	originalNoOpt := &StructWithOptionalCustom{
		Required: "test only",
	}
	
	encodedNoOpt, err := Marshal(originalNoOpt)
	if err != nil {
		t.Fatalf("Marshal without optional fields failed: %v", err)
	}
	
	t.Logf("Encoded without optional fields: %d bytes", len(encodedNoOpt))
	
	var decodedNoOpt StructWithOptionalCustom
	if err := Unmarshal(encodedNoOpt, &decodedNoOpt); err != nil {
		t.Fatalf("Unmarshal without optional fields failed: %v", err)
	}
	
	if decodedNoOpt.Required != originalNoOpt.Required {
		t.Errorf("Required mismatch: got %q, want %q", decodedNoOpt.Required, originalNoOpt.Required)
	}
	if decodedNoOpt.OptionalCustom != nil {
		t.Error("OptionalCustom should be nil")
	}
	if decodedNoOpt.OptionalStd != nil {
		t.Error("OptionalStd should be nil")
	}
}

// Test custom marshaling with explicit tagging
type StructWithExplicitCustom struct {
	ID         int64             `asn1:"integer"`
	CustomExpl ISDNAddressString `asn1:"octetstring,explicit,tag:5"`
	Name       string            `asn1:"utf8string"`
}

func TestCustomMarshalerExplicitTagging(t *testing.T) {
	original := &StructWithExplicitCustom{
		ID: 42,
		CustomExpl: ISDNAddressString{
			Nature:        NatureInternational,
			NumberingPlan: NumberingE164,
			Digits:        "999",
		},
		Name: "test",
	}
	
	encoded, err := Marshal(original)
	if err != nil {
		t.Fatalf("Marshal with explicit tagging failed: %v", err)
	}
	
	t.Logf("Encoded with explicit tagging: %d bytes: %02X", len(encoded), encoded)
	
	var decoded StructWithExplicitCustom
	if err := Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Unmarshal with explicit tagging failed: %v", err)
	}
	
	if decoded.ID != original.ID {
		t.Errorf("ID mismatch: got %d, want %d", decoded.ID, original.ID)
	}
	if decoded.CustomExpl.Digits != original.CustomExpl.Digits {
		t.Errorf("CustomExpl.Digits mismatch: got %q, want %q", 
			decoded.CustomExpl.Digits, original.CustomExpl.Digits)
	}
	if decoded.CustomExpl.Nature != original.CustomExpl.Nature {
		t.Errorf("CustomExpl.Nature mismatch: got %v, want %v", 
			decoded.CustomExpl.Nature, original.CustomExpl.Nature)
	}
	if decoded.Name != original.Name {
		t.Errorf("Name mismatch: got %q, want %q", decoded.Name, original.Name)
	}
}
