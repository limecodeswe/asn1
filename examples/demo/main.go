package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/limecodeswe/asn1"
	"github.com/limecodeswe/asn1/examples"
)

func main() {
	fmt.Println("ASN.1 Library Demonstration")
	fmt.Println("============================")

	// Demonstrate basic types
	demonstrateBasicTypes()

	// Demonstrate structured types
	demonstrateStructuredTypes()

	// Demonstrate Person example
	demonstratePersonExample()

	// Demonstrate context-specific tags
	demonstrateContextSpecificTags()

	// Demonstrate object identifiers
	demonstrateObjectIdentifiers()

	// Demonstrate encoding/decoding round trip
	demonstrateRoundTrip()

	// Demonstrate new types (CHOICE, ENUMERATED, Time types)
	demonstrateNewTypes()

	// Demonstrate complex hierarchical structure
	demonstrateComplexHierarchy()
}

func demonstrateBasicTypes() {
	fmt.Println("\n1. Basic ASN.1 Types")
	fmt.Println("-------------------")

	// BOOLEAN
	boolTrue := asn1.NewBoolean(true)
	boolFalse := asn1.NewBoolean(false)
	fmt.Printf("Boolean true:  %s\n", boolTrue.String())
	fmt.Printf("Boolean false: %s\n", boolFalse.String())

	// INTEGER
	intSmall := asn1.NewInteger(42)
	intLarge := asn1.NewInteger(123456789)
	intNegative := asn1.NewInteger(-12345)
	fmt.Printf("Integer small:    %s\n", intSmall.String())
	fmt.Printf("Integer large:    %s\n", intLarge.String())
	fmt.Printf("Integer negative: %s\n", intNegative.String())

	// OCTET STRING
	octetString := asn1.NewOctetStringFromString("Hello, ASN.1!")
	octetBinary := asn1.NewOctetString([]byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD, 0xEF})
	fmt.Printf("Octet string text:   %s\n", octetString.String())
	fmt.Printf("Octet string binary: %s\n", octetBinary.String())

	// NULL
	null := asn1.NewNull()
	fmt.Printf("Null: %s\n", null.String())

	// BIT STRING
	bitString1 := asn1.NewBitStringFromBits("10101010")
	bitString2 := asn1.NewBitStringFromBits("110011")
	fmt.Printf("Bit string 1: %s\n", bitString1.String())
	fmt.Printf("Bit string 2: %s\n", bitString2.String())

	// STRING types
	utf8String := asn1.NewUTF8String("Hello, 世界! 🌍")
	printableString := asn1.NewPrintableString("Hello World")
	ia5String := asn1.NewIA5String("user@example.com")
	fmt.Printf("UTF8 string:      %s\n", utf8String.String())
	fmt.Printf("Printable string: %s\n", printableString.String())
	fmt.Printf("IA5 string:       %s\n", ia5String.String())
}

func demonstrateStructuredTypes() {
	fmt.Println("\n2. Structured Types (SEQUENCE and SET)")
	fmt.Println("------------------------------------")

	// Create a SEQUENCE
	seq := asn1.NewSequence()
	seq.Add(asn1.NewInteger(1))
	seq.Add(asn1.NewUTF8String("First"))
	seq.Add(asn1.NewBoolean(true))
	seq.Add(asn1.NewOctetStringFromString("data"))

	fmt.Printf("Sequence: %s\n", seq.String())
	fmt.Println("Sequence compact view:")
	fmt.Println(seq.CompactString())

	// Create a SET
	set := asn1.NewSet()
	set.Add(asn1.NewInteger(100))
	set.Add(asn1.NewPrintableString("SetElement"))
	set.Add(asn1.NewNull())

	fmt.Printf("\nSet: %s\n", set.String())
	fmt.Println("Set compact view:")
	fmt.Println(set.CompactString())

	// Nested structures
	nested := asn1.NewSequence()
	nested.Add(asn1.NewInteger(999))
	nested.Add(seq) // Add the sequence from above
	nested.Add(set) // Add the set from above

	fmt.Printf("\nNested structure: %s\n", nested.String())
	fmt.Println("Nested structure compact view:")
	fmt.Println(nested.CompactString())
}

func demonstratePersonExample() {
	fmt.Println("\n3. Realistic Person Example")
	fmt.Println("---------------------------")

	// Create a manager
	manager := examples.NewPerson(1, "Alice Johnson", "alice.johnson@company.com", true).
		SetDepartment("Engineering").
		SetPhoneNumber("+1-555-0100").
		SetSalary(120000).
		SetEmployeeType(0). // full-time
		SetPermissions([]string{"admin", "read", "write", "delete"})

	// Create an employee with the manager
	employee := examples.NewPerson(123, "John Doe", "john.doe@company.com", true).
		SetDepartment("Engineering").
		SetPhoneNumber("+1-555-0123").
		SetSalary(85000).
		SetEmployeeType(0). // full-time
		SetManager(manager).
		SetPermissions([]string{"read", "write"}).
		SetMetadata([]byte("Some additional metadata about the employee"))

	// Set birthday
	birthday := time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC)
	employee.SetBirthday(birthday)

	fmt.Println("Manager:")
	fmt.Println(manager.String())

	fmt.Println("\nEmployee:")
	fmt.Println(employee.String())

	// Create a person directory
	directory := examples.NewPersonDirectory()
	directory.AddPerson(manager)
	directory.AddPerson(employee)

	// Add some more employees
	employee2 := examples.NewPerson(124, "Jane Smith", "jane.smith@company.com", true).
		SetDepartment("Sales").
		SetEmployeeType(1). // part-time
		SetPermissions([]string{"read"})

	employee3 := examples.NewPerson(125, "Bob Wilson", "bob.wilson@company.com", false).
		SetDepartment("Marketing").
		SetEmployeeType(2). // contractor
		SetPermissions([]string{"read", "write"})

	directory.AddPerson(employee2)
	directory.AddPerson(employee3)

	fmt.Println("\nPerson Directory:")
	fmt.Println(directory.CompactString())

	// Encode the directory
	encoded, err := directory.Encode()
	if err != nil {
		log.Printf("Failed to encode directory: %v", err)
	} else {
		fmt.Printf("\nDirectory encoded to %d bytes\n", len(encoded))
		fmt.Printf("Encoded data (first 64 bytes): %02X\n", encoded[:min(64, len(encoded))])
	}
}

func demonstrateContextSpecificTags() {
	fmt.Println("\n4. Context-Specific Tags")
	fmt.Println("-------------------------")

	// Create context-specific tagged elements
	contextTag0 := asn1.NewContextSpecificTag(0, false)
	contextElement0 := asn1.NewStructured(contextTag0)
	contextElement0.Add(asn1.NewUTF8String("Optional field 0"))

	contextTag1 := asn1.NewContextSpecificTag(1, true)
	contextElement1 := asn1.NewStructured(contextTag1)
	innerSeq := asn1.NewSequence()
	innerSeq.Add(asn1.NewInteger(42))
	innerSeq.Add(asn1.NewBoolean(true))
	contextElement1.Add(innerSeq)

	contextTag2 := asn1.NewContextSpecificTag(2, false)
	contextElement2 := asn1.NewStructured(contextTag2)
	contextElement2.Add(asn1.NewOctetStringFromString("Another optional field"))

	// Create a sequence with context-specific elements
	mainSeq := asn1.NewSequence()
	mainSeq.Add(asn1.NewInteger(1))
	mainSeq.Add(asn1.NewUTF8String("Required field"))
	mainSeq.Add(contextElement0)
	mainSeq.Add(contextElement1)
	mainSeq.Add(contextElement2)

	fmt.Println("Sequence with context-specific tags:")
	fmt.Println(mainSeq.CompactString())
}

func demonstrateObjectIdentifiers() {
	fmt.Println("\n5. Object Identifiers")
	fmt.Println("--------------------")

	// Common OIDs
	oids := []string{
		"1.2.840.113549.1.1.1",    // RSA encryption
		"1.2.840.113549.1.1.11",   // SHA-256 with RSA encryption
		"2.5.4.3",                 // Common Name
		"2.5.4.6",                 // Country Name
		"2.5.4.10",                // Organization Name
		"1.3.6.1.4.1.311.60.2.1.3", // Microsoft OID
	}

	descriptions := []string{
		"RSA Encryption",
		"SHA-256 with RSA",
		"Common Name (CN)",
		"Country Name (C)",
		"Organization Name (O)",
		"Microsoft Certificate Extension",
	}

	for i, oidStr := range oids {
		oid, err := asn1.NewObjectIdentifierFromString(oidStr)
		if err != nil {
			log.Printf("Failed to create OID %s: %v", oidStr, err)
			continue
		}

		fmt.Printf("%-30s %s\n", descriptions[i]+":", oid.String())

		// Show encoding size
		encoded, err := oid.Encode()
		if err != nil {
			log.Printf("Failed to encode OID: %v", err)
		} else {
			fmt.Printf("  Encoded: %02X (%d bytes)\n", encoded, len(encoded))
		}
	}
}

func demonstrateRoundTrip() {
	fmt.Println("\n6. Encoding/Decoding Round Trip")
	fmt.Println("-------------------------------")

	// Create a complex structure
	original := asn1.NewSequence()
	original.Add(asn1.NewInteger(42))
	original.Add(asn1.NewUTF8String("Test String"))
	original.Add(asn1.NewBoolean(true))

	innerSeq := asn1.NewSequence()
	innerSeq.Add(asn1.NewOctetStringFromString("inner data"))
	innerSeq.Add(asn1.NewInteger(-999))
	original.Add(innerSeq)

	oid, _ := asn1.NewObjectIdentifierFromString("1.2.3.4.5")
	original.Add(oid)

	bitString := asn1.NewBitStringFromBits("11010010")
	original.Add(bitString)

	fmt.Println("Original structure:")
	fmt.Println(original.CompactString())

	// Encode
	encoded, err := original.Encode()
	if err != nil {
		log.Printf("Encoding failed: %v", err)
		return
	}

	fmt.Printf("\nEncoded to %d bytes\n", len(encoded))
	fmt.Printf("Encoded data: %02X\n", encoded)

	// Decode back to generic ASN.1 values
	objects, err := asn1.DecodeAll(encoded)
	if err != nil {
		log.Printf("Decoding failed: %v", err)
		return
	}

	fmt.Printf("\nDecoded %d top-level objects:\n", len(objects))
	for i, obj := range objects {
		fmt.Printf("  [%d] %s\n", i, obj.String())
	}

	// Show round-trip integrity
	if len(objects) == 1 {
		reencoded, err := objects[0].Encode()
		if err != nil {
			log.Printf("Re-encoding failed: %v", err)
			return
		}

		fmt.Printf("\nRound-trip integrity check:\n")
		fmt.Printf("Original size:  %d bytes\n", len(encoded))
		fmt.Printf("Re-encoded size: %d bytes\n", len(reencoded))
		
		if len(encoded) == len(reencoded) {
			identical := true
			for i := range encoded {
				if encoded[i] != reencoded[i] {
					identical = false
					break
				}
			}
			if identical {
				fmt.Println("✓ Round-trip successful - data is identical")
			} else {
				fmt.Println("✗ Round-trip failed - data differs")
			}
		} else {
			fmt.Println("✗ Round-trip failed - size differs")
		}
	}
}

func demonstrateNewTypes() {
	fmt.Println("\n7. New ASN.1 Types")
	fmt.Println("------------------")

	// CHOICE type demonstration
	fmt.Println("\nCHOICE Type:")
	
	// Create different choices
	boolChoice := asn1.NewChoiceWithID(asn1.NewBoolean(true), "boolean_option")
	intChoice := asn1.NewChoiceWithID(asn1.NewInteger(42), "integer_option")
	stringChoice := asn1.NewChoiceWithID(asn1.NewUTF8String("Hello Choice"), "string_option")
	
	choices := []*asn1.ASN1Choice{boolChoice, intChoice, stringChoice}
	for _, choice := range choices {
		encoded, err := choice.Encode()
		if err != nil {
			log.Printf("CHOICE encoding failed: %v", err)
			continue
		}
		fmt.Printf("  %s -> %d bytes\n", choice.String(), len(encoded))
	}

	// ENUMERATED type demonstration
	fmt.Println("\nENUMERATED Type:")
	
	statusOK := asn1.NewEnumeratedWithName(0, "OK")
	statusWarning := asn1.NewEnumeratedWithName(1, "WARNING")
	statusError := asn1.NewEnumeratedWithName(2, "ERROR")
	statusCritical := asn1.NewEnumeratedWithName(3, "CRITICAL")
	
	enums := []*asn1.ASN1Enumerated{statusOK, statusWarning, statusError, statusCritical}
	for _, enum := range enums {
		encoded, err := enum.Encode()
		if err != nil {
			log.Printf("ENUMERATED encoding failed: %v", err)
			continue
		}
		fmt.Printf("  %s -> %d bytes\n", enum.String(), len(encoded))
	}

	// Time types demonstration
	fmt.Println("\nTime Types:")
	
	// UTCTime
	now := time.Now()
	utcTime := asn1.NewUTCTime(now)
	utcEncoded, err := utcTime.Encode()
	if err != nil {
		log.Printf("UTCTime encoding failed: %v", err)
	} else {
		fmt.Printf("  %s -> %d bytes\n", utcTime.String(), len(utcEncoded))
	}
	
	// GeneralizedTime
	genTime := asn1.NewGeneralizedTime(now)
	genEncoded, err := genTime.Encode()
	if err != nil {
		log.Printf("GeneralizedTime encoding failed: %v", err)
	} else {
		fmt.Printf("  %s -> %d bytes\n", genTime.String(), len(genEncoded))
	}

	// Demonstrate round-trip with new types
	fmt.Println("\nRound-trip test with new types:")
	
	// Test ENUMERATED round-trip
	originalEnum := asn1.NewEnumeratedWithName(42, "answer")
	enumEncoded, err := originalEnum.Encode()
	if err != nil {
		log.Printf("ENUMERATED encoding failed: %v", err)
	} else {
		decodedEnum, consumed, err := asn1.DecodeEnumerated(enumEncoded)
		if err != nil {
			log.Printf("ENUMERATED decoding failed: %v", err)
		} else if consumed != len(enumEncoded) {
			fmt.Printf("  ✗ ENUMERATED: consumed %d bytes, expected %d\n", consumed, len(enumEncoded))
		} else if decodedEnum.Int64() != originalEnum.Int64() {
			fmt.Printf("  ✗ ENUMERATED: value mismatch\n")
		} else {
			fmt.Printf("  ✓ ENUMERATED round-trip successful\n")
		}
	}
	
	// Test UTCTime round-trip
	testTime := time.Date(2023, 12, 25, 14, 30, 45, 0, time.UTC)
	originalUTC := asn1.NewUTCTime(testTime)
	utcEncoded2, err := originalUTC.Encode()
	if err != nil {
		log.Printf("UTCTime encoding failed: %v", err)
	} else {
		decodedUTC, consumed, err := asn1.DecodeUTCTime(utcEncoded2)
		if err != nil {
			log.Printf("UTCTime decoding failed: %v", err)
		} else if consumed != len(utcEncoded2) {
			fmt.Printf("  ✗ UTCTime: consumed %d bytes, expected %d\n", consumed, len(utcEncoded2))
		} else if !decodedUTC.Time().Equal(originalUTC.Time()) {
			fmt.Printf("  ✗ UTCTime: time mismatch\n")
		} else {
			fmt.Printf("  ✓ UTCTime round-trip successful\n")
		}
	}

	fmt.Println("\nLibrary now supports all essential ASN.1 types including CHOICE!")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func demonstrateComplexHierarchy() {
	fmt.Println("\n8. Complex Hierarchical Structure Example")
	fmt.Println("==========================================")

	// Create a complex nested structure that demonstrates the library's capabilities
	// This represents a simplified X.509-like certificate structure with choices and enumerations

	// Certificate Information with CHOICE and ENUMERATED types
	certStatus := asn1.NewEnumeratedWithName(0, "VALID")
	certAlgorithm := asn1.NewChoiceWithID(asn1.NewObjectIdentifierFromStringUnchecked("1.2.840.113549.1.1.11"), "sha256WithRSAEncryption")
	
	// Time validity period
	notBefore := asn1.NewUTCTime(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC))
	notAfter := asn1.NewGeneralizedTime(time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC))
	
	// Subject information (nested sequences)
	subjectName := asn1.NewSequence()
	subjectRDNs := asn1.NewSequence()
	subjectName.Add(subjectRDNs)
		// Add country
		countrySeq := asn1.NewSequence()
		countrySeq.Add(asn1.NewObjectIdentifierFromStringUnchecked("2.5.4.6")) // countryName OID
		countrySeq.Add(asn1.NewPrintableString("US"))
		subjectRDNs.Add(countrySeq)
		
		// Add organization  
		orgSeq := asn1.NewSequence()
		orgSeq.Add(asn1.NewObjectIdentifierFromStringUnchecked("2.5.4.10")) // organizationName OID
		orgSeq.Add(asn1.NewUTF8String("Example Corp"))
		subjectRDNs.Add(orgSeq)
		
		// Add common name
		cnSeq := asn1.NewSequence()
		cnSeq.Add(asn1.NewObjectIdentifierFromStringUnchecked("2.5.4.3")) // commonName OID
		cnSeq.Add(asn1.NewUTF8String("www.example.com"))
		subjectRDNs.Add(cnSeq)

	// Extensions with context-specific tags and choices
	extensions := asn1.NewSequence()
	
	// Key Usage extension (critical)
	keyUsageExt := asn1.NewSequence()
	keyUsageExt.Add(asn1.NewObjectIdentifierFromStringUnchecked("2.5.29.15")) // keyUsage OID
	keyUsageExt.Add(asn1.NewBoolean(true)) // critical
	keyUsageBits := asn1.NewBitStringFromBits("11000000") // digitalSignature + keyEncipherment
	keyUsageExt.Add(asn1.NewOctetString(keyUsageBits.Value()))
	extensions.Add(keyUsageExt)
	
	// Subject Alternative Names extension
	sanExt := asn1.NewSequence()
	sanExt.Add(asn1.NewObjectIdentifierFromStringUnchecked("2.5.29.17")) // subjectAltName OID
	sanExt.Add(asn1.NewBoolean(false)) // not critical
	
	// SAN values (with context-specific tags)
	sanValues := asn1.NewSequence()
	
	// DNS names [2] IMPLICIT IA5String
	dnsTag := asn1.NewContextSpecificTag(2, false)
	dnsName1 := asn1.NewStructured(dnsTag)
	dnsName1.Add(asn1.NewIA5String("www.example.com"))
	sanValues.Add(dnsName1)
	
	dnsName2 := asn1.NewStructured(dnsTag)
	dnsName2.Add(asn1.NewIA5String("example.com"))
	sanValues.Add(dnsName2)
	
	// IP address [7] IMPLICIT OCTET STRING
	ipTag := asn1.NewContextSpecificTag(7, false)
	ipAddr := asn1.NewStructured(ipTag)
	ipAddr.Add(asn1.NewOctetString([]byte{192, 168, 1, 100}))
	sanValues.Add(ipAddr)
	
	sanExt.Add(asn1.NewOctetString(func() []byte {
		encoded, _ := sanValues.Encode()
		return encoded
	}()))
	extensions.Add(sanExt)

	// Main certificate structure
	certificate := asn1.NewSequence()
	
	// TBS Certificate (To Be Signed)
	tbsCert := asn1.NewSequence()
	
	// Version [0] EXPLICIT INTEGER DEFAULT v1
	versionTag := asn1.NewContextSpecificTag(0, true)
	version := asn1.NewStructured(versionTag)
	version.Add(asn1.NewInteger(2)) // v3
	tbsCert.Add(version)
	
	// Serial number
	tbsCert.Add(asn1.NewInteger(123456789))
	
	// Signature algorithm
	sigAlgSeq := asn1.NewSequence()
	sigAlgSeq.Add(asn1.NewObjectIdentifierFromStringUnchecked("1.2.840.113549.1.1.11"))
	sigAlgSeq.Add(asn1.NewNull())
	tbsCert.Add(sigAlgSeq)
	
	// Issuer (same as subject for self-signed)
	tbsCert.Add(subjectName)
	
	// Validity period
	validity := asn1.NewSequence()
	validity.Add(notBefore)
	validity.Add(notAfter)
	tbsCert.Add(validity)
	
	// Subject
	tbsCert.Add(subjectName)
	
	// Subject Public Key Info
	spki := asn1.NewSequence()
	spkiAlg := asn1.NewSequence()
	spkiAlg.Add(asn1.NewObjectIdentifierFromStringUnchecked("1.2.840.113549.1.1.1")) // RSA
	spkiAlg.Add(asn1.NewNull())
	spki.Add(spkiAlg)
	
	// Mock public key (normally this would be the actual key)
	publicKeyBits := asn1.NewBitStringFromBits("1100000000110000")
	spki.Add(publicKeyBits)
	tbsCert.Add(spki)
	
	// Extensions [3] EXPLICIT SEQUENCE OF Extension
	extTag := asn1.NewContextSpecificTag(3, true)
	extWrapper := asn1.NewStructured(extTag)
	extWrapper.Add(extensions)
	tbsCert.Add(extWrapper)
	
	// Add TBS certificate to main certificate
	certificate.Add(tbsCert)
	
	// Signature algorithm (again)
	certificate.Add(sigAlgSeq)
	
	// Signature value (mock)
	signature := asn1.NewBitStringFromBits("1111000011110000111100001111000011110000111100001111000011110000")
	certificate.Add(signature)

	// Add some metadata sequence with choices and enumerations
	metadata := asn1.NewSequence()
	metadata.Add(certStatus) // ENUMERATED status
	metadata.Add(certAlgorithm) // CHOICE algorithm
	
	// Trust level with nested choices
	trustLevelChoice := asn1.NewChoiceWithID(asn1.NewEnumeratedWithName(2, "HIGH_TRUST"), "trust_level")
	metadata.Add(trustLevelChoice)
	
	// Timestamps
	issuedAt := asn1.NewUTCTimeNow()
	lastVerified := asn1.NewGeneralizedTimeNow()
	metadata.Add(issuedAt)
	metadata.Add(lastVerified)

	// Final wrapper structure
	finalStructure := asn1.NewSequence()
	finalStructure.Add(certificate)
	finalStructure.Add(metadata)

	fmt.Println("Complex Hierarchical Structure:")
	fmt.Println("-------------------------------")
	fmt.Println("Structure: X.509-like Certificate with Extensions, Choices, and Enumerations")
	fmt.Println()
	fmt.Printf("%s\n", finalStructure.String())
	
	fmt.Println("\nCompact View:")
	fmt.Printf("%s\n", finalStructure.CompactString())

	// Encoding test
	encoded, err := finalStructure.Encode()
	if err != nil {
		log.Printf("Encoding failed: %v", err)
	} else {
		fmt.Printf("\nEncoded to %d bytes\n", len(encoded))
		fmt.Printf("First 128 bytes: %02X\n", encoded[:min(128, len(encoded))])
		
		// Decode back to verify round-trip
		decoded, err := asn1.DecodeAll(encoded)
		if err != nil {
			log.Printf("Decoding failed: %v", err)
		} else {
			fmt.Printf("Successfully decoded %d top-level objects\n", len(decoded))
			
			// Show the structure depth
			fmt.Println("\nStructure Analysis:")
			analyzeStructureDepth(finalStructure, 0)
		}
	}
}

func analyzeStructureDepth(element asn1.ASN1Object, depth int) {
	indent := strings.Repeat("  ", depth)
	
	switch v := element.(type) {
	case *asn1.ASN1Structured:
		// Create a simple tag description since Tag doesn't have String()
		tagDesc := fmt.Sprintf("Tag{class=%d, number=%d, constructed=%t}", 
			v.Tag().Class, v.Tag().Number, v.Tag().Constructed)
		elementCount := len(v.Elements())
		fmt.Printf("%s%s (%d elements)\n", indent, tagDesc, elementCount)
		for _, elem := range v.Elements() {
			analyzeStructureDepth(elem, depth+1)
		}
	case *asn1.ASN1Choice:
		fmt.Printf("%s%s -> %s\n", indent, "CHOICE", v.String())
		analyzeStructureDepth(v.Value(), depth+1)
	case *asn1.ASN1Enumerated:
		fmt.Printf("%s%s\n", indent, v.String())
	case *asn1.ASN1UTCTime:
		fmt.Printf("%s%s\n", indent, v.String())
	case *asn1.ASN1GeneralizedTime:
		fmt.Printf("%s%s\n", indent, v.String())
	default:
		fmt.Printf("%s%s\n", indent, element.String())
	}
}