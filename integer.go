package asn1

import (
	"fmt"
	"math/big"
)

// ASN1Integer represents an ASN.1 INTEGER
type ASN1Integer struct {
	value *big.Int
}

// NewInteger creates a new ASN1Integer from an int64
func NewInteger(value int64) *ASN1Integer {
	return &ASN1Integer{value: big.NewInt(value)}
}

// NewIntegerFromBigInt creates a new ASN1Integer from a big.Int
func NewIntegerFromBigInt(value *big.Int) *ASN1Integer {
	// Create a copy to avoid external modification
	copied := new(big.Int).Set(value)
	return &ASN1Integer{value: copied}
}

// NewIntegerFromBytes creates a new ASN1Integer from byte representation
func NewIntegerFromBytes(data []byte) *ASN1Integer {
	value := new(big.Int).SetBytes(data)
	// Handle two's complement for negative numbers
	if len(data) > 0 && (data[0]&0x80) != 0 {
		// This is a negative number in two's complement
		// Calculate the negative value
		modulus := new(big.Int).Lsh(big.NewInt(1), uint(len(data)*8))
		value.Sub(value, modulus)
	}
	return &ASN1Integer{value: value}
}

// Value returns the integer value as a big.Int
func (i *ASN1Integer) Value() *big.Int {
	// Return a copy to prevent external modification
	return new(big.Int).Set(i.value)
}

// Int64 returns the integer value as an int64 if it fits
func (i *ASN1Integer) Int64() (int64, error) {
	if !i.value.IsInt64() {
		return 0, fmt.Errorf("integer value too large for int64")
	}
	return i.value.Int64(), nil
}

// Tag returns the ASN.1 tag for INTEGER
func (i *ASN1Integer) Tag() Tag {
	return NewUniversalTag(TagInteger, false)
}

// Encode returns the BER encoding of the integer
func (i *ASN1Integer) Encode() ([]byte, error) {
	value := i.encodeIntegerValue()
	return EncodeTLV(i.Tag(), value)
}

func (i *ASN1Integer) encodeIntegerValue() []byte {
	if i.value.Sign() == 0 {
		return []byte{0x00}
	}

	bytes := i.value.Bytes()
	
	// For positive numbers, if the high bit is set, prepend a zero byte
	// to avoid confusion with negative numbers in two's complement
	if i.value.Sign() > 0 && len(bytes) > 0 && (bytes[0]&0x80) != 0 {
		bytes = append([]byte{0x00}, bytes...)
	}
	
	// For negative numbers, use two's complement representation
	if i.value.Sign() < 0 {
		// For negative numbers, we need to compute the minimum number of bytes
		// to represent the number in two's complement form
		
		// Special case for small negative numbers that fit in one byte
		if i.value.Cmp(big.NewInt(-128)) >= 0 && i.value.Cmp(big.NewInt(0)) < 0 {
			val := i.value.Int64()
			return []byte{byte(val)}
		}
		
		// For larger negative numbers, calculate two's complement
		abs := new(big.Int).Abs(i.value)
		
		// Find minimum number of bits needed
		bitLen := abs.BitLen()
		// We need at least bitLen + 1 bits for the sign
		if bitLen%8 == 0 {
			bitLen += 8
		} else {
			bitLen = ((bitLen / 8) + 1) * 8
		}
		byteLen := bitLen / 8
		
		// Create modulus (2^(8*byteLen))
		modulus := new(big.Int).Lsh(big.NewInt(1), uint(byteLen*8))
		
		// Calculate two's complement: modulus + value (since value is negative)
		twosComp := new(big.Int).Add(modulus, i.value)
		bytes = twosComp.Bytes()
		
		// Pad with 0xFF bytes if needed to maintain the correct length
		for len(bytes) < byteLen {
			bytes = append([]byte{0xFF}, bytes...)
		}
	}
	
	return bytes
}

// String returns a string representation of the integer
func (i *ASN1Integer) String() string {
	return fmt.Sprintf("INTEGER %s", i.value.String())
}

// TaggedString returns a string representation with tag information
func (i *ASN1Integer) TaggedString() string {
	return fmt.Sprintf("%s INTEGER: %s", i.Tag().TagString(), i.value.String())
}

// DecodeIntegerValue decodes an integer value from raw bytes
func DecodeIntegerValue(data []byte) (*big.Int, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("integer value cannot be empty")
	}
	
	value := new(big.Int).SetBytes(data)
	
	// Handle two's complement for negative numbers
	if (data[0] & 0x80) != 0 {
		// This is a negative number in two's complement
		// Calculate the negative value
		modulus := new(big.Int).Lsh(big.NewInt(1), uint(len(data)*8))
		value.Sub(value, modulus)
	}
	
	return value, nil
}

// DecodeInteger decodes an ASN1Integer from BER-encoded data
func DecodeInteger(data []byte) (*ASN1Integer, int, error) {
	asn1Value, consumed, err := DecodeTLV(data)
	if err != nil {
		return nil, 0, err
	}

	if asn1Value.tag.Class != 0 || asn1Value.tag.Number != TagInteger {
		return nil, 0, fmt.Errorf("expected INTEGER tag, got class=%d number=%d", asn1Value.tag.Class, asn1Value.tag.Number)
	}

	value, err := DecodeIntegerValue(asn1Value.value)
	if err != nil {
		return nil, 0, err
	}

	return NewIntegerFromBigInt(value), consumed, nil
}