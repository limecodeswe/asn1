package asn1

import (
	"fmt"
)

// ASN1Choice represents an ASN.1 CHOICE type that can hold one of several alternatives
type ASN1Choice struct {
	value    ASN1Object // The actual value chosen
	choiceID string     // Identifier for which choice was made (optional, for debugging)
}

// NewChoice creates a new CHOICE with the given value
func NewChoice(value ASN1Object) *ASN1Choice {
	return &ASN1Choice{
		value: value,
	}
}

// NewChoiceWithID creates a new CHOICE with the given value and choice identifier
func NewChoiceWithID(value ASN1Object, choiceID string) *ASN1Choice {
	return &ASN1Choice{
		value:    value,
		choiceID: choiceID,
	}
}

// Value returns the chosen value
func (c *ASN1Choice) Value() ASN1Object {
	return c.value
}

// ChoiceID returns the choice identifier (may be empty)
func (c *ASN1Choice) ChoiceID() string {
	return c.choiceID
}

// SetValue sets the chosen value
func (c *ASN1Choice) SetValue(value ASN1Object) {
	c.value = value
}

// SetChoiceID sets the choice identifier
func (c *ASN1Choice) SetChoiceID(choiceID string) {
	c.choiceID = choiceID
}

// Tag returns the tag of the chosen value
func (c *ASN1Choice) Tag() Tag {
	if c.value == nil {
		return Tag{}
	}
	return c.value.Tag()
}

// Encode returns the BER encoding of the chosen value
// Note: CHOICE is encoded as the chosen alternative directly
func (c *ASN1Choice) Encode() ([]byte, error) {
	if c.value == nil {
		return nil, fmt.Errorf("choice has no value set")
	}
	return c.value.Encode()
}

// String returns a string representation of the CHOICE
func (c *ASN1Choice) String() string {
	if c.value == nil {
		return "CHOICE{empty}"
	}
	if c.choiceID != "" {
		return fmt.Sprintf("CHOICE{%s: %s}", c.choiceID, c.value.String())
	}
	return fmt.Sprintf("CHOICE{%s}", c.value.String())
}

// DecodeChoice attempts to decode a CHOICE from the given data
// Since CHOICE is encoded as the chosen alternative directly, this function
// tries to match against a list of possible alternatives
func DecodeChoice(data []byte, alternatives []func([]byte) (ASN1Object, int, error)) (*ASN1Choice, int, error) {
	if len(data) == 0 {
		return nil, 0, fmt.Errorf("empty data")
	}

	var lastErr error
	for i, decoder := range alternatives {
		obj, consumed, err := decoder(data)
		if err == nil {
			choice := NewChoiceWithID(obj, fmt.Sprintf("alternative_%d", i))
			return choice, consumed, nil
		}
		lastErr = err
	}

	return nil, 0, fmt.Errorf("no alternative matched: %w", lastErr)
}

// DecodeChoiceWithTags attempts to decode a CHOICE from the given data
// by matching the tag against expected tags for each alternative
func DecodeChoiceWithTags(data []byte, expectedTags []Tag, decoders []func([]byte) (ASN1Object, int, error)) (*ASN1Choice, int, error) {
	if len(data) == 0 {
		return nil, 0, fmt.Errorf("empty data")
	}

	if len(expectedTags) != len(decoders) {
		return nil, 0, fmt.Errorf("number of expected tags must match number of decoders")
	}

	// Decode the tag from the data
	tag, _, err := DecodeTag(data)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to decode tag: %w", err)
	}

	// Find matching tag
	for i, expectedTag := range expectedTags {
		if tag.Class == expectedTag.Class && 
		   tag.Constructed == expectedTag.Constructed && 
		   tag.Number == expectedTag.Number {
			obj, consumed, err := decoders[i](data)
			if err != nil {
				return nil, 0, fmt.Errorf("failed to decode alternative %d: %w", i, err)
			}
			choice := NewChoiceWithID(obj, fmt.Sprintf("alternative_%d", i))
			return choice, consumed, nil
		}
	}

	return nil, 0, fmt.Errorf("no matching tag found for choice: got %+v", tag)
}