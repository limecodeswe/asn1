package asn1

import (
	"fmt"
	"io"
)

// EncodeTLV encodes a Tag-Length-Value structure using BER rules
func EncodeTLV(tag Tag, value []byte) ([]byte, error) {
	var result []byte

	// Encode tag
	tagBytes, err := EncodeTag(tag)
	if err != nil {
		return nil, fmt.Errorf("failed to encode tag: %w", err)
	}
	result = append(result, tagBytes...)

	// Encode length
	lengthBytes, err := EncodeLength(len(value))
	if err != nil {
		return nil, fmt.Errorf("failed to encode length: %w", err)
	}
	result = append(result, lengthBytes...)

	// Append value
	result = append(result, value...)

	return result, nil
}

// EncodeTag encodes an ASN.1 tag using BER rules
func EncodeTag(tag Tag) ([]byte, error) {
	if tag.Number < 0 {
		return nil, fmt.Errorf("tag number cannot be negative")
	}

	// Single byte tag encoding for tag numbers 0-30
	if tag.Number <= 30 {
		var b byte
		b |= byte(tag.Class << 6)
		if tag.Constructed {
			b |= 0x20
		}
		b |= byte(tag.Number)
		return []byte{b}, nil
	}

	// Multi-byte tag encoding for tag numbers > 30
	var result []byte
	
	// First byte: class (bits 7-6), constructed (bit 5), and 0x1F (bits 4-0)
	var firstByte byte
	firstByte |= byte(tag.Class << 6)
	if tag.Constructed {
		firstByte |= 0x20
	}
	firstByte |= 0x1F
	result = append(result, firstByte)

	// Encode tag number using base-128 encoding
	number := tag.Number
	var numberBytes []byte
	
	// Handle tag number 0 specially
	if number == 0 {
		numberBytes = []byte{0}
	} else {
		for number > 0 {
			numberBytes = append([]byte{byte(number & 0x7F)}, numberBytes...)
			number >>= 7
		}
	}

	// Set continuation bit (bit 7) for all bytes except the last
	for i := 0; i < len(numberBytes)-1; i++ {
		numberBytes[i] |= 0x80
	}

	result = append(result, numberBytes...)
	return result, nil
}

// EncodeLength encodes a length using BER rules
func EncodeLength(length int) ([]byte, error) {
	if length < 0 {
		return nil, fmt.Errorf("length cannot be negative")
	}

	// Short form: length < 128
	if length < 0x80 {
		return []byte{byte(length)}, nil
	}

	// Long form: length >= 128
	// First, determine how many bytes we need
	var lengthBytes []byte
	tempLength := length
	for tempLength > 0 {
		lengthBytes = append([]byte{byte(tempLength & 0xFF)}, lengthBytes...)
		tempLength >>= 8
	}

	// First byte: 0x80 | number of length bytes
	result := []byte{0x80 | byte(len(lengthBytes))}
	result = append(result, lengthBytes...)

	return result, nil
}

// DecodeTLV decodes a Tag-Length-Value structure from BER encoding
func DecodeTLV(data []byte) (*ASN1Value, int, error) {
	if len(data) == 0 {
		return nil, 0, fmt.Errorf("empty data")
	}

	offset := 0

	// Decode tag
	tag, tagLen, err := DecodeTag(data[offset:])
	if err != nil {
		return nil, 0, fmt.Errorf("failed to decode tag: %w", err)
	}
	offset += tagLen

	if offset >= len(data) {
		return nil, 0, fmt.Errorf("insufficient data for length")
	}

	// Decode length
	length, lengthLen, err := DecodeLength(data[offset:])
	if err != nil {
		return nil, 0, fmt.Errorf("failed to decode length: %w", err)
	}
	offset += lengthLen

	// Check if we have enough data for the value
	if offset+length > len(data) {
		return nil, 0, fmt.Errorf("insufficient data for value: need %d bytes, have %d", length, len(data)-offset)
	}

	// Extract value
	value := make([]byte, length)
	copy(value, data[offset:offset+length])
	offset += length

	return NewASN1Value(tag, value), offset, nil
}

// DecodeTag decodes an ASN.1 tag from BER encoding
func DecodeTag(data []byte) (Tag, int, error) {
	if len(data) == 0 {
		return Tag{}, 0, fmt.Errorf("empty data")
	}

	firstByte := data[0]
	class := int((firstByte & 0xC0) >> 6)
	constructed := (firstByte & 0x20) != 0
	tagNumber := int(firstByte & 0x1F)

	// Single byte tag
	if tagNumber != 0x1F {
		return Tag{
			Class:       class,
			Constructed: constructed,
			Number:      tagNumber,
		}, 1, nil
	}

	// Multi-byte tag
	if len(data) < 2 {
		return Tag{}, 0, fmt.Errorf("insufficient data for multi-byte tag")
	}

	offset := 1
	tagNumber = 0

	for {
		if offset >= len(data) {
			return Tag{}, 0, fmt.Errorf("incomplete multi-byte tag")
		}

		b := data[offset]
		offset++

		// Check for overflow
		if tagNumber > (1<<(32-7))-1 {
			return Tag{}, 0, fmt.Errorf("tag number too large")
		}

		tagNumber = (tagNumber << 7) | int(b&0x7F)

		// If bit 7 is not set, this is the last byte
		if (b & 0x80) == 0 {
			break
		}
	}

	return Tag{
		Class:       class,
		Constructed: constructed,
		Number:      tagNumber,
	}, offset, nil
}

// DecodeLength decodes a length from BER encoding
func DecodeLength(data []byte) (int, int, error) {
	if len(data) == 0 {
		return 0, 0, fmt.Errorf("empty data")
	}

	firstByte := data[0]

	// Short form
	if (firstByte & 0x80) == 0 {
		return int(firstByte), 1, nil
	}

	// Long form
	numLengthBytes := int(firstByte & 0x7F)

	// Indefinite form not supported in DER/BER strict encoding
	if numLengthBytes == 0 {
		return 0, 0, fmt.Errorf("indefinite length not supported")
	}

	if numLengthBytes > 4 {
		return 0, 0, fmt.Errorf("length too large: %d bytes", numLengthBytes)
	}

	if len(data) < 1+numLengthBytes {
		return 0, 0, fmt.Errorf("insufficient data for length: need %d bytes, have %d", 1+numLengthBytes, len(data))
	}

	length := 0
	for i := 1; i <= numLengthBytes; i++ {
		length = (length << 8) | int(data[i])
	}

	return length, 1 + numLengthBytes, nil
}

// DecodeAll decodes all ASN.1 objects from the given data
func DecodeAll(data []byte) ([]ASN1Object, error) {
	var objects []ASN1Object
	offset := 0

	for offset < len(data) {
		obj, consumed, err := DecodeTLV(data[offset:])
		if err != nil {
			return nil, fmt.Errorf("failed to decode object at offset %d: %w", offset, err)
		}
		objects = append(objects, obj)
		offset += consumed
	}

	return objects, nil
}

// WriteToWriter writes the BER-encoded object to an io.Writer
func WriteToWriter(w io.Writer, obj ASN1Object) error {
	encoded, err := obj.Encode()
	if err != nil {
		return fmt.Errorf("failed to encode object: %w", err)
	}
	
	_, err = w.Write(encoded)
	if err != nil {
		return fmt.Errorf("failed to write encoded data: %w", err)
	}
	
	return nil
}

// Convenience functions for easy encoding/decoding

// Encode encodes any ASN1Object to bytes - convenience wrapper
func Encode(obj ASN1Object) ([]byte, error) {
	return obj.Encode()
}

// Decode decodes the first ASN1Object from bytes - convenience wrapper  
func Decode(data []byte) (ASN1Object, error) {
	objects, err := DecodeAll(data)
	if err != nil {
		return nil, err
	}
	if len(objects) == 0 {
		return nil, fmt.Errorf("no objects found in data")
	}
	return objects[0], nil
}

// EncodeToHex encodes an ASN1Object and returns hex string representation
func EncodeToHex(obj ASN1Object) (string, error) {
	encoded, err := obj.Encode()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%02X", encoded), nil
}