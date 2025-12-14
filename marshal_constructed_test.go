package asn1

import (
	"testing"
)

func TestUnmarshalConstructedContextSpecific(t *testing.T) {
	type Address struct {
		Street string `asn1:"utf8string"`
		City   string `asn1:"utf8string"`
	}

	type Person struct {
		Name    string   `asn1:"utf8string"`
		Address *Address `asn1:"sequence,optional,tag:0"` // Context-specific constructed
	}

	tests := []struct {
		name    string
		input   *Person
		wantErr bool
	}{
		{
			name: "with optional nested sequence",
			input: &Person{
				Name: "John",
				Address: &Address{
					Street: "123 Main St",
					City:   "Springfield",
				},
			},
			wantErr: false,
		},
		{
			name: "without optional nested sequence",
			input: &Person{
				Name:    "Jane",
				Address: nil,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal
			encoded, err := Marshal(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Marshal() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			// Unmarshal
			var decoded Person
			err = Unmarshal(encoded, &decoded)
			if err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}

			// Verify
			if decoded.Name != tt.input.Name {
				t.Errorf("Name = %v, want %v", decoded.Name, tt.input.Name)
			}

			if tt.input.Address != nil {
				if decoded.Address == nil {
					t.Fatal("Address is nil, expected non-nil")
				}
				if decoded.Address.Street != tt.input.Address.Street {
					t.Errorf("Address.Street = %v, want %v", decoded.Address.Street, tt.input.Address.Street)
				}
				if decoded.Address.City != tt.input.Address.City {
					t.Errorf("Address.City = %v, want %v", decoded.Address.City, tt.input.Address.City)
				}
			} else {
				if decoded.Address != nil {
					t.Error("Address should be nil")
				}
			}
		})
	}
}

func TestUnmarshalNestedConstructed(t *testing.T) {
	type Inner struct {
		Value int64 `asn1:"integer"`
	}

	type Middle struct {
		Name  string `asn1:"utf8string"`
		Inner *Inner `asn1:"sequence,optional,tag:0"`
	}

	type Outer struct {
		ID     int64   `asn1:"integer"`
		Middle *Middle `asn1:"sequence,optional,tag:1"`
	}

	input := &Outer{
		ID: 123,
		Middle: &Middle{
			Name: "middle",
			Inner: &Inner{
				Value: 456,
			},
		},
	}

	// Marshal
	encoded, err := Marshal(input)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// Unmarshal
	var decoded Outer
	err = Unmarshal(encoded, &decoded)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	// Verify
	if decoded.ID != input.ID {
		t.Errorf("ID = %v, want %v", decoded.ID, input.ID)
	}
	if decoded.Middle == nil {
		t.Fatal("Middle is nil")
	}
	if decoded.Middle.Name != input.Middle.Name {
		t.Errorf("Middle.Name = %v, want %v", decoded.Middle.Name, input.Middle.Name)
	}
	if decoded.Middle.Inner == nil {
		t.Fatal("Middle.Inner is nil")
	}
	if decoded.Middle.Inner.Value != input.Middle.Inner.Value {
		t.Errorf("Middle.Inner.Value = %v, want %v", decoded.Middle.Inner.Value, input.Middle.Inner.Value)
	}
}

func TestUnmarshalSequenceOf(t *testing.T) {
	type Item struct {
		ID   int64  `asn1:"integer"`
		Name string `asn1:"utf8string"`
	}

	type Container struct {
		Items []Item `asn1:"sequence"`
	}

	input := &Container{
		Items: []Item{
			{ID: 1, Name: "first"},
			{ID: 2, Name: "second"},
			{ID: 3, Name: "third"},
		},
	}

	// Marshal
	encoded, err := Marshal(input)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// Unmarshal
	var decoded Container
	err = Unmarshal(encoded, &decoded)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	// Verify
	if len(decoded.Items) != len(input.Items) {
		t.Fatalf("len(Items) = %v, want %v", len(decoded.Items), len(input.Items))
	}
	for i, item := range decoded.Items {
		if item.ID != input.Items[i].ID {
			t.Errorf("Items[%d].ID = %v, want %v", i, item.ID, input.Items[i].ID)
		}
		if item.Name != input.Items[i].Name {
			t.Errorf("Items[%d].Name = %v, want %v", i, item.Name, input.Items[i].Name)
		}
	}
}
