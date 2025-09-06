package asn1

import (
	"fmt"
	"strconv"
	"strings"
)

// ASN1ObjectIdentifier represents an ASN.1 OBJECT IDENTIFIER
type ASN1ObjectIdentifier struct {
	components []int
}

// NewObjectIdentifier creates a new ASN1ObjectIdentifier
func NewObjectIdentifier(components []int) *ASN1ObjectIdentifier {
	if len(components) < 2 {
		panic("object identifier must have at least 2 components")
	}
	if components[0] < 0 || components[0] > 2 {
		panic("first component must be 0, 1, or 2")
	}
	if components[0] < 2 && (components[1] < 0 || components[1] > 39) {
		panic("second component must be 0-39 when first component is 0 or 1")
	}
	
	// Create a copy to prevent external modification
	copied := make([]int, len(components))
	copy(copied, components)
	return &ASN1ObjectIdentifier{components: copied}
}

// NewObjectIdentifierFromString creates a new ASN1ObjectIdentifier from a dot-separated string
func NewObjectIdentifierFromString(oid string) (*ASN1ObjectIdentifier, error) {
	parts := strings.Split(oid, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("object identifier must have at least 2 components")
	}
	
	components := make([]int, len(parts))
	for i, part := range parts {
		val, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("invalid component %q: %w", part, err)
		}
		if val < 0 {
			return nil, fmt.Errorf("component cannot be negative: %d", val)
		}
		components[i] = val
	}
	
	return NewObjectIdentifier(components), nil
}

// Components returns the OID components
func (o *ASN1ObjectIdentifier) Components() []int {
	// Return a copy to prevent external modification
	result := make([]int, len(o.components))
	copy(result, o.components)
	return result
}

// String returns the dot-separated string representation
func (o *ASN1ObjectIdentifier) String() string {
	parts := make([]string, len(o.components))
	for i, component := range o.components {
		parts[i] = strconv.Itoa(component)
	}
	return fmt.Sprintf("OBJECT IDENTIFIER %s", strings.Join(parts, "."))
}

// DotNotation returns just the dot-separated string without the type prefix
func (o *ASN1ObjectIdentifier) DotNotation() string {
	parts := make([]string, len(o.components))
	for i, component := range o.components {
		parts[i] = strconv.Itoa(component)
	}
	return strings.Join(parts, ".")
}

// Tag returns the ASN.1 tag for OBJECT IDENTIFIER
func (o *ASN1ObjectIdentifier) Tag() Tag {
	return NewUniversalTag(TagOID, false)
}

// Encode returns the BER encoding of the object identifier
func (o *ASN1ObjectIdentifier) Encode() ([]byte, error) {
	if len(o.components) < 2 {
		return nil, fmt.Errorf("object identifier must have at least 2 components")
	}
	
	var content []byte
	
	// First subidentifier combines the first two components
	firstSubid := o.components[0]*40 + o.components[1]
	content = append(content, encodeSubidentifier(firstSubid)...)
	
	// Remaining components are encoded individually
	for i := 2; i < len(o.components); i++ {
		content = append(content, encodeSubidentifier(o.components[i])...)
	}
	
	return EncodeTLV(o.Tag(), content)
}

// encodeSubidentifier encodes a single subidentifier using base-128 encoding
func encodeSubidentifier(value int) []byte {
	if value == 0 {
		return []byte{0}
	}
	
	var result []byte
	for value > 0 {
		result = append([]byte{byte(value & 0x7F)}, result...)
		value >>= 7
	}
	
	// Set the continuation bit (bit 7) for all bytes except the last
	for i := 0; i < len(result)-1; i++ {
		result[i] |= 0x80
	}
	
	return result
}

// DecodeObjectIdentifierValue decodes an object identifier value from raw bytes
func DecodeObjectIdentifierValue(data []byte) ([]int, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("object identifier value cannot be empty")
	}
	
	var components []int
	offset := 0
	
	// First subidentifier combines the first two components
	firstSubid, consumed, err := decodeSubidentifier(data[offset:])
	if err != nil {
		return nil, fmt.Errorf("failed to decode first subidentifier: %w", err)
	}
	offset += consumed
	
	// Split the first subidentifier into first two components
	if firstSubid < 40 {
		components = []int{0, firstSubid}
	} else if firstSubid < 80 {
		components = []int{1, firstSubid - 40}
	} else {
		components = []int{2, firstSubid - 80}
	}
	
	// Decode remaining subidentifiers
	for offset < len(data) {
		subid, consumed, err := decodeSubidentifier(data[offset:])
		if err != nil {
			return nil, fmt.Errorf("failed to decode subidentifier at offset %d: %w", offset, err)
		}
		components = append(components, subid)
		offset += consumed
	}
	
	return components, nil
}

// decodeSubidentifier decodes a single subidentifier from base-128 encoding
func decodeSubidentifier(data []byte) (int, int, error) {
	if len(data) == 0 {
		return 0, 0, fmt.Errorf("empty data")
	}
	
	value := 0
	offset := 0
	
	for {
		if offset >= len(data) {
			return 0, 0, fmt.Errorf("incomplete subidentifier")
		}
		
		b := data[offset]
		offset++
		
		// Check for overflow
		if value > (1<<(32-7))-1 {
			return 0, 0, fmt.Errorf("subidentifier too large")
		}
		
		value = (value << 7) | int(b&0x7F)
		
		// If bit 7 is not set, this is the last byte
		if (b & 0x80) == 0 {
			break
		}
	}
	
	return value, offset, nil
}

// DecodeObjectIdentifier decodes an ASN1ObjectIdentifier from BER-encoded data
func DecodeObjectIdentifier(data []byte) (*ASN1ObjectIdentifier, int, error) {
	asn1Value, consumed, err := DecodeTLV(data)
	if err != nil {
		return nil, 0, err
	}

	if asn1Value.tag.Class != 0 || asn1Value.tag.Number != TagOID {
		return nil, 0, fmt.Errorf("expected OBJECT IDENTIFIER tag, got class=%d number=%d", asn1Value.tag.Class, asn1Value.tag.Number)
	}

	components, err := DecodeObjectIdentifierValue(asn1Value.value)
	if err != nil {
		return nil, 0, err
	}

	return NewObjectIdentifier(components), consumed, nil
}