package asn1

import (
	"fmt"
	"strings"
)

// ASN1OctetString represents an ASN.1 OCTET STRING
type ASN1OctetString struct {
	value []byte
}

// NewOctetString creates a new ASN1OctetString
func NewOctetString(value []byte) *ASN1OctetString {
	// Create a copy to prevent external modification
	copied := make([]byte, len(value))
	copy(copied, value)
	return &ASN1OctetString{value: copied}
}

// NewOctetStringFromString creates a new ASN1OctetString from a string
func NewOctetStringFromString(value string) *ASN1OctetString {
	return NewOctetString([]byte(value))
}

// Value returns the octet string value
func (o *ASN1OctetString) Value() []byte {
	// Return a copy to prevent external modification
	result := make([]byte, len(o.value))
	copy(result, o.value)
	return result
}

// StringValue returns the octet string as a string
func (o *ASN1OctetString) StringValue() string {
	return string(o.value)
}

// Tag returns the ASN.1 tag for OCTET STRING
func (o *ASN1OctetString) Tag() Tag {
	return NewUniversalTag(TagOctetString, false)
}

// Encode returns the BER encoding of the octet string
func (o *ASN1OctetString) Encode() ([]byte, error) {
	return EncodeTLV(o.Tag(), o.value)
}

// String returns a string representation of the octet string
func (o *ASN1OctetString) String() string {
	// Try to display as a readable string if all bytes are printable
	if o.isPrintable() {
		return fmt.Sprintf("OCTET STRING \"%s\"", string(o.value))
	}
	
	// Otherwise, display as hex
	var hexParts []string
	for _, b := range o.value {
		hexParts = append(hexParts, fmt.Sprintf("%02X", b))
	}
	return fmt.Sprintf("OCTET STRING %s", strings.Join(hexParts, " "))
}

// CompactString returns a compact hex representation
func (o *ASN1OctetString) CompactString() string {
	return fmt.Sprintf("OCTET STRING (0x%x)", o.value)
}

// isPrintable checks if all bytes in the octet string are printable ASCII
func (o *ASN1OctetString) isPrintable() bool {
	for _, b := range o.value {
		if b < 32 || b > 126 {
			return false
		}
	}
	return true
}

// DecodeOctetString decodes an ASN1OctetString from BER-encoded data
func DecodeOctetString(data []byte) (*ASN1OctetString, int, error) {
	asn1Value, consumed, err := DecodeTLV(data)
	if err != nil {
		return nil, 0, err
	}

	if asn1Value.tag.Class != 0 || asn1Value.tag.Number != TagOctetString {
		return nil, 0, fmt.Errorf("expected OCTET STRING tag, got class=%d number=%d", asn1Value.tag.Class, asn1Value.tag.Number)
	}

	return NewOctetString(asn1Value.value), consumed, nil
}