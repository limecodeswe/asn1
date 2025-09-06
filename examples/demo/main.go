package main

import (
	"fmt"
	"log"
	"time"

	"github.com/limecodeswe/asn1"
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
	manager := asn1.NewPerson(1, "Alice Johnson", "alice.johnson@company.com", true).
		SetDepartment("Engineering").
		SetPhoneNumber("+1-555-0100").
		SetSalary(120000).
		SetEmployeeType(0). // full-time
		SetPermissions([]string{"admin", "read", "write", "delete"})

	// Create an employee with the manager
	employee := asn1.NewPerson(123, "John Doe", "john.doe@company.com", true).
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
	directory := asn1.NewPersonDirectory()
	directory.AddPerson(manager)
	directory.AddPerson(employee)

	// Add some more employees
	employee2 := asn1.NewPerson(124, "Jane Smith", "jane.smith@company.com", true).
		SetDepartment("Sales").
		SetEmployeeType(1). // part-time
		SetPermissions([]string{"read"})

	employee3 := asn1.NewPerson(125, "Bob Wilson", "bob.wilson@company.com", false).
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}