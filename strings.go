package asn1

import (
	"fmt"
	"unicode/utf8"
)

// ASN1UTF8String represents an ASN.1 UTF8String
type ASN1UTF8String struct {
	value string
}

// NewUTF8String creates a new ASN1UTF8String
func NewUTF8String(value string) *ASN1UTF8String {
	if !utf8.ValidString(value) {
		panic("invalid UTF-8 string")
	}
	return &ASN1UTF8String{value: value}
}

// Value returns the string value
func (s *ASN1UTF8String) Value() string {
	return s.value
}

// Tag returns the ASN.1 tag for UTF8String
func (s *ASN1UTF8String) Tag() Tag {
	return NewUniversalTag(TagUTF8String, false)
}

// Encode returns the BER encoding of the UTF8 string
func (s *ASN1UTF8String) Encode() ([]byte, error) {
	return EncodeTLV(s.Tag(), []byte(s.value))
}

// String returns a string representation
func (s *ASN1UTF8String) String() string {
	return fmt.Sprintf("UTF8String \"%s\"", s.value)
}

// ASN1PrintableString represents an ASN.1 PrintableString
type ASN1PrintableString struct {
	value string
}

// NewPrintableString creates a new ASN1PrintableString
func NewPrintableString(value string) *ASN1PrintableString {
	if !isPrintableString(value) {
		panic("string contains non-printable characters")
	}
	return &ASN1PrintableString{value: value}
}

// Value returns the string value
func (s *ASN1PrintableString) Value() string {
	return s.value
}

// Tag returns the ASN.1 tag for PrintableString
func (s *ASN1PrintableString) Tag() Tag {
	return NewUniversalTag(TagPrintableString, false)
}

// Encode returns the BER encoding of the printable string
func (s *ASN1PrintableString) Encode() ([]byte, error) {
	return EncodeTLV(s.Tag(), []byte(s.value))
}

// String returns a string representation
func (s *ASN1PrintableString) String() string {
	return fmt.Sprintf("PrintableString \"%s\"", s.value)
}

// ASN1IA5String represents an ASN.1 IA5String (ASCII string)
type ASN1IA5String struct {
	value string
}

// NewIA5String creates a new ASN1IA5String
func NewIA5String(value string) *ASN1IA5String {
	if !isIA5String(value) {
		panic("string contains non-IA5 characters")
	}
	return &ASN1IA5String{value: value}
}

// Value returns the string value
func (s *ASN1IA5String) Value() string {
	return s.value
}

// Tag returns the ASN.1 tag for IA5String
func (s *ASN1IA5String) Tag() Tag {
	return NewUniversalTag(TagIA5String, false)
}

// Encode returns the BER encoding of the IA5 string
func (s *ASN1IA5String) Encode() ([]byte, error) {
	return EncodeTLV(s.Tag(), []byte(s.value))
}

// String returns a string representation
func (s *ASN1IA5String) String() string {
	return fmt.Sprintf("IA5String \"%s\"", s.value)
}

// Helper functions for string validation

// isPrintableString checks if a string contains only PrintableString characters
func isPrintableString(s string) bool {
	for _, r := range s {
		if !isPrintableChar(r) {
			return false
		}
	}
	return true
}

// isPrintableChar checks if a rune is a valid PrintableString character
func isPrintableChar(r rune) bool {
	// PrintableString characters: A-Z, a-z, 0-9, space, and: ' ( ) + , - . / : = ?
	if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
		return true
	}
	switch r {
	case ' ', '\'', '(', ')', '+', ',', '-', '.', '/', ':', '=', '?':
		return true
	}
	return false
}

// isIA5String checks if a string contains only IA5String (ASCII) characters
func isIA5String(s string) bool {
	for _, r := range s {
		if r < 0 || r > 127 {
			return false
		}
	}
	return true
}

// DecodeUTF8String decodes an ASN1UTF8String from BER-encoded data
func DecodeUTF8String(data []byte) (*ASN1UTF8String, int, error) {
	asn1Value, consumed, err := DecodeTLV(data)
	if err != nil {
		return nil, 0, err
	}

	if asn1Value.tag.Class != 0 || asn1Value.tag.Number != TagUTF8String {
		return nil, 0, fmt.Errorf("expected UTF8String tag, got class=%d number=%d", asn1Value.tag.Class, asn1Value.tag.Number)
	}

	value := string(asn1Value.value)
	if !utf8.ValidString(value) {
		return nil, 0, fmt.Errorf("invalid UTF-8 string")
	}

	return NewUTF8String(value), consumed, nil
}

// DecodePrintableString decodes an ASN1PrintableString from BER-encoded data
func DecodePrintableString(data []byte) (*ASN1PrintableString, int, error) {
	asn1Value, consumed, err := DecodeTLV(data)
	if err != nil {
		return nil, 0, err
	}

	if asn1Value.tag.Class != 0 || asn1Value.tag.Number != TagPrintableString {
		return nil, 0, fmt.Errorf("expected PrintableString tag, got class=%d number=%d", asn1Value.tag.Class, asn1Value.tag.Number)
	}

	value := string(asn1Value.value)
	if !isPrintableString(value) {
		return nil, 0, fmt.Errorf("string contains non-printable characters")
	}

	return NewPrintableString(value), consumed, nil
}

// DecodeIA5String decodes an ASN1IA5String from BER-encoded data
func DecodeIA5String(data []byte) (*ASN1IA5String, int, error) {
	asn1Value, consumed, err := DecodeTLV(data)
	if err != nil {
		return nil, 0, err
	}

	if asn1Value.tag.Class != 0 || asn1Value.tag.Number != TagIA5String {
		return nil, 0, fmt.Errorf("expected IA5String tag, got class=%d number=%d", asn1Value.tag.Class, asn1Value.tag.Number)
	}

	value := string(asn1Value.value)
	if !isIA5String(value) {
		return nil, 0, fmt.Errorf("string contains non-IA5 characters")
	}

	return NewIA5String(value), consumed, nil
}