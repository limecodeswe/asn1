package asn1

import "fmt"

// ASN1Boolean represents an ASN.1 BOOLEAN
type ASN1Boolean struct {
	value bool
}

// NewBoolean creates a new ASN1Boolean
func NewBoolean(value bool) *ASN1Boolean {
	return &ASN1Boolean{value: value}
}

// Value returns the boolean value
func (b *ASN1Boolean) Value() bool {
	return b.value
}

// Tag returns the ASN.1 tag for BOOLEAN
func (b *ASN1Boolean) Tag() Tag {
	return NewUniversalTag(TagBoolean, false)
}

// Encode returns the BER encoding of the boolean
func (b *ASN1Boolean) Encode() ([]byte, error) {
	var value []byte
	if b.value {
		value = []byte{0xFF}
	} else {
		value = []byte{0x00}
	}
	return EncodeTLV(b.Tag(), value)
}

// String returns a string representation of the boolean
func (b *ASN1Boolean) String() string {
	return fmt.Sprintf("BOOLEAN %t", b.value)
}

// TaggedString returns a string representation with tag information
func (b *ASN1Boolean) TaggedString() string {
	valueStr := "TRUE"
	if !b.value {
		valueStr = "FALSE"
	}
	return fmt.Sprintf("%s BOOLEAN: %s", b.Tag().TagString(), valueStr)
}

// DecodeBooleanValue decodes a boolean value from raw bytes
func DecodeBooleanValue(data []byte) (bool, error) {
	if len(data) != 1 {
		return false, fmt.Errorf("boolean value must be exactly 1 byte, got %d", len(data))
	}
	return data[0] != 0, nil
}

// DecodeBoolean decodes an ASN1Boolean from BER-encoded data
func DecodeBoolean(data []byte) (*ASN1Boolean, int, error) {
	asn1Value, consumed, err := DecodeTLV(data)
	if err != nil {
		return nil, 0, err
	}

	if asn1Value.tag.Class != 0 || asn1Value.tag.Number != TagBoolean {
		return nil, 0, fmt.Errorf("expected BOOLEAN tag, got class=%d number=%d", asn1Value.tag.Class, asn1Value.tag.Number)
	}

	value, err := DecodeBooleanValue(asn1Value.value)
	if err != nil {
		return nil, 0, err
	}

	return NewBoolean(value), consumed, nil
}