package asn1

import (
	"fmt"
	"time"
)

// Person represents a realistic ASN.1 structure for a person
// This demonstrates how to use the library for real-world applications
type Person struct {
	// Required fields
	ID       int64  // INTEGER
	Name     string // UTF8String
	Email    string // IA5String
	IsActive bool   // BOOLEAN

	// Optional fields (will be context-specific tagged)
	Department   *string    // [0] OPTIONAL PrintableString
	PhoneNumber  *string    // [1] OPTIONAL PrintableString  
	Birthday     *time.Time // [2] OPTIONAL UTCTime
	Salary       *int64     // [3] OPTIONAL INTEGER
	Manager      *Person    // [4] OPTIONAL Person (recursive structure)
	Permissions  []string   // [5] OPTIONAL SEQUENCE OF PrintableString
	Metadata     []byte     // [6] OPTIONAL OCTET STRING
	EmployeeType *int       // [7] OPTIONAL INTEGER (0=full-time, 1=part-time, 2=contractor)
}

// NewPerson creates a new Person with required fields
func NewPerson(id int64, name, email string, isActive bool) *Person {
	return &Person{
		ID:       id,
		Name:     name,
		Email:    email,
		IsActive: isActive,
	}
}

// SetDepartment sets the optional department field
func (p *Person) SetDepartment(dept string) *Person {
	p.Department = &dept
	return p
}

// SetPhoneNumber sets the optional phone number field  
func (p *Person) SetPhoneNumber(phone string) *Person {
	p.PhoneNumber = &phone
	return p
}

// SetBirthday sets the optional birthday field
func (p *Person) SetBirthday(birthday time.Time) *Person {
	p.Birthday = &birthday
	return p
}

// SetSalary sets the optional salary field
func (p *Person) SetSalary(salary int64) *Person {
	p.Salary = &salary
	return p
}

// SetManager sets the optional manager field
func (p *Person) SetManager(manager *Person) *Person {
	p.Manager = manager
	return p
}

// SetPermissions sets the optional permissions field
func (p *Person) SetPermissions(permissions []string) *Person {
	p.Permissions = make([]string, len(permissions))
	copy(p.Permissions, permissions)
	return p
}

// SetMetadata sets the optional metadata field
func (p *Person) SetMetadata(metadata []byte) *Person {
	p.Metadata = make([]byte, len(metadata))
	copy(p.Metadata, metadata)
	return p
}

// SetEmployeeType sets the optional employee type field
func (p *Person) SetEmployeeType(empType int) *Person {
	p.EmployeeType = &empType
	return p
}

// Tag returns the ASN.1 tag for Person (SEQUENCE)
func (p *Person) Tag() Tag {
	return NewUniversalTag(TagSequence, true)
}

// Encode encodes the Person as an ASN.1 SEQUENCE
func (p *Person) Encode() ([]byte, error) {
	seq := NewSequence()

	// Add required fields
	seq.Add(NewInteger(p.ID))
	seq.Add(NewUTF8String(p.Name))
	seq.Add(NewIA5String(p.Email))
	seq.Add(NewBoolean(p.IsActive))

	// Add optional fields with context-specific tags
	if p.Department != nil {
		dept := NewPrintableString(*p.Department)
		contextTag := NewContextSpecificTag(0, false)
		contextSpecific := NewStructured(contextTag)
		contextSpecific.Add(dept)
		seq.Add(contextSpecific)
	}

	if p.PhoneNumber != nil {
		phone := NewPrintableString(*p.PhoneNumber)
		contextTag := NewContextSpecificTag(1, false)
		contextSpecific := NewStructured(contextTag)
		contextSpecific.Add(phone)
		seq.Add(contextSpecific)
	}

	if p.Birthday != nil {
		// Convert time to UTCTime format (YYMMDDHHMMSSZ)
		utcTime := p.Birthday.UTC()
		timeStr := utcTime.Format("060102150405Z")
		utcTimeObj := NewIA5String(timeStr) // Simplified, normally would be UTCTime type
		contextTag := NewContextSpecificTag(2, false)
		contextSpecific := NewStructured(contextTag)
		contextSpecific.Add(utcTimeObj)
		seq.Add(contextSpecific)
	}

	if p.Salary != nil {
		salary := NewInteger(*p.Salary)
		contextTag := NewContextSpecificTag(3, false)
		contextSpecific := NewStructured(contextTag)
		contextSpecific.Add(salary)
		seq.Add(contextSpecific)
	}

	if p.Manager != nil {
		contextTag := NewContextSpecificTag(4, true)
		contextSpecific := NewStructured(contextTag)
		contextSpecific.Add(p.Manager)
		seq.Add(contextSpecific)
	}

	if len(p.Permissions) > 0 {
		permSeq := NewSequence()
		for _, perm := range p.Permissions {
			permSeq.Add(NewPrintableString(perm))
		}
		contextTag := NewContextSpecificTag(5, true)
		contextSpecific := NewStructured(contextTag)
		contextSpecific.Add(permSeq)
		seq.Add(contextSpecific)
	}

	if len(p.Metadata) > 0 {
		metadata := NewOctetString(p.Metadata)
		contextTag := NewContextSpecificTag(6, false)
		contextSpecific := NewStructured(contextTag)
		contextSpecific.Add(metadata)
		seq.Add(contextSpecific)
	}

	if p.EmployeeType != nil {
		empType := NewInteger(int64(*p.EmployeeType))
		contextTag := NewContextSpecificTag(7, false)
		contextSpecific := NewStructured(contextTag)
		contextSpecific.Add(empType)
		seq.Add(contextSpecific)
	}

	return seq.Encode()
}

// String returns a string representation of the Person
func (p *Person) String() string {
	result := fmt.Sprintf("Person {\n")
	result += fmt.Sprintf("  ID: %d\n", p.ID)
	result += fmt.Sprintf("  Name: %q\n", p.Name)
	result += fmt.Sprintf("  Email: %q\n", p.Email)
	result += fmt.Sprintf("  IsActive: %t\n", p.IsActive)

	if p.Department != nil {
		result += fmt.Sprintf("  Department: %q\n", *p.Department)
	}
	if p.PhoneNumber != nil {
		result += fmt.Sprintf("  PhoneNumber: %q\n", *p.PhoneNumber)
	}
	if p.Birthday != nil {
		result += fmt.Sprintf("  Birthday: %s\n", p.Birthday.Format("2006-01-02"))
	}
	if p.Salary != nil {
		result += fmt.Sprintf("  Salary: %d\n", *p.Salary)
	}
	if p.Manager != nil {
		result += fmt.Sprintf("  Manager: %q (ID: %d)\n", p.Manager.Name, p.Manager.ID)
	}
	if len(p.Permissions) > 0 {
		result += fmt.Sprintf("  Permissions: %v\n", p.Permissions)
	}
	if len(p.Metadata) > 0 {
		result += fmt.Sprintf("  Metadata: %d bytes\n", len(p.Metadata))
	}
	if p.EmployeeType != nil {
		empTypeStr := "unknown"
		switch *p.EmployeeType {
		case 0:
			empTypeStr = "full-time"
		case 1:
			empTypeStr = "part-time"
		case 2:
			empTypeStr = "contractor"
		}
		result += fmt.Sprintf("  EmployeeType: %s (%d)\n", empTypeStr, *p.EmployeeType)
	}

	result += "}"
	return result
}

// CompactString returns a compact string representation
func (p *Person) CompactString() string {
	return fmt.Sprintf("Person{ID: %d, Name: %q, Email: %q, Active: %t}", 
		p.ID, p.Name, p.Email, p.IsActive)
}

// PersonDirectory represents a collection of persons (demonstrates SEQUENCE OF)
type PersonDirectory struct {
	persons []*Person
}

// NewPersonDirectory creates a new PersonDirectory
func NewPersonDirectory() *PersonDirectory {
	return &PersonDirectory{
		persons: make([]*Person, 0),
	}
}

// AddPerson adds a person to the directory
func (pd *PersonDirectory) AddPerson(person *Person) {
	pd.persons = append(pd.persons, person)
}

// Persons returns all persons in the directory
func (pd *PersonDirectory) Persons() []*Person {
	result := make([]*Person, len(pd.persons))
	copy(result, pd.persons)
	return result
}

// Tag returns the ASN.1 tag for PersonDirectory (SEQUENCE)
func (pd *PersonDirectory) Tag() Tag {
	return NewUniversalTag(TagSequence, true)
}

// Encode encodes the PersonDirectory as an ASN.1 SEQUENCE OF Person
func (pd *PersonDirectory) Encode() ([]byte, error) {
	seq := NewSequence()
	for _, person := range pd.persons {
		seq.Add(person)
	}
	return seq.Encode()
}

// String returns a string representation of the PersonDirectory
func (pd *PersonDirectory) String() string {
	return fmt.Sprintf("PersonDirectory (%d persons)", len(pd.persons))
}

// CompactString returns a compact representation showing all persons
func (pd *PersonDirectory) CompactString() string {
	result := fmt.Sprintf("PersonDirectory {\n")
	for i, person := range pd.persons {
		result += fmt.Sprintf("  [%d] %s\n", i, person.CompactString())
	}
	result += "}"
	return result
}