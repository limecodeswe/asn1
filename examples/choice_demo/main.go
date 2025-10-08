package main

import (
	"fmt"
	"log"
	"time"

	"github.com/limecodeswe/asn1"
)

func main() {
	fmt.Println("ASN.1 CHOICE with Struct Tags Demo")
	fmt.Println("===================================")

	// Demo 1: interface{} approach for simple choices
	demonstrateInterfaceChoice()

	// Demo 2: Union struct approach for complex choices
	demonstrateUnionChoice()

	// Demo 3: Manual ASN1Choice integration
	demonstrateManualChoice()
}

func demonstrateInterfaceChoice() {
	fmt.Println("\n1. Interface{} Choice Approach")
	fmt.Println("------------------------------")

	type Message struct {
		ID      int64       `asn1:"integer"`
		Content interface{} `asn1:"choice"` // Can hold any type
	}

	// Test with different types
	examples := []struct {
		name    string
		content interface{}
	}{
		{"boolean", true},
		{"integer", int64(42)},
		{"string", "hello world"},
		{"bytes", []byte("binary data")},
		{"time", time.Now()},
	}

	for _, example := range examples {
		fmt.Printf("Testing %s choice...\n", example.name)

		msg := &Message{
			ID:      123,
			Content: example.content,
		}

		// Marshal
		encoded, err := asn1.Marshal(msg)
		if err != nil {
			log.Printf("Marshal failed: %v", err)
			continue
		}

		fmt.Printf("  Encoded %d bytes\n", len(encoded))

		// Unmarshal
		var decoded Message
		err = asn1.Unmarshal(encoded, &decoded)
		if err != nil {
			log.Printf("  Unmarshal failed: %v", err)
			continue
		}

		fmt.Printf("  Original:  %T = %v\n", msg.Content, msg.Content)
		fmt.Printf("  Decoded:   %T = %v\n", decoded.Content, decoded.Content)
		fmt.Printf("  Success: %t\n", fmt.Sprintf("%v", msg.Content) == fmt.Sprintf("%v", decoded.Content))
	}
}

func demonstrateUnionChoice() {
	fmt.Println("\n2. Union Struct Choice Approach")
	fmt.Println("--------------------------------")

	// Define a choice struct with all possible alternatives
	type MessageChoice struct {
		BoolValue   *bool      `asn1:"boolean,choice:bool,tag:0"`
		IntValue    *int64     `asn1:"integer,choice:int,tag:1"`
		StringValue *string    `asn1:"utf8string,choice:string,tag:2"`
		BytesValue  *[]byte    `asn1:"octetstring,choice:bytes,tag:3"`
		TimeValue   *time.Time `asn1:"utctime,choice:time,tag:4"`
	}

	type Document struct {
		ID       int64         `asn1:"integer"`
		Title    string        `asn1:"utf8string"`
		Metadata MessageChoice `asn1:"choice"`
	}

	// Test with boolean choice
	boolVal := true
	doc1 := &Document{
		ID:    1,
		Title: "Boolean Document",
		Metadata: MessageChoice{
			BoolValue: &boolVal,
		},
	}

	// Test with string choice
	stringVal := "important"
	doc2 := &Document{
		ID:    2,
		Title: "String Document",
		Metadata: MessageChoice{
			StringValue: &stringVal,
		},
	}

	// Test with integer choice
	intVal := int64(42)
	doc3 := &Document{
		ID:    3,
		Title: "Integer Document",
		Metadata: MessageChoice{
			IntValue: &intVal,
		},
	}

	documents := []*Document{doc1, doc2, doc3}
	names := []string{"boolean", "string", "integer"}

	for i, doc := range documents {
		fmt.Printf("Testing %s union choice...\n", names[i])

		encoded, err := asn1.Marshal(doc)
		if err != nil {
			log.Printf("  Marshal failed: %v", err)
			continue
		}

		fmt.Printf("  Encoded %d bytes\n", len(encoded))

		var decoded Document
		err = asn1.Unmarshal(encoded, &decoded)
		if err != nil {
			log.Printf("  Unmarshal failed: %v", err)
			continue
		}

		fmt.Printf("  Round-trip successful for %s choice\n", names[i])
	}
}

func demonstrateManualChoice() {
	fmt.Println("\n3. Manual ASN1Choice Integration")
	fmt.Println("---------------------------------")

	type Document struct {
		ID       int64            `asn1:"integer"`
		Title    string           `asn1:"utf8string"`
		Metadata *asn1.ASN1Choice // Direct use of ASN1Choice
	}

	// Create choices manually
	choices := []*asn1.ASN1Choice{
		asn1.NewChoiceWithID(asn1.NewBoolean(true), "approved"),
		asn1.NewChoiceWithID(asn1.NewInteger(42), "priority"),
		asn1.NewChoiceWithID(asn1.NewUTF8String("confidential"), "classification"),
	}

	names := []string{"boolean", "integer", "string"}

	for i, choice := range choices {
		fmt.Printf("Testing %s manual choice...\n", names[i])

		doc := &Document{
			ID:       int64(100 + i),
			Title:    fmt.Sprintf("Document %d", i+1),
			Metadata: choice,
		}

		encoded, err := asn1.Marshal(doc)
		if err != nil {
			log.Printf("  Marshal failed: %v", err)
			continue
		}

		fmt.Printf("  Encoded %d bytes\n", len(encoded))

		var decoded Document
		err = asn1.Unmarshal(encoded, &decoded)
		if err != nil {
			log.Printf("  Unmarshal failed: %v", err)
			continue
		}

		fmt.Printf("  Original choice: %s\n", choice.String())
		if decoded.Metadata != nil {
			fmt.Printf("  Decoded choice:  %s\n", decoded.Metadata.String())
		} else {
			fmt.Printf("  Decoded choice:  <nil>\n")
		}
	}
}
