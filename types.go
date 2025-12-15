// Package asn1 provides a complete ASN.1 library with BER encoding support.
package asn1

import (
	"fmt"
)

// ASN.1 Universal Class Tag Numbers
const (
	TagBoolean         = 1
	TagInteger         = 2
	TagBitString       = 3
	TagOctetString     = 4
	TagNull            = 5
	TagOID             = 6
	TagEnumerated      = 10
	TagUTF8String      = 12
	TagSequence        = 16
	TagSet             = 17
	TagPrintableString = 19
	TagIA5String       = 22
	TagUTCTime         = 23
	TagGeneralizedTime = 24
)

// Tag represents an ASN.1 tag
type Tag struct {
	Class       int  // 0=Universal, 1=Application, 2=Context-specific, 3=Private
	Constructed bool // true if constructed, false if primitive
	Number      int  // tag number
}

// ClassString returns the string representation of the tag class
func (t Tag) ClassString() string {
	switch t.Class {
	case 0:
		return "UNIVERSAL"
	case 1:
		return "APPLICATION"
	case 2:
		return "CONTEXT"
	case 3:
		return "PRIVATE"
	default:
		return fmt.Sprintf("CLASS_%d", t.Class)
	}
}

// TagString returns the tag in the format [CLASS NUMBER]
func (t Tag) TagString() string {
	return fmt.Sprintf("[%s %d]", t.ClassString(), t.Number)
}

// NewTag creates a new Tag
func NewTag(class int, constructed bool, number int) Tag {
	return Tag{
		Class:       class,
		Constructed: constructed,
		Number:      number,
	}
}

// NewUniversalTag creates a new universal class tag
func NewUniversalTag(number int, constructed bool) Tag {
	return Tag{
		Class:       0,
		Constructed: constructed,
		Number:      number,
	}
}

// NewContextSpecificTag creates a new context-specific tag
func NewContextSpecificTag(number int, constructed bool) Tag {
	return Tag{
		Class:       2,
		Constructed: constructed,
		Number:      number,
	}
}

// ASN1Marshaler is the interface implemented by types that can marshal themselves to ASN.1.
// The MarshalASN1 method should return the raw ASN.1-encoded bytes of the value,
// without any tag wrapping. The library will handle tag wrapping based on struct tags.
type ASN1Marshaler interface {
	MarshalASN1() ([]byte, error)
}

// ASN1Unmarshaler is the interface implemented by types that can unmarshal themselves from ASN.1.
// The UnmarshalASN1 method receives the raw ASN.1-encoded bytes of the value,
// without any tag wrapping. The library handles tag unwrapping before calling this method.
type ASN1Unmarshaler interface {
	UnmarshalASN1([]byte) error
}

// ASN1Object represents any ASN.1 object
type ASN1Object interface {
	// Encode returns the BER encoding of the object
	Encode() ([]byte, error)
	// String returns a string representation of the object
	String() string
	// Tag returns the ASN.1 tag of the object
	Tag() Tag
	// TaggedString returns a string representation with tag information
	TaggedString() string
}

// ASN1Value represents an ASN.1 object with its tag and value
type ASN1Value struct {
	tag   Tag
	value []byte
}

// NewASN1Value creates a new ASN1Value
func NewASN1Value(tag Tag, value []byte) *ASN1Value {
	copied := make([]byte, len(value))
	copy(copied, value)
	return &ASN1Value{
		tag:   tag,
		value: copied,
	}
}

// Tag returns the tag of the ASN1Value
func (v *ASN1Value) Tag() Tag {
	return v.tag
}

// Value returns the raw value bytes
func (v *ASN1Value) Value() []byte {
	result := make([]byte, len(v.value))
	copy(result, v.value)
	return result
}

// Encode returns the BER encoding of the ASN1Value
func (v *ASN1Value) Encode() ([]byte, error) {
	return EncodeTLV(v.tag, v.value)
}

// String returns a string representation of the ASN1Value
func (v *ASN1Value) String() string {
	return fmt.Sprintf("ASN1Value{tag: %v, value: %x}", v.tag, v.value)
}

// TaggedString returns a string representation with tag information
func (v *ASN1Value) TaggedString() string {
	return fmt.Sprintf("%s ASN1Value: %x", v.tag.TagString(), v.value)
}

// ASN1Structured represents structured ASN.1 types (SEQUENCE, SET)
type ASN1Structured struct {
	tag      Tag
	elements []ASN1Object
}

// NewSequence creates a new SEQUENCE
func NewSequence() *ASN1Structured {
	return &ASN1Structured{
		tag:      NewUniversalTag(TagSequence, true),
		elements: make([]ASN1Object, 0),
	}
}

// NewSet creates a new SET
func NewSet() *ASN1Structured {
	return &ASN1Structured{
		tag:      NewUniversalTag(TagSet, true),
		elements: make([]ASN1Object, 0),
	}
}

// NewStructured creates a new structured object with the given tag
func NewStructured(tag Tag) *ASN1Structured {
	return &ASN1Structured{
		tag:      tag,
		elements: make([]ASN1Object, 0),
	}
}

// Add adds an element to the structured object
func (s *ASN1Structured) Add(element ASN1Object) {
	s.elements = append(s.elements, element)
}

// Elements returns the elements of the structured object
func (s *ASN1Structured) Elements() []ASN1Object {
	result := make([]ASN1Object, len(s.elements))
	copy(result, s.elements)
	return result
}

// Tag returns the tag of the structured object
func (s *ASN1Structured) Tag() Tag {
	return s.tag
}

// Encode returns the BER encoding of the structured object
func (s *ASN1Structured) Encode() ([]byte, error) {
	var content []byte
	for _, element := range s.elements {
		encoded, err := element.Encode()
		if err != nil {
			return nil, fmt.Errorf("failed to encode element: %w", err)
		}
		content = append(content, encoded...)
	}
	return EncodeTLV(s.tag, content)
}

// String returns a string representation of the structured object
func (s *ASN1Structured) String() string {
	var typeName string
	if s.tag.Class == 0 {
		switch s.tag.Number {
		case TagSequence:
			typeName = "SEQUENCE"
		case TagSet:
			typeName = "SET"
		default:
			typeName = fmt.Sprintf("STRUCTURED[%d]", s.tag.Number)
		}
	} else {
		typeName = fmt.Sprintf("STRUCTURED[%d,%d]", s.tag.Class, s.tag.Number)
	}
	
	return fmt.Sprintf("%s (%d elements)", typeName, len(s.elements))
}

// TaggedString returns a string representation with tag information
func (s *ASN1Structured) TaggedString() string {
	var typeName string
	if s.tag.Class == 0 {
		switch s.tag.Number {
		case TagSequence:
			typeName = "SEQUENCE"
		case TagSet:
			typeName = "SET"
		default:
			typeName = fmt.Sprintf("STRUCTURED_%d", s.tag.Number)
		}
	} else {
		// For non-universal tags, just use "STRUCTURED" since the tag already contains class and number info
		typeName = "STRUCTURED"
	}
	
	return fmt.Sprintf("%s %s", s.tag.TagString(), typeName)
}

// CompactString returns a compact string representation showing the structure
func (s *ASN1Structured) CompactString() string {
	return s.compactStringWithIndent(0)
}

func (s *ASN1Structured) compactStringWithIndent(indent int) string {
	var typeName string
	if s.tag.Class == 0 {
		switch s.tag.Number {
		case TagSequence:
			typeName = "SEQUENCE"
		case TagSet:
			typeName = "SET"
		default:
			typeName = fmt.Sprintf("STRUCTURED_%d", s.tag.Number)
		}
	} else {
		// For non-universal tags, just use "STRUCTURED" since the tag already contains class and number info
		typeName = "STRUCTURED"
	}

	result := fmt.Sprintf("%s %s: {\n", s.tag.TagString(), typeName)
	for i, element := range s.elements {
		for j := 0; j <= indent; j++ {
			result += "  "
		}
		if structured, ok := element.(*ASN1Structured); ok {
			result += structured.compactStringWithIndent(indent + 1)
		} else {
			result += element.TaggedString()
		}
		if i < len(s.elements)-1 {
			result += ","
		}
		result += "\n"
	}
	for j := 0; j < indent; j++ {
		result += "  "
	}
	result += "}"
	return result
}