package examples

import (
	"time"

	"github.com/limecodeswe/asn1"
)

// PersonWithTags demonstrates struct tags for ASN.1 encoding
type PersonWithTags struct {
	// Required fields with explicit types
	ID       int64  `asn1:"integer"`
	Name     string `asn1:"utf8string"`
	Email    string `asn1:"ia5string"`
	IsActive bool   `asn1:"boolean"`

	// Optional fields with context-specific tags
	Department   *string         `asn1:"printablestring,optional,tag:0"`
	PhoneNumber  *string         `asn1:"printablestring,optional,tag:1"`
	Birthday     *time.Time      `asn1:"utctime,optional,tag:2"`
	Salary       *int64          `asn1:"integer,optional,tag:3"`
	Manager      *PersonWithTags `asn1:"sequence,optional,tag:4"`
	Permissions  []string        `asn1:"sequence,optional,tag:5"`
	Metadata     []byte          `asn1:"octetstring,optional,tag:6"`
	EmployeeType *int            `asn1:"integer,optional,tag:7"`
}

// NewPersonWithTags creates a new Person using struct tags
func NewPersonWithTags(id int64, name, email string, isActive bool) *PersonWithTags {
	return &PersonWithTags{
		ID:       id,
		Name:     name,
		Email:    email,
		IsActive: isActive,
	}
}

// SetDepartment sets the optional department field
func (p *PersonWithTags) SetDepartment(dept string) *PersonWithTags {
	p.Department = &dept
	return p
}

// SetPhoneNumber sets the optional phone number field
func (p *PersonWithTags) SetPhoneNumber(phone string) *PersonWithTags {
	p.PhoneNumber = &phone
	return p
}

// SetBirthday sets the optional birthday field
func (p *PersonWithTags) SetBirthday(birthday time.Time) *PersonWithTags {
	p.Birthday = &birthday
	return p
}

// SetSalary sets the optional salary field
func (p *PersonWithTags) SetSalary(salary int64) *PersonWithTags {
	p.Salary = &salary
	return p
}

// SetManager sets the optional manager field
func (p *PersonWithTags) SetManager(manager *PersonWithTags) *PersonWithTags {
	p.Manager = manager
	return p
}

// SetPermissions sets the optional permissions field
func (p *PersonWithTags) SetPermissions(permissions []string) *PersonWithTags {
	p.Permissions = make([]string, len(permissions))
	copy(p.Permissions, permissions)
	return p
}

// SetMetadata sets the optional metadata field
func (p *PersonWithTags) SetMetadata(metadata []byte) *PersonWithTags {
	p.Metadata = make([]byte, len(metadata))
	copy(p.Metadata, metadata)
	return p
}

// SetEmployeeType sets the optional employee type field
func (p *PersonWithTags) SetEmployeeType(empType int) *PersonWithTags {
	p.EmployeeType = &empType
	return p
}

// MarshalASN1 encodes the PersonWithTags as ASN.1 using struct tags
func (p *PersonWithTags) MarshalASN1() ([]byte, error) {
	return asn1.Marshal(p)
}

// UnmarshalASN1 decodes ASN.1 data into PersonWithTags using struct tags
func (p *PersonWithTags) UnmarshalASN1(data []byte) error {
	return asn1.Unmarshal(data, p)
}

// Example demonstrates the usage of struct tags
func ExampleStructTags() (*PersonWithTags, []byte, error) {
	// Create a person with required fields
	person := NewPersonWithTags(123, "Alice Johnson", "alice@example.com", true)

	// Add optional fields using method chaining
	person.SetDepartment("Engineering").
		SetPhoneNumber("+1-555-0123").
		SetSalary(75000).
		SetEmployeeType(0). // 0=full-time, 1=part-time, 2=contractor
		SetPermissions([]string{"read", "write", "admin"})

	// Set birthday
	birthday := time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC)
	person.SetBirthday(birthday)

	// Add some metadata
	person.SetMetadata([]byte("employee-record-v1"))

	// Create a manager relationship (recursive structure)
	manager := NewPersonWithTags(1, "Bob Smith", "bob@example.com", true)
	manager.SetDepartment("Management")
	person.SetManager(manager)

	// Encode using struct tags - much simpler than manual encoding!
	encoded, err := person.MarshalASN1()
	if err != nil {
		return nil, nil, err
	}

	return person, encoded, nil
}

// ExampleRoundTrip demonstrates encoding and decoding with struct tags
func ExampleRoundTrip() error {
	// Create and encode a person
	_, encoded, err := ExampleStructTags()
	if err != nil {
		return err
	}

	// Decode back into a new struct
	var decoded PersonWithTags
	if err := decoded.UnmarshalASN1(encoded); err != nil {
		return err
	}

	// The decoded struct should match the original
	// (You would implement comparison logic here)

	return nil
}

// Alternative approach: using embedded struct tags for different ASN.1 schemes
type PersonV2 struct {
	// Required fields
	ID       int64  `asn1:"integer"`
	Name     string `asn1:"utf8string"`
	Email    string `asn1:"ia5string"`
	IsActive bool   `asn1:"boolean"`

	// Optional fields with different tagging scheme
	Department  *string    `asn1:"printablestring,optional,tag:10"` // Different tag numbers
	PhoneNumber *string    `asn1:"printablestring,optional,tag:11"`
	Birthday    *time.Time `asn1:"generalizedtime,optional,tag:12"` // Different time type
	Salary      *int64     `asn1:"integer,optional,tag:13"`

	// Fields that can be omitted entirely
	ExtraInfo string `asn1:"utf8string,omitempty"`
	TempData  []byte `asn1:"-"` // Ignored field
}

// Company represents a company with employees using struct tags
type Company struct {
	Name      string            `asn1:"utf8string"`
	ID        int64             `asn1:"integer"`
	Founded   time.Time         `asn1:"generalizedtime"`
	Employees []*PersonWithTags `asn1:"sequence"` // SEQUENCE OF Person
	Active    bool              `asn1:"boolean"`
}

// MarshalASN1 encodes the Company as ASN.1 using struct tags
func (c *Company) MarshalASN1() ([]byte, error) {
	return asn1.Marshal(c)
}

// UnmarshalASN1 decodes ASN.1 data into Company using struct tags
func (c *Company) UnmarshalASN1(data []byte) error {
	return asn1.Unmarshal(data, c)
}
