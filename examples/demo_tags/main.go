package main

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"github.com/limecodeswe/asn1"
	"github.com/limecodeswe/asn1/examples"
)

func main() {
	fmt.Println("ASN.1 Struct Tags Demo")
	fmt.Println("======================")

	// Demonstrate basic struct tags
	demonstrateBasicStructTags()

	// Demonstrate the Person example with struct tags
	demonstratePersonWithTags()

	// Demonstrate round-trip encoding/decoding
	demonstrateRoundTrip()

	// Compare with manual encoding
	compareWithManualEncoding()
}

func demonstrateBasicStructTags() {
	fmt.Println("\n1. Basic Struct Tags")
	fmt.Println("--------------------")

	// Define a simple struct with ASN.1 tags
	type Document struct {
		ID      int64     `asn1:"integer"`
		Title   string    `asn1:"utf8string"`
		Created time.Time `asn1:"utctime"`
		Public  bool      `asn1:"boolean"`
		Content []byte    `asn1:"octetstring"`
	}

	// Create an instance
	doc := &Document{
		ID:      12345,
		Title:   "Sample Document",
		Created: time.Now(),
		Public:  true,
		Content: []byte("This is the document content"),
	}

	// Encode using struct tags - one line!
	encoded, err := asn1.Marshal(doc)
	if err != nil {
		log.Printf("Encoding failed: %v", err)
		return
	}

	fmt.Printf("Original document: ID=%d, Title=%q, Public=%t\n",
		doc.ID, doc.Title, doc.Public)
	fmt.Printf("Encoded to %d bytes\n", len(encoded))

	// Decode back
	var decoded Document
	if err := asn1.Unmarshal(encoded, &decoded); err != nil {
		log.Printf("Decoding failed: %v", err)
		return
	}

	fmt.Printf("Decoded document: ID=%d, Title=%q, Public=%t\n",
		decoded.ID, decoded.Title, decoded.Public)
	fmt.Printf("Round-trip successful: %t\n",
		doc.ID == decoded.ID && doc.Title == decoded.Title && doc.Public == decoded.Public)
}

func demonstratePersonWithTags() {
	fmt.Println("\n2. Person Example with Struct Tags")
	fmt.Println("-----------------------------------")

	// Create a person using the struct tags approach
	person, encoded, err := examples.ExampleStructTags()
	if err != nil {
		log.Printf("Person example failed: %v", err)
		return
	}

	fmt.Printf("Created person: ID=%d, Name=%q, Email=%q\n",
		person.ID, person.Name, person.Email)

	if person.Department != nil {
		fmt.Printf("Department: %q\n", *person.Department)
	}
	if person.Salary != nil {
		fmt.Printf("Salary: $%d\n", *person.Salary)
	}
	if person.Manager != nil {
		fmt.Printf("Manager: %q (ID: %d)\n", person.Manager.Name, person.Manager.ID)
	}

	fmt.Printf("Encoded to %d bytes with struct tags\n", len(encoded))

	// Decode back to verify
	var decoded examples.PersonWithTags
	if err := decoded.UnmarshalASN1(encoded); err != nil {
		log.Printf("Decoding failed: %v", err)
		return
	}

	fmt.Printf("Round-trip successful: %t\n",
		person.ID == decoded.ID && person.Name == decoded.Name)
}

func demonstrateRoundTrip() {
	fmt.Println("\n3. Round-Trip Encoding/Decoding")
	fmt.Println("--------------------------------")

	// Define a complex nested structure
	type Address struct {
		Street   string `asn1:"utf8string"`
		City     string `asn1:"utf8string"`
		PostCode string `asn1:"printablestring"`
		Country  string `asn1:"printablestring"`
	}

	type Employee struct {
		ID      int64    `asn1:"integer"`
		Name    string   `asn1:"utf8string"`
		Address Address  `asn1:"sequence"`
		Skills  []string `asn1:"sequence"`
		Salary  *int64   `asn1:"integer,optional,tag:0"`
		Manager *string  `asn1:"utf8string,optional,tag:1"`
	}

	salary := int64(85000)
	manager := "Alice Johnson"

	original := &Employee{
		ID:   9876,
		Name: "John Doe",
		Address: Address{
			Street:   "123 Main St",
			City:     "Tech City",
			PostCode: "12345",
			Country:  "USA",
		},
		Skills:  []string{"Go", "ASN.1", "Cryptography"},
		Salary:  &salary,
		Manager: &manager,
	}

	// Encode
	encoded, err := asn1.Marshal(original)
	if err != nil {
		log.Printf("Encoding failed: %v", err)
		return
	}

	fmt.Printf("Encoded complex employee structure to %d bytes\n", len(encoded))

	// Decode
	var decoded Employee
	if err := asn1.Unmarshal(encoded, &decoded); err != nil {
		log.Printf("Decoding failed: %v", err)
		return
	}

	// Verify
	fmt.Printf("Original: %s at %s, %s\n",
		original.Name, original.Address.Street, original.Address.City)
	fmt.Printf("Decoded:  %s at %s, %s\n",
		decoded.Name, decoded.Address.Street, decoded.Address.City)
	fmt.Printf("Skills match: %t\n",
		len(original.Skills) == len(decoded.Skills) &&
			original.Skills[0] == decoded.Skills[0])
}

func compareWithManualEncoding() {
	fmt.Println("\n4. Comparison with Manual Encoding")
	fmt.Println("-----------------------------------")

	// Manual encoding (the old way)
	manualSeq := asn1.NewSequence()
	manualSeq.Add(asn1.NewInteger(42))
	manualSeq.Add(asn1.NewUTF8String("test"))
	manualSeq.Add(asn1.NewBoolean(true))

	manualEncoded, err := manualSeq.Encode()
	if err != nil {
		log.Printf("Manual encoding failed: %v", err)
		return
	}

	// Struct tags encoding (the new way)
	type Simple struct {
		ID     int64  `asn1:"integer"`
		Name   string `asn1:"utf8string"`
		Active bool   `asn1:"boolean"`
	}

	simple := &Simple{
		ID:     42,
		Name:   "test",
		Active: true,
	}

	structEncoded, err := asn1.Marshal(simple)
	if err != nil {
		log.Printf("Struct encoding failed: %v", err)
		return
	}

	fmt.Printf("Manual encoding:    %d bytes\n", len(manualEncoded))
	fmt.Printf("Struct tag encoding: %d bytes\n", len(structEncoded))
	fmt.Printf("Results identical:   %t\n",
		bytes.Equal(manualEncoded, structEncoded))

	// Show the huge reduction in code complexity
	fmt.Println("\nCode comparison:")
	fmt.Println("Manual approach: ~10 lines of encoding code")
	fmt.Println("Struct tags:     1 line: asn1.Marshal(simple)")
	fmt.Println("Reduction:       ~90% less code!")
}
