package asn1

import "fmt"

// ASN1Null represents an ASN.1 NULL
type ASN1Null struct{}

// NewNull creates a new ASN1Null
func NewNull() *ASN1Null {
	return &ASN1Null{}
}

// Tag returns the ASN.1 tag for NULL
func (n *ASN1Null) Tag() Tag {
	return NewUniversalTag(TagNull, false)
}

// Encode returns the BER encoding of the null value
func (n *ASN1Null) Encode() ([]byte, error) {
	// NULL has no content, only tag and length (which is 0)
	return EncodeTLV(n.Tag(), []byte{})
}

// String returns a string representation of the null value
func (n *ASN1Null) String() string {
	return "NULL"
}

// DecodeNull decodes an ASN1Null from BER-encoded data
func DecodeNull(data []byte) (*ASN1Null, int, error) {
	asn1Value, consumed, err := DecodeTLV(data)
	if err != nil {
		return nil, 0, err
	}

	if asn1Value.tag.Class != 0 || asn1Value.tag.Number != TagNull {
		return nil, 0, fmt.Errorf("expected NULL tag, got class=%d number=%d", asn1Value.tag.Class, asn1Value.tag.Number)
	}

	if len(asn1Value.value) != 0 {
		return nil, 0, fmt.Errorf("NULL value must be empty, got %d bytes", len(asn1Value.value))
	}

	return NewNull(), consumed, nil
}