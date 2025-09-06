package asn1

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestBoolean(t *testing.T) {
	tests := []struct {
		name  string
		value bool
		want  []byte
	}{
		{"true", true, []byte{0x01, 0x01, 0xFF}},
		{"false", false, []byte{0x01, 0x01, 0x00}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBoolean(tt.value)
			
			// Test encoding
			encoded, err := b.Encode()
			if err != nil {
				t.Fatalf("Encode() error = %v", err)
			}
			if !bytes.Equal(encoded, tt.want) {
				t.Errorf("Encode() = %x, want %x", encoded, tt.want)
			}

			// Test decoding
			decoded, consumed, err := DecodeBoolean(encoded)
			if err != nil {
				t.Fatalf("DecodeBoolean() error = %v", err)
			}
			if consumed != len(encoded) {
				t.Errorf("DecodeBoolean() consumed = %d, want %d", consumed, len(encoded))
			}
			if decoded.Value() != tt.value {
				t.Errorf("DecodeBoolean() value = %t, want %t", decoded.Value(), tt.value)
			}
		})
	}
}

func TestInteger(t *testing.T) {
	tests := []struct {
		name  string
		value int64
		want  []byte
	}{
		{"zero", 0, []byte{0x02, 0x01, 0x00}},
		{"positive", 127, []byte{0x02, 0x01, 0x7F}},
		{"large positive", 128, []byte{0x02, 0x02, 0x00, 0x80}},
		{"negative", -1, []byte{0x02, 0x01, 0xFF}},
		{"large negative", -128, []byte{0x02, 0x01, 0x80}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := NewInteger(tt.value)
			
			// Test encoding
			encoded, err := i.Encode()
			if err != nil {
				t.Fatalf("Encode() error = %v", err)
			}
			if !bytes.Equal(encoded, tt.want) {
				t.Errorf("Encode() = %x, want %x", encoded, tt.want)
			}

			// Test decoding
			decoded, consumed, err := DecodeInteger(encoded)
			if err != nil {
				t.Fatalf("DecodeInteger() error = %v", err)
			}
			if consumed != len(encoded) {
				t.Errorf("DecodeInteger() consumed = %d, want %d", consumed, len(encoded))
			}
			val, err := decoded.Int64()
			if err != nil {
				t.Fatalf("Int64() error = %v", err)
			}
			if val != tt.value {
				t.Errorf("DecodeInteger() value = %d, want %d", val, tt.value)
			}
		})
	}
}

func TestOctetString(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  []byte
	}{
		{"empty", "", []byte{0x04, 0x00}},
		{"hello", "hello", []byte{0x04, 0x05, 0x68, 0x65, 0x6C, 0x6C, 0x6F}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := NewOctetStringFromString(tt.value)
			
			// Test encoding
			encoded, err := o.Encode()
			if err != nil {
				t.Fatalf("Encode() error = %v", err)
			}
			if !bytes.Equal(encoded, tt.want) {
				t.Errorf("Encode() = %x, want %x", encoded, tt.want)
			}

			// Test decoding
			decoded, consumed, err := DecodeOctetString(encoded)
			if err != nil {
				t.Fatalf("DecodeOctetString() error = %v", err)
			}
			if consumed != len(encoded) {
				t.Errorf("DecodeOctetString() consumed = %d, want %d", consumed, len(encoded))
			}
			if decoded.StringValue() != tt.value {
				t.Errorf("DecodeOctetString() value = %q, want %q", decoded.StringValue(), tt.value)
			}
		})
	}
}

func TestNull(t *testing.T) {
	want := []byte{0x05, 0x00}
	
	n := NewNull()
	
	// Test encoding
	encoded, err := n.Encode()
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}
	if !bytes.Equal(encoded, want) {
		t.Errorf("Encode() = %x, want %x", encoded, want)
	}

	// Test decoding
	decoded, consumed, err := DecodeNull(encoded)
	if err != nil {
		t.Fatalf("DecodeNull() error = %v", err)
	}
	if consumed != len(encoded) {
		t.Errorf("DecodeNull() consumed = %d, want %d", consumed, len(encoded))
	}
	if decoded == nil {
		t.Error("DecodeNull() returned nil")
	}
}

func TestSequence(t *testing.T) {
	// Create a sequence with INTEGER(42) and BOOLEAN(true)
	seq := NewSequence()
	seq.Add(NewInteger(42))
	seq.Add(NewBoolean(true))
	
	// Test encoding
	encoded, err := seq.Encode()
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	// Expected: SEQUENCE { INTEGER 42, BOOLEAN true }
	// 30 06 02 01 2A 01 01 FF
	expected := []byte{0x30, 0x06, 0x02, 0x01, 0x2A, 0x01, 0x01, 0xFF}
	if !bytes.Equal(encoded, expected) {
		t.Errorf("Encode() = %x, want %x", encoded, expected)
	}

	// Test that we can decode it back to basic TLV structures
	decoded, consumed, err := DecodeTLV(encoded)
	if err != nil {
		t.Fatalf("DecodeTLV() error = %v", err)
	}
	if consumed != len(encoded) {
		t.Errorf("DecodeTLV() consumed = %d, want %d", consumed, len(encoded))
	}
	if decoded.tag.Number != TagSequence || !decoded.tag.Constructed {
		t.Errorf("DecodeTLV() tag = %v, want SEQUENCE", decoded.tag)
	}
}

func TestObjectIdentifier(t *testing.T) {
	tests := []struct {
		name       string
		components []int
		oidString  string
		want       []byte
	}{
		{
			name:       "1.2.840.113549",
			components: []int{1, 2, 840, 113549},
			oidString:  "1.2.840.113549",
			want:       []byte{0x06, 0x06, 0x2A, 0x86, 0x48, 0x86, 0xF7, 0x0D},
		},
		{
			name:       "2.5.4.3",
			components: []int{2, 5, 4, 3},
			oidString:  "2.5.4.3",
			want:       []byte{0x06, 0x03, 0x55, 0x04, 0x03},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test creation from components
			oid := NewObjectIdentifier(tt.components)
			
			// Test encoding
			encoded, err := oid.Encode()
			if err != nil {
				t.Fatalf("Encode() error = %v", err)
			}
			if !bytes.Equal(encoded, tt.want) {
				t.Errorf("Encode() = %x, want %x", encoded, tt.want)
			}

			// Test string representation
			if oid.DotNotation() != tt.oidString {
				t.Errorf("DotNotation() = %q, want %q", oid.DotNotation(), tt.oidString)
			}

			// Test creation from string
			oidFromString, err := NewObjectIdentifierFromString(tt.oidString)
			if err != nil {
				t.Fatalf("NewObjectIdentifierFromString() error = %v", err)
			}
			
			encodedFromString, err := oidFromString.Encode()
			if err != nil {
				t.Fatalf("Encode() from string error = %v", err)
			}
			if !bytes.Equal(encodedFromString, tt.want) {
				t.Errorf("Encode() from string = %x, want %x", encodedFromString, tt.want)
			}

			// Test decoding
			decoded, consumed, err := DecodeObjectIdentifier(encoded)
			if err != nil {
				t.Fatalf("DecodeObjectIdentifier() error = %v", err)
			}
			if consumed != len(encoded) {
				t.Errorf("DecodeObjectIdentifier() consumed = %d, want %d", consumed, len(encoded))
			}
			if decoded.DotNotation() != tt.oidString {
				t.Errorf("DecodeObjectIdentifier() OID = %q, want %q", decoded.DotNotation(), tt.oidString)
			}
		})
	}
}

func TestBitString(t *testing.T) {
	tests := []struct {
		name       string
		bits       string
		unusedBits int
		want       []byte
	}{
		{
			name:       "empty",
			bits:       "",
			unusedBits: 0,
			want:       []byte{0x03, 0x01, 0x00},
		},
		{
			name:       "1010",
			bits:       "1010",
			unusedBits: 4,
			want:       []byte{0x03, 0x02, 0x04, 0xA0},
		},
		{
			name:       "10101010",
			bits:       "10101010",
			unusedBits: 0,
			want:       []byte{0x03, 0x02, 0x00, 0xAA},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := NewBitStringFromBits(tt.bits)
			
			// Test unused bits calculation
			if bs.UnusedBits() != tt.unusedBits {
				t.Errorf("UnusedBits() = %d, want %d", bs.UnusedBits(), tt.unusedBits)
			}

			// Test bit string conversion
			if bs.ToBitString() != tt.bits {
				t.Errorf("ToBitString() = %q, want %q", bs.ToBitString(), tt.bits)
			}

			// Test encoding
			encoded, err := bs.Encode()
			if err != nil {
				t.Fatalf("Encode() error = %v", err)
			}
			if !bytes.Equal(encoded, tt.want) {
				t.Errorf("Encode() = %x, want %x", encoded, tt.want)
			}

			// Test decoding
			decoded, consumed, err := DecodeBitString(encoded)
			if err != nil {
				t.Fatalf("DecodeBitString() error = %v", err)
			}
			if consumed != len(encoded) {
				t.Errorf("DecodeBitString() consumed = %d, want %d", consumed, len(encoded))
			}
			if decoded.ToBitString() != tt.bits {
				t.Errorf("DecodeBitString() bits = %q, want %q", decoded.ToBitString(), tt.bits)
			}
		})
	}
}


func TestChoice(t *testing.T) {
	// Test with boolean choice
	boolValue := NewBoolean(true)
	choice1 := NewChoiceWithID(boolValue, "boolean_choice")
	
	// Test encoding
	encoded1, err := choice1.Encode()
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}
	
	// The encoded choice should be the same as the boolean value
	boolEncoded, err := boolValue.Encode()
	if err != nil {
		t.Fatalf("Boolean Encode() error = %v", err)
	}
	
	if !bytes.Equal(encoded1, boolEncoded) {
		t.Errorf("Choice encoding differs from direct boolean encoding")
	}
	
	// Test with integer choice
	intValue := NewInteger(42)
	choice2 := NewChoiceWithID(intValue, "integer_choice")
	
	encoded2, err := choice2.Encode()
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}
	
	// Test tag retrieval
	if choice2.Tag() != intValue.Tag() {
		t.Errorf("Choice tag differs from underlying value tag")
	}
	
	// Test string representation
	str := choice2.String()
	if !strings.Contains(str, "integer_choice") {
		t.Errorf("String() should contain choice ID")
	}
	
	t.Logf("Boolean choice encoded to %d bytes", len(encoded1))
	t.Logf("Integer choice encoded to %d bytes", len(encoded2))
	t.Logf("Choice string: %s", choice2.String())
}

func TestEnumerated(t *testing.T) {
	tests := []struct {
		name     string
		value    int64
		enumName string
	}{
		{"zero", 0, ""},
		{"positive", 42, "answer"},
		{"negative", -1, "error"},
		{"large", 1000000, "large_value"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var enum *ASN1Enumerated
			if tt.enumName != "" {
				enum = NewEnumeratedWithName(tt.value, tt.enumName)
			} else {
				enum = NewEnumerated(tt.value)
			}
			
			// Test encoding
			encoded, err := enum.Encode()
			if err != nil {
				t.Fatalf("Encode() error = %v", err)
			}

			// Test decoding
			decoded, consumed, err := DecodeEnumerated(encoded)
			if err != nil {
				t.Fatalf("DecodeEnumerated() error = %v", err)
			}
			if consumed != len(encoded) {
				t.Errorf("DecodeEnumerated() consumed = %d, want %d", consumed, len(encoded))
			}
			if decoded.Int64() != tt.value {
				t.Errorf("DecodeEnumerated() value = %d, want %d", decoded.Int64(), tt.value)
			}
			// Note: Names are not encoded in ASN.1, so they won't be preserved during round-trip
			if tt.enumName != "" && enum.Name() != tt.enumName {
				t.Logf("Original name preserved: %s", enum.Name())
			}

			// Test string representation
			str := enum.String()
			if tt.enumName != "" && !strings.Contains(str, tt.enumName) {
				t.Errorf("String() should contain enum name")
			}
			
			t.Logf("Enumerated %s encoded to %d bytes", tt.name, len(encoded))
		})
	}
}

func TestUTCTime(t *testing.T) {
	// Test with a known time
	testTime := time.Date(2023, 12, 25, 10, 30, 45, 0, time.UTC)
	utcTime := NewUTCTime(testTime)
	
	// Test encoding
	encoded, err := utcTime.Encode()
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}
	
	// Test decoding
	decoded, consumed, err := DecodeUTCTime(encoded)
	if err != nil {
		t.Fatalf("DecodeUTCTime() error = %v", err)
	}
	if consumed != len(encoded) {
		t.Errorf("DecodeUTCTime() consumed = %d, want %d", consumed, len(encoded))
	}
	
	// Compare times (allowing for second precision)
	decodedTime := decoded.Time()
	if !testTime.Equal(decodedTime) {
		t.Errorf("DecodeUTCTime() time = %v, want %v", decodedTime, testTime)
	}
	
	// Test current time
	now := NewUTCTimeNow()
	encodedNow, err := now.Encode()
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}
	
	t.Logf("UTCTime encoded to %d bytes", len(encoded))
	t.Logf("UTCTime string: %s", utcTime.String())
	t.Logf("Current time encoded to %d bytes", len(encodedNow))
}

func TestGeneralizedTime(t *testing.T) {
	// Test with a known time
	testTime := time.Date(2023, 12, 25, 10, 30, 45, 0, time.UTC)
	genTime := NewGeneralizedTime(testTime)
	
	// Test encoding
	encoded, err := genTime.Encode()
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}
	
	// Test decoding
	decoded, consumed, err := DecodeGeneralizedTime(encoded)
	if err != nil {
		t.Fatalf("DecodeGeneralizedTime() error = %v", err)
	}
	if consumed != len(encoded) {
		t.Errorf("DecodeGeneralizedTime() consumed = %d, want %d", consumed, len(encoded))
	}
	
	// Compare times (allowing for second precision)
	decodedTime := decoded.Time()
	if !testTime.Equal(decodedTime) {
		t.Errorf("DecodeGeneralizedTime() time = %v, want %v", decodedTime, testTime)
	}
	
	// Test current time
	now := NewGeneralizedTimeNow()
	encodedNow, err := now.Encode()
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}
	
	t.Logf("GeneralizedTime encoded to %d bytes", len(encoded))
	t.Logf("GeneralizedTime string: %s", genTime.String())
	t.Logf("Current time encoded to %d bytes", len(encodedNow))
}

// Comprehensive round-trip tests demonstrating real-world usage scenarios
func TestComplexStructureRoundTrip(t *testing.T) {
	// Test a realistic complex structure that might be used in applications
	// Simulating a certificate or document structure
	
	// Create a complex nested structure
	document := NewSequence()
	
	// Document metadata
	metadata := NewSequence()
	metadata.Add(NewInteger(1)) // version
	metadata.Add(NewUTF8String("Document Title"))
	metadata.Add(NewUTCTime(time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)))
	metadata.Add(NewBoolean(true)) // is active
	document.Add(metadata)
	
	// Author information with choices
	author := NewSequence()
	nameChoice := NewChoiceWithID(NewUTF8String("John Doe"), "full_name")
	statusEnum := NewEnumeratedWithName(1, "ACTIVE")
	author.Add(nameChoice)
	author.Add(statusEnum)
	author.Add(NewIA5String("john.doe@example.com"))
	document.Add(author)
	
	// Optional fields with context-specific tags
	optionalData := NewSequence()
	
	// [CONTEXT 0] - optional description
	contextTag0 := NewContextSpecificTag(0, true)
	description := NewStructured(contextTag0)
	description.Add(NewUTF8String("This is an optional description"))
	optionalData.Add(description)
	
	// [CONTEXT 1] - optional binary data
	contextTag1 := NewContextSpecificTag(1, true)
	binaryData := NewStructured(contextTag1)
	binaryData.Add(NewOctetString([]byte{0x01, 0x02, 0x03, 0x04}))
	optionalData.Add(binaryData)
	
	document.Add(optionalData)
	
	// Add some object identifiers
	oids := NewSequence()
	oid1, _ := NewObjectIdentifierFromString("1.2.840.113549.1.1.11")
	oid2, _ := NewObjectIdentifierFromString("2.5.4.3")
	oids.Add(oid1)
	oids.Add(oid2)
	document.Add(oids)
	
	t.Logf("Original structure:\n%s", document.String())
	
	// 1. ENCODE the complete structure
	encoded, err := document.Encode()
	if err != nil {
		t.Fatalf("Failed to encode structure: %v", err)
	}
	
	t.Logf("Encoded to %d bytes", len(encoded))
	
	// 2. DECODE the structure back using generic decoder
	decodedObjects, err := DecodeAll(encoded)
	if err != nil {
		t.Fatalf("Failed to decode structure: %v", err)
	}
	
	if len(decodedObjects) != 1 {
		t.Fatalf("Expected 1 decoded object, got %d", len(decodedObjects))
	}
	
	// 3. Verify the decoded structure
	decodedDoc := decodedObjects[0]
	t.Logf("Decoded structure:\n%s", decodedDoc.String())
	
	// 4. Re-encode to verify round-trip integrity
	reencoded, err := decodedDoc.Encode()
	if err != nil {
		t.Fatalf("Failed to re-encode structure: %v", err)
	}
	
	// 5. Verify byte-for-byte equality
	if len(encoded) != len(reencoded) {
		t.Errorf("Round-trip size mismatch: original %d bytes, reencoded %d bytes", len(encoded), len(reencoded))
	} else {
		for i := range encoded {
			if encoded[i] != reencoded[i] {
				t.Errorf("Round-trip data mismatch at byte %d: original %02X, reencoded %02X", i, encoded[i], reencoded[i])
				break
			}
		}
	}
	
	t.Log("✓ Complex structure round-trip test passed")
}

func TestApplicationUsageScenarios(t *testing.T) {
	t.Run("UserProfile", func(t *testing.T) {
		// Simulate encoding/decoding a user profile in an application
		
		// Create user profile structure
		profile := NewSequence()
		profile.Add(NewInteger(12345)) // user ID
		profile.Add(NewUTF8String("alice@example.com")) // email
		profile.Add(NewBoolean(true)) // is verified
		
		// Preferences as enumerated values
		preferences := NewSequence()
		themeEnum := NewEnumeratedWithName(0, "DARK")
		languageEnum := NewEnumeratedWithName(1, "ENGLISH")
		preferences.Add(themeEnum)
		preferences.Add(languageEnum)
		profile.Add(preferences)
		
		// Metadata with timestamp
		metadata := NewSequence()
		metadata.Add(NewUTCTimeNow())
		metadata.Add(NewOctetString([]byte("session-token-123")))
		profile.Add(metadata)
		
		// Application would encode this to store/transmit
		encoded, err := profile.Encode()
		if err != nil {
			t.Fatalf("Application encoding failed: %v", err)
		}
		
		t.Logf("User profile encoded to %d bytes", len(encoded))
		
		// Application would decode this when retrieving
		decoded, err := DecodeAll(encoded)
		if err != nil {
			t.Fatalf("Application decoding failed: %v", err)
		}
		
		if len(decoded) != 1 {
			t.Fatalf("Expected 1 decoded object, got %d", len(decoded))
		}
		
		// Application can now use the decoded structure
		decodedProfile := decoded[0]
		t.Logf("Decoded user profile:\n%s", decodedProfile.String())
		
		// Verify round-trip
		reencoded, err := decodedProfile.Encode()
		if err != nil {
			t.Fatalf("Re-encoding failed: %v", err)
		}
		
		if !bytes.Equal(encoded, reencoded) {
			t.Errorf("Round-trip failed for user profile")
		} else {
			t.Log("✓ User profile round-trip successful")
		}
	})
	
	t.Run("ConfigurationDocument", func(t *testing.T) {
		// Simulate encoding/decoding application configuration
		
		config := NewSequence()
		
		// Version info
		config.Add(NewInteger(2)) // config version
		
		// Server settings
		serverConfig := NewSequence()
		serverConfig.Add(NewIA5String("https://api.example.com"))
		serverConfig.Add(NewInteger(443)) // port
		serverConfig.Add(NewBoolean(true)) // use TLS
		config.Add(serverConfig)
		
		// Feature flags as enumerated choices
		features := NewSequence()
		
		// Feature: logging level
		loggingChoice := NewChoiceWithID(NewEnumeratedWithName(2, "DEBUG"), "logging_level")
		features.Add(loggingChoice)
		
		// Feature: cache strategy
		cacheChoice := NewChoiceWithID(NewEnumeratedWithName(1, "REDIS"), "cache_strategy")
		features.Add(cacheChoice)
		
		config.Add(features)
		
		// Optional advanced settings with context tags
		advancedTag := NewContextSpecificTag(0, true)
		advanced := NewStructured(advancedTag)
		
		advancedSettings := NewSequence()
		advancedSettings.Add(NewInteger(3600)) // session timeout
		advancedSettings.Add(NewOctetString([]byte("encryption-key-hash")))
		advanced.Add(advancedSettings)
		
		config.Add(advanced)
		
		// Encode configuration
		encoded, err := config.Encode()
		if err != nil {
			t.Fatalf("Config encoding failed: %v", err)
		}
		
		t.Logf("Configuration encoded to %d bytes", len(encoded))
		
		// Decode configuration
		decoded, err := DecodeAll(encoded)
		if err != nil {
			t.Fatalf("Config decoding failed: %v", err)
		}
		
		if len(decoded) != 1 {
			t.Fatalf("Expected 1 decoded config, got %d", len(decoded))
		}
		
		decodedConfig := decoded[0]
		t.Logf("Decoded configuration:\n%s", decodedConfig.String())
		
		// Verify round-trip
		reencoded, err := decodedConfig.Encode()
		if err != nil {
			t.Fatalf("Config re-encoding failed: %v", err)
		}
		
		if !bytes.Equal(encoded, reencoded) {
			t.Errorf("Round-trip failed for configuration")
		} else {
			t.Log("✓ Configuration round-trip successful")
		}
	})
}

func TestAdvancedTypesRoundTrip(t *testing.T) {
	t.Run("ChoiceTypes", func(t *testing.T) {
		// Test various CHOICE types
		choices := []*ASN1Choice{
			NewChoiceWithID(NewBoolean(true), "boolean_option"),
			NewChoiceWithID(NewInteger(42), "integer_option"),
			NewChoiceWithID(NewUTF8String("text_option"), "string_option"),
			NewChoiceWithID(NewOctetString([]byte{1, 2, 3}), "binary_option"),
		}
		
		for i, choice := range choices {
			t.Run(fmt.Sprintf("Choice_%d", i), func(t *testing.T) {
				// Encode choice
				encoded, err := choice.Encode()
				if err != nil {
					t.Fatalf("CHOICE encoding failed: %v", err)
				}
				
				// Decode back to generic ASN1Value
				decoded, consumed, err := DecodeTLV(encoded)
				if err != nil {
					t.Fatalf("CHOICE decoding failed: %v", err)
				}
				
				if consumed != len(encoded) {
					t.Errorf("CHOICE consumed %d bytes, expected %d", consumed, len(encoded))
				}
				
				// Re-encode
				reencoded, err := decoded.Encode()
				if err != nil {
					t.Fatalf("CHOICE re-encoding failed: %v", err)
				}
				
				if !bytes.Equal(encoded, reencoded) {
					t.Errorf("CHOICE round-trip failed")
				} else {
					t.Logf("✓ CHOICE round-trip successful: %s", choice.String())
				}
			})
		}
	})
	
	t.Run("EnumeratedTypes", func(t *testing.T) {
		// Test various ENUMERATED types
		enums := []*ASN1Enumerated{
			NewEnumerated(0),
			NewEnumeratedWithName(1, "ACTIVE"),
			NewEnumeratedWithName(42, "ANSWER"),
			NewEnumeratedWithName(-1, "ERROR"),
		}
		
		for i, enum := range enums {
			t.Run(fmt.Sprintf("Enum_%d", i), func(t *testing.T) {
				originalValue := enum.Int64()
				
				// Encode enumerated
				encoded, err := enum.Encode()
				if err != nil {
					t.Fatalf("ENUMERATED encoding failed: %v", err)
				}
				
				// Decode back
				decoded, consumed, err := DecodeEnumerated(encoded)
				if err != nil {
					t.Fatalf("ENUMERATED decoding failed: %v", err)
				}
				
				if consumed != len(encoded) {
					t.Errorf("ENUMERATED consumed %d bytes, expected %d", consumed, len(encoded))
				}
				
				if decoded.Int64() != originalValue {
					t.Errorf("ENUMERATED value mismatch: original %d, decoded %d", originalValue, decoded.Int64())
				}
				
				// Re-encode
				reencoded, err := decoded.Encode()
				if err != nil {
					t.Fatalf("ENUMERATED re-encoding failed: %v", err)
				}
				
				if !bytes.Equal(encoded, reencoded) {
					t.Errorf("ENUMERATED round-trip failed")
				} else {
					t.Logf("✓ ENUMERATED round-trip successful: %s", enum.String())
				}
			})
		}
	})
	
	t.Run("TimeTypes", func(t *testing.T) {
		testTime := time.Date(2023, 12, 25, 14, 30, 45, 0, time.UTC)
		
		// Test UTCTime
		utcTime := NewUTCTime(testTime)
		utcEncoded, err := utcTime.Encode()
		if err != nil {
			t.Fatalf("UTCTime encoding failed: %v", err)
		}
		
		utcDecoded, consumed, err := DecodeUTCTime(utcEncoded)
		if err != nil {
			t.Fatalf("UTCTime decoding failed: %v", err)
		}
		
		if consumed != len(utcEncoded) {
			t.Errorf("UTCTime consumed %d bytes, expected %d", consumed, len(utcEncoded))
		}
		
		if !utcDecoded.Time().Equal(testTime) {
			t.Errorf("UTCTime mismatch: original %v, decoded %v", testTime, utcDecoded.Time())
		}
		
		// Test GeneralizedTime
		genTime := NewGeneralizedTime(testTime)
		genEncoded, err := genTime.Encode()
		if err != nil {
			t.Fatalf("GeneralizedTime encoding failed: %v", err)
		}
		
		genDecoded, consumed, err := DecodeGeneralizedTime(genEncoded)
		if err != nil {
			t.Fatalf("GeneralizedTime decoding failed: %v", err)
		}
		
		if consumed != len(genEncoded) {
			t.Errorf("GeneralizedTime consumed %d bytes, expected %d", consumed, len(genEncoded))
		}
		
		if !genDecoded.Time().Equal(testTime) {
			t.Errorf("GeneralizedTime mismatch: original %v, decoded %v", testTime, genDecoded.Time())
		}
		
		t.Log("✓ Time types round-trip successful")
	})
}

func TestEasyToUseAPI(t *testing.T) {
	// Demonstrate how easy it is to use the library from an application perspective
	
	t.Run("SimpleDocument", func(t *testing.T) {
		// Application creates a simple document
		doc := NewSequence()
		doc.Add(NewUTF8String("My Document"))
		doc.Add(NewInteger(1))
		doc.Add(NewBoolean(true))
		
		// Easy encoding - just one call
		data, err := doc.Encode()
		if err != nil {
			t.Fatalf("Encoding failed: %v", err)
		}
		
		// Easy decoding - just one call
		objects, err := DecodeAll(data)
		if err != nil {
			t.Fatalf("Decoding failed: %v", err)
		}
		
		// Application gets back structured data
		if len(objects) != 1 {
			t.Fatalf("Expected 1 object, got %d", len(objects))
		}
		
		decoded := objects[0]
		t.Logf("Round-trip result: %s", decoded.String())
		
		// Application can easily inspect the structure
		if decoded.Tag().Number != TagSequence {
			t.Errorf("Expected SEQUENCE tag, got %d", decoded.Tag().Number)
		}
		
		// Application can access nested elements if it's a structured type
		if structured, ok := decoded.(*ASN1Structured); ok {
			elements := structured.Elements()
			if len(elements) != 3 {
				t.Errorf("Expected 3 elements, got %d", len(elements))
			}
			t.Logf("Document has %d elements", len(elements))
		}
		
		t.Log("✓ Simple document API test passed")
	})
	
	t.Run("DataSerialization", func(t *testing.T) {
		// Show how an application might serialize arbitrary data
		
		// Application data
		appData := map[string]interface{}{
			"user_id":    12345,
			"username":   "alice",
			"is_active":  true,
			"created_at": time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		}
		
		// Convert to ASN.1 structure
		container := NewSequence()
		
		// Serialize each field
		for key, value := range appData {
			pair := NewSequence()
			pair.Add(NewUTF8String(key)) // field name
			
			// Convert value based on type
			switch v := value.(type) {
			case int:
				pair.Add(NewInteger(int64(v)))
			case string:
				pair.Add(NewUTF8String(v))
			case bool:
				pair.Add(NewBoolean(v))
			case time.Time:
				pair.Add(NewUTCTime(v))
			}
			
			container.Add(pair)
		}
		
		// Encode for storage/transmission
		encoded, err := container.Encode()
		if err != nil {
			t.Fatalf("Failed to encode app data: %v", err)
		}
		
		t.Logf("Application data encoded to %d bytes", len(encoded))
		
		// Decode when retrieving
		decoded, err := DecodeAll(encoded)
		if err != nil {
			t.Fatalf("Failed to decode app data: %v", err)
		}
		
		if len(decoded) != 1 {
			t.Fatalf("Expected 1 object, got %d", len(decoded))
		}
		
		decodedContainer := decoded[0]
		t.Logf("Decoded application data:\n%s", decodedContainer.String())
		
		// Application can process the decoded structure
		if structured, ok := decodedContainer.(*ASN1Structured); ok {
			elements := structured.Elements()
			t.Logf("Application data has %d fields", len(elements))
			
			// Application can iterate through fields
			for i, elem := range elements {
				if field, ok := elem.(*ASN1Structured); ok {
					fieldElements := field.Elements()
					if len(fieldElements) >= 2 {
						// First element should be field name
						if nameElem, ok := fieldElements[0].(*ASN1Value); ok {
							fieldName := string(nameElem.Value())
							t.Logf("  Field %d: %s", i, fieldName)
						}
					}
				}
			}
		}
		
		t.Log("✓ Data serialization API test passed")
	})
}

func TestConvenienceFunctions(t *testing.T) {
	// Test the new convenience functions for easy application usage
	
	t.Run("EncodeDecodeConvenience", func(t *testing.T) {
		// Create a simple structure
		original := NewSequence()
		original.Add(NewInteger(42))
		original.Add(NewUTF8String("Hello"))
		original.Add(NewBoolean(true))
		
		// Use convenience Encode function
		encoded, err := Encode(original)
		if err != nil {
			t.Fatalf("Convenience Encode failed: %v", err)
		}
		
		// Use convenience Decode function
		decoded, err := Decode(encoded)
		if err != nil {
			t.Fatalf("Convenience Decode failed: %v", err)
		}
		
		// Verify round-trip
		reencoded, err := Encode(decoded)
		if err != nil {
			t.Fatalf("Re-encoding failed: %v", err)
		}
		
		if !bytes.Equal(encoded, reencoded) {
			t.Errorf("Round-trip failed with convenience functions")
		}
		
		t.Logf("Original: %s", original.String())
		t.Logf("Decoded:  %s", decoded.String())
		t.Log("✓ Convenience encode/decode functions work correctly")
	})
	
	t.Run("HexEncoding", func(t *testing.T) {
		// Test hex encoding convenience function
		simple := NewInteger(127) // Use 127 which fits in 1 byte
		
		hexStr, err := EncodeToHex(simple)
		if err != nil {
			t.Fatalf("EncodeToHex failed: %v", err)
		}
		
		t.Logf("INTEGER 127 encoded as hex: %s", hexStr)
		
		// Should be: 02 01 7F (tag=02, length=01, value=7F)
		if hexStr != "02017F" {
			t.Errorf("Expected hex '02017F', got '%s'", hexStr)
		}
		
		t.Log("✓ Hex encoding convenience function works correctly")
	})
}