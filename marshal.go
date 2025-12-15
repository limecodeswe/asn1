package asn1

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// MarshalOptions contains options for marshaling
type MarshalOptions struct {
	// UseContextTags controls whether to use context-specific tags for optional fields
	UseContextTags bool
}

// DefaultMarshalOptions returns default marshaling options
func DefaultMarshalOptions() *MarshalOptions {
	return &MarshalOptions{
		UseContextTags: true,
	}
}

// Marshal encodes a Go struct to ASN.1 using struct tags
func Marshal(v interface{}) ([]byte, error) {
	return MarshalWithOptions(v, DefaultMarshalOptions())
}

// MarshalWithOptions encodes a Go struct to ASN.1 using struct tags with custom options
func MarshalWithOptions(v interface{}, opts *MarshalOptions) ([]byte, error) {
	obj, err := marshalValue(reflect.ValueOf(v), opts)
	if err != nil {
		return nil, err
	}
	return obj.Encode()
}

// Unmarshal decodes ASN.1 data into a Go struct using struct tags
func Unmarshal(data []byte, v interface{}) error {
	return UnmarshalWithOptions(data, v, DefaultMarshalOptions())
}

// UnmarshalWithOptions decodes ASN.1 data into a Go struct using struct tags with custom options
func UnmarshalWithOptions(data []byte, v interface{}, opts *MarshalOptions) error {
	// Decode the ASN.1 data first
	asn1Value, _, err := DecodeTLV(data)
	if err != nil {
		return fmt.Errorf("failed to decode ASN.1 data: %w", err)
	}

	// Convert ASN1Value to higher-level object if it's a structured type
	var obj ASN1Object = asn1Value
	if asn1Value.Tag().Constructed {
		// It's a structured type, create an ASN1Structured
		structured := NewStructured(asn1Value.Tag())

		// Parse the content to extract individual elements
		content := asn1Value.Value()
		offset := 0

		for offset < len(content) {
			elementValue, consumed, err := DecodeTLV(content[offset:])
			if err != nil {
				return fmt.Errorf("failed to decode element: %w", err)
			}

			// Convert element to higher-level object
			element := convertToHighLevelObject(elementValue)
			structured.Add(element)
			offset += consumed
		}

		obj = structured
	} else {
		// It's a primitive type, convert to specific object
		obj = convertToHighLevelObject(asn1Value)
	}

	return unmarshalValue(obj, reflect.ValueOf(v).Elem(), opts)
}

// convertToHighLevelObject converts an ASN1Value to its appropriate higher-level object
func convertToHighLevelObject(val *ASN1Value) ASN1Object {
	tag := val.Tag()

	if tag.Constructed {
		// It's a structured type
		structured := NewStructured(tag)

		// Parse the content to extract individual elements
		content := val.Value()
		offset := 0

		for offset < len(content) {
			elementValue, consumed, err := DecodeTLV(content[offset:])
			if err != nil {
				// If we can't parse the content, return as ASN1Value
				return val
			}

			// Recursively convert elements
			element := convertToHighLevelObject(elementValue)
			structured.Add(element)
			offset += consumed
		}

		return structured
	} else {
		// It's a primitive type, convert to specific object
		return convertPrimitiveValue(val)
	}
}

// convertPrimitiveValue converts an ASN1Value to its specific typed object
func convertPrimitiveValue(val *ASN1Value) ASN1Object {
	tag := val.Tag()
	value := val.Value()

	// Handle context-specific tags - don't try to decode them
	// With implicit tagging, the value is raw data, not a TLV structure
	// The unmarshalStruct function will handle tag restoration via restoreTag
	if tag.Class == 2 {
		return val
	}

	// Handle universal class tags
	if tag.Class != 0 {
		return val // Only handle universal tags here
	}

	switch tag.Number {
	case TagBoolean:
		if len(value) == 1 {
			return NewBoolean(value[0] != 0)
		}
		return val
	case TagInteger:
		integer, err := DecodeIntegerValue(value)
		if err != nil {
			return val
		}
		return NewIntegerFromBigInt(integer)
	case TagOctetString:
		return NewOctetString(value)
	case TagUTF8String:
		return NewUTF8String(string(value))
	case TagPrintableString:
		return NewPrintableString(string(value))
	case TagIA5String:
		return NewIA5String(string(value))
	case TagUTCTime:
		// Re-encode as proper TLV and use the decoder
		tlvData, err := EncodeTLV(tag, value)
		if err != nil {
			return val
		}
		utcTime, _, err := DecodeUTCTime(tlvData)
		if err != nil {
			return val
		}
		return utcTime
	case TagGeneralizedTime:
		// Re-encode as proper TLV and use the decoder
		tlvData, err := EncodeTLV(tag, value)
		if err != nil {
			return val
		}
		genTime, _, err := DecodeGeneralizedTime(tlvData)
		if err != nil {
			return val
		}
		return genTime
	default:
		return val
	}
} // fieldInfo represents parsed ASN.1 field information from struct tags
type fieldInfo struct {
	Name      string
	Type      string
	Optional  bool
	Tag       int
	HasTag    bool
	Omitempty bool
	Explicit  bool // If true, use explicit tagging (wrap); if false, use implicit tagging (replace)
}

// parseASN1Tag parses an ASN.1 struct tag
func parseASN1Tag(tag string) (*fieldInfo, error) {
	if tag == "" {
		return nil, fmt.Errorf("empty ASN.1 tag")
	}

	parts := strings.Split(tag, ",")
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid ASN.1 tag format")
	}

	info := &fieldInfo{
		Type: strings.ToLower(strings.TrimSpace(parts[0])),
	}

	// Parse options
	for i := 1; i < len(parts); i++ {
		part := strings.TrimSpace(parts[i])
		switch {
		case part == "optional":
			info.Optional = true
		case part == "omitempty":
			info.Omitempty = true
		case part == "explicit":
			info.Explicit = true
		case strings.HasPrefix(part, "tag:"):
			tagStr := strings.TrimPrefix(part, "tag:")
			tagNum, err := strconv.Atoi(tagStr)
			if err != nil {
				return nil, fmt.Errorf("invalid tag number: %s", tagStr)
			}
			info.Tag = tagNum
			info.HasTag = true
		}
	}

	return info, nil
}

// marshalValue converts a Go value to an ASN.1 object based on its type and tags
func marshalValue(v reflect.Value, opts *MarshalOptions) (ASN1Object, error) {
	// Handle pointer types
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil, fmt.Errorf("nil pointer cannot be marshaled")
		}
		return marshalValue(v.Elem(), opts)
	}

	switch v.Kind() {
	case reflect.Struct:
		return marshalStruct(v, opts)
	case reflect.Slice:
		if v.Type().Elem().Kind() == reflect.Uint8 {
			// []byte -> OCTET STRING
			return NewOctetString(v.Bytes()), nil
		}
		return marshalSlice(v, opts)
	case reflect.String:
		// Default to UTF8String, but this should be overridden by tags
		return NewUTF8String(v.String()), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return NewInteger(v.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return NewInteger(int64(v.Uint())), nil
	case reflect.Bool:
		return NewBoolean(v.Bool()), nil
	case reflect.Interface:
		// Handle interface{} for CHOICE types
		if v.IsNil() {
			return nil, fmt.Errorf("nil interface cannot be marshaled")
		}
		return marshalValue(v.Elem(), opts)
	default:
		return nil, fmt.Errorf("unsupported type: %v", v.Type())
	}
}

// marshalStruct converts a Go struct to an ASN.1 SEQUENCE
func marshalStruct(v reflect.Value, opts *MarshalOptions) (ASN1Object, error) {
	if v.Type() == reflect.TypeOf(time.Time{}) {
		// Special handling for time.Time
		t := v.Interface().(time.Time)
		return NewUTCTime(t), nil
	}

	t := v.Type()
	seq := NewSequence()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Skip unexported fields
		if !fieldType.IsExported() {
			continue
		}

		// Parse ASN.1 tag
		tag := fieldType.Tag.Get("asn1")
		if tag == "-" {
			continue // Skip this field
		}

		var info *fieldInfo
		var err error

		if tag != "" {
			info, err = parseASN1Tag(tag)
			if err != nil {
				return nil, fmt.Errorf("field %s: %w", fieldType.Name, err)
			}
		} else {
			// Default behavior without tags
			info = &fieldInfo{Type: "auto"}
		}

		// Handle optional fields (pointers)
		if field.Kind() == reflect.Ptr && field.IsNil() {
			if info.Optional {
				continue // Skip nil optional fields
			}
			return nil, fmt.Errorf("required field %s is nil", fieldType.Name)
		}

		// Marshal the field value
		var obj ASN1Object
		if info.Type == "auto" {
			obj, err = marshalValue(field, opts)
		} else {
			obj, err = marshalTypedValue(field, info, opts)
		}
		if err != nil {
			return nil, fmt.Errorf("field %s: %w", fieldType.Name, err)
		}

		// Apply context-specific tag if specified
		if info.HasTag && opts.UseContextTags {
			if info.Explicit {
				// EXPLICIT tagging: wrap the object with context-specific tag
				// The wrapper is always constructed for EXPLICIT tagging
				contextTag := NewContextSpecificTag(info.Tag, true)
				wrapped := NewStructured(contextTag)
				wrapped.Add(obj)
				obj = wrapped
			} else {
				// IMPLICIT tagging: replace the object's tag with context-specific tag
				obj = replaceTag(obj, info.Tag)
			}
		}

		seq.Add(obj)
	}

	return seq, nil
}

// marshalSlice converts a Go slice to an ASN.1 SEQUENCE OF
func marshalSlice(v reflect.Value, opts *MarshalOptions) (ASN1Object, error) {
	seq := NewSequence()

	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)
		obj, err := marshalValue(elem, opts)
		if err != nil {
			return nil, fmt.Errorf("slice element %d: %w", i, err)
		}
		seq.Add(obj)
	}

	return seq, nil
}

// tryCustomMarshal attempts to use a custom marshaler if the value implements ASN1Marshaler.
// Returns (result, nil) if successful, (nil, error) if custom marshaler exists but fails,
// or (nil, nil) if no custom marshaler is available.
func tryCustomMarshal(v reflect.Value, info *fieldInfo) (ASN1Object, error) {
	// Try both the value and its pointer receiver
	if v.CanInterface() {
		if m, ok := v.Interface().(ASN1Marshaler); ok {
			rawBytes, err := m.MarshalASN1()
			if err != nil {
				return nil, fmt.Errorf("custom marshaler failed: %w", err)
			}
			return wrapCustomMarshaledBytes(rawBytes, info)
		}
	}
	
	// Also try with pointer receiver
	if v.CanAddr() && v.Addr().CanInterface() {
		if m, ok := v.Addr().Interface().(ASN1Marshaler); ok {
			rawBytes, err := m.MarshalASN1()
			if err != nil {
				return nil, fmt.Errorf("custom marshaler failed: %w", err)
			}
			return wrapCustomMarshaledBytes(rawBytes, info)
		}
	}
	
	return nil, nil
}

// marshalTypedValue converts a Go value to a specific ASN.1 type based on the field info
func marshalTypedValue(v reflect.Value, info *fieldInfo, opts *MarshalOptions) (ASN1Object, error) {
	// Check if the value implements custom marshaler interface
	obj, err := tryCustomMarshal(v, info)
	if err != nil {
		return nil, err
	}
	if obj != nil {
		return obj, nil
	}

	// Handle pointer types
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil, fmt.Errorf("cannot marshal nil pointer")
		}
		return marshalTypedValue(v.Elem(), info, opts)
	}

	switch info.Type {
	case "boolean":
		if v.Kind() != reflect.Bool {
			return nil, fmt.Errorf("expected bool for boolean type, got %v", v.Type())
		}
		return NewBoolean(v.Bool()), nil

	case "integer":
		switch v.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return NewInteger(v.Int()), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return NewInteger(int64(v.Uint())), nil
		default:
			return nil, fmt.Errorf("expected integer type for integer, got %v", v.Type())
		}

	case "octetstring":
		if v.Kind() == reflect.Slice && v.Type().Elem().Kind() == reflect.Uint8 {
			return NewOctetString(v.Bytes()), nil
		}
		return nil, fmt.Errorf("expected []byte for octetstring, got %v", v.Type())

	case "utf8string":
		if v.Kind() != reflect.String {
			return nil, fmt.Errorf("expected string for utf8string, got %v", v.Type())
		}
		return NewUTF8String(v.String()), nil

	case "printablestring":
		if v.Kind() != reflect.String {
			return nil, fmt.Errorf("expected string for printablestring, got %v", v.Type())
		}
		return NewPrintableString(v.String()), nil

	case "ia5string":
		if v.Kind() != reflect.String {
			return nil, fmt.Errorf("expected string for ia5string, got %v", v.Type())
		}
		return NewIA5String(v.String()), nil

	case "utctime":
		if v.Type() == reflect.TypeOf(time.Time{}) {
			return NewUTCTime(v.Interface().(time.Time)), nil
		}
		return nil, fmt.Errorf("expected time.Time for utctime, got %v", v.Type())

	case "generalizedtime":
		if v.Type() == reflect.TypeOf(time.Time{}) {
			return NewGeneralizedTime(v.Interface().(time.Time)), nil
		}
		return nil, fmt.Errorf("expected time.Time for generalizedtime, got %v", v.Type())

	case "sequence":
		if v.Kind() == reflect.Struct {
			return marshalStruct(v, opts)
		} else if v.Kind() == reflect.Slice {
			return marshalSlice(v, opts)
		}
		return nil, fmt.Errorf("expected struct or slice for sequence, got %v", v.Type())

	case "choice":
		// Handle CHOICE types - the field should be interface{} or a choice struct
		if v.Kind() == reflect.Interface {
			if v.IsNil() {
				return nil, fmt.Errorf("choice field is nil")
			}
			return marshalValue(v.Elem(), opts)
		} else if v.Kind() == reflect.Struct {
			return marshalChoiceStruct(v, opts)
		}
		return nil, fmt.Errorf("expected interface{} or struct for choice, got %v", v.Type())

	default:
		return nil, fmt.Errorf("unsupported ASN.1 type: %s", info.Type)
	}
}

// wrapCustomMarshaledBytes wraps custom marshaled bytes with the appropriate ASN.1 tag
func wrapCustomMarshaledBytes(rawBytes []byte, info *fieldInfo) (ASN1Object, error) {
	// Create an ASN.1 object based on the field's type tag
	switch info.Type {
	case "octetstring":
		return NewOctetString(rawBytes), nil
	case "integer":
		// For integer, we need to decode the bytes as an integer value
		intVal, err := DecodeIntegerValue(rawBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to decode custom marshaled integer: %w", err)
		}
		return NewIntegerFromBigInt(intVal), nil
	case "utf8string":
		return NewUTF8String(string(rawBytes)), nil
	case "printablestring":
		return NewPrintableString(string(rawBytes)), nil
	case "ia5string":
		return NewIA5String(string(rawBytes)), nil
	case "sequence":
		// For sequence, the custom marshaler should return properly encoded sequence content
		// We need to decode it as a TLV and convert to ASN1Structured
		asn1Value, _, err := DecodeTLV(rawBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to decode custom marshaled sequence: %w", err)
		}
		return convertToHighLevelObject(asn1Value), nil
	case "boolean":
		if len(rawBytes) != 1 {
			return nil, fmt.Errorf("invalid boolean value from custom marshaler")
		}
		return NewBoolean(rawBytes[0] != 0), nil
	case "bitstring":
		// For bit string, assume the raw bytes are the bit string content with no unused bits
		return NewBitString(rawBytes, 0), nil
	default:
		// For unknown types or generic cases, wrap as octet string
		return NewOctetString(rawBytes), nil
	}
}

// extractRawBytes extracts the raw value bytes from an ASN.1 object for custom unmarshaling
func extractRawBytes(obj ASN1Object) ([]byte, error) {
	// For different ASN.1 object types, extract their value bytes
	switch o := obj.(type) {
	case *ASN1OctetString:
		return o.Value(), nil
	case *ASN1Integer:
		// For integers, we need the raw encoded value bytes (not the TLV structure)
		// We encode then decode to extract just the value portion
		encoded, err := o.Encode()
		if err != nil {
			return nil, err
		}
		// Decode TLV to get just the value part (skips tag and length)
		val, _, err := DecodeTLV(encoded)
		if err != nil {
			return nil, err
		}
		return val.Value(), nil
	case *ASN1UTF8String:
		return []byte(o.Value()), nil
	case *ASN1PrintableString:
		return []byte(o.Value()), nil
	case *ASN1IA5String:
		return []byte(o.Value()), nil
	case *ASN1Boolean:
		// ASN.1 BOOLEAN encoding: 0x00 = false, 0xFF (or any non-zero) = true
		if o.Value() {
			return []byte{0xFF}, nil
		}
		return []byte{0x00}, nil
	case *ASN1BitString:
		return o.Value(), nil
	case *ASN1Structured:
		// For structured types, encode and return just the content
		encoded, err := o.Encode()
		if err != nil {
			return nil, err
		}
		// Decode TLV to get just the value part
		val, _, err := DecodeTLV(encoded)
		if err != nil {
			return nil, err
		}
		return val.Value(), nil
	case *ASN1Value:
		return o.Value(), nil
	default:
		return nil, fmt.Errorf("unsupported ASN1Object type for custom unmarshaling: %T", obj)
	}
}

// unmarshalValue converts an ASN.1 object to a Go value
func unmarshalValue(obj ASN1Object, v reflect.Value, opts *MarshalOptions) error {
	// Check if the value implements custom unmarshaler interface
	// We need to check with pointer receiver since UnmarshalASN1 typically modifies the value
	if v.CanAddr() && v.Addr().CanInterface() {
		if u, ok := v.Addr().Interface().(ASN1Unmarshaler); ok {
			// Get the raw encoded bytes from the ASN.1 object
			rawBytes, err := extractRawBytes(obj)
			if err != nil {
				return fmt.Errorf("failed to extract raw bytes for custom unmarshaler: %w", err)
			}
			return u.UnmarshalASN1(rawBytes)
		}
	}

	// Handle pointer types
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			// Create new instance
			v.Set(reflect.New(v.Type().Elem()))
		}
		return unmarshalValue(obj, v.Elem(), opts)
	}

	switch v.Kind() {
	case reflect.Struct:
		return unmarshalStruct(obj, v, opts)
	case reflect.Slice:
		return unmarshalSlice(obj, v, opts)
	case reflect.String:
		return unmarshalString(obj, v)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return unmarshalInt(obj, v)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return unmarshalUint(obj, v)
	case reflect.Bool:
		return unmarshalBool(obj, v)
	case reflect.Interface:
		// Handle interface{} for choice types
		return unmarshalInterface(obj, v)
	default:
		return fmt.Errorf("unsupported type for unmarshaling: %v", v.Type())
	}
}

// unmarshalStruct converts an ASN.1 SEQUENCE to a Go struct
func unmarshalStruct(obj ASN1Object, v reflect.Value, opts *MarshalOptions) error {
	if v.Type() == reflect.TypeOf(time.Time{}) {
		// Special handling for time.Time
		return unmarshalTime(obj, v)
	}

	structured, ok := obj.(*ASN1Structured)
	if !ok {
		return fmt.Errorf("expected ASN1Structured for struct, got %T", obj)
	}

	elements := structured.Elements()
	t := v.Type()
	elementIndex := 0

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Skip unexported fields
		if !fieldType.IsExported() {
			continue
		}

		// Parse ASN.1 tag
		tag := fieldType.Tag.Get("asn1")
		if tag == "-" {
			continue // Skip this field
		}

		var info *fieldInfo
		var err error

		if tag != "" {
			info, err = parseASN1Tag(tag)
			if err != nil {
				return fmt.Errorf("field %s: %w", fieldType.Name, err)
			}
		} else {
			info = &fieldInfo{Type: "auto"}
		}

		// Check if we have more elements
		if elementIndex >= len(elements) {
			if info.Optional {
				continue // Skip optional fields if no more elements
			}
			return fmt.Errorf("not enough elements for required field %s", fieldType.Name)
		}

		element := elements[elementIndex]

		// Handle context-specific tags
		if info.HasTag && opts.UseContextTags {
			// Check if the tag matches before consuming the element
			if element.Tag().Class == 2 && element.Tag().Number == info.Tag {
				// Tag matches, consume the element
				elementIndex++
				
				if info.Explicit {
					// EXPLICIT tagging: unwrap to get the inner element
					if wrapped, ok := element.(*ASN1Structured); ok {
						wrappedElements := wrapped.Elements()
						if len(wrappedElements) == 1 {
							element = wrappedElements[0]
						}
					}
				} else {
					// IMPLICIT tagging: restore the original tag
					element = restoreTag(element, info.Type)
				}
			} else {
				// Tag doesn't match
				if info.Optional {
					// Optional field not present, skip without consuming element
					continue
				}
				// Required field with wrong tag - this is an error
				return fmt.Errorf("field %s: expected tag [CONTEXT %d], got %s", 
					fieldType.Name, info.Tag, element.Tag().TagString())
			}
		} else {
			// No specific tag expected, consume the element
			elementIndex++
		}

		// Unmarshal the element
		if err := unmarshalValue(element, field, opts); err != nil {
			return fmt.Errorf("field %s: %w", fieldType.Name, err)
		}
	}

	return nil
}

// unmarshalSlice converts an ASN.1 SEQUENCE OF to a Go slice
func unmarshalSlice(obj ASN1Object, v reflect.Value, opts *MarshalOptions) error {
	if v.Type().Elem().Kind() == reflect.Uint8 {
		// Handle []byte special case
		if octets, ok := obj.(*ASN1OctetString); ok {
			v.SetBytes(octets.Value())
			return nil
		}
		return fmt.Errorf("expected ASN1OctetString for []byte, got %T", obj)
	}

	structured, ok := obj.(*ASN1Structured)
	if !ok {
		return fmt.Errorf("expected ASN1Structured for slice, got %T", obj)
	}

	elements := structured.Elements()
	slice := reflect.MakeSlice(v.Type(), len(elements), len(elements))

	for i, element := range elements {
		if err := unmarshalValue(element, slice.Index(i), opts); err != nil {
			return fmt.Errorf("slice element %d: %w", i, err)
		}
	}

	v.Set(slice)
	return nil
}

// marshalChoiceStruct handles structs marked as CHOICE types
func marshalChoiceStruct(v reflect.Value, opts *MarshalOptions) (ASN1Object, error) {
	t := v.Type()

	// Look for exactly one non-nil pointer field
	var chosenField reflect.Value
	var chosenInfo *fieldInfo
	var chosenFieldName string
	var fieldCount int

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Skip unexported fields
		if !fieldType.IsExported() {
			continue
		}

		// Parse ASN.1 tag
		tag := fieldType.Tag.Get("asn1")
		if tag == "-" {
			continue // Skip this field
		}

		// For choice structs, all fields should be pointers
		if field.Kind() != reflect.Ptr {
			continue
		}

		if !field.IsNil() {
			if fieldCount > 0 {
				return nil, fmt.Errorf("choice struct has multiple non-nil fields: %s and %s", chosenFieldName, fieldType.Name)
			}

			chosenField = field
			chosenFieldName = fieldType.Name
			fieldCount++

			if tag != "" {
				var err error
				chosenInfo, err = parseASN1Tag(tag)
				if err != nil {
					return nil, fmt.Errorf("field %s: %w", fieldType.Name, err)
				}
			} else {
				chosenInfo = &fieldInfo{Type: "auto"}
			}
		}
	}

	if fieldCount == 0 {
		return nil, fmt.Errorf("choice struct has no non-nil fields")
	}

	// Marshal the chosen field
	var obj ASN1Object
	var err error
	if chosenInfo.Type == "auto" {
		obj, err = marshalValue(chosenField, opts)
	} else {
		obj, err = marshalTypedValue(chosenField, chosenInfo, opts)
	}
	if err != nil {
		return nil, fmt.Errorf("field %s: %w", chosenFieldName, err)
	}

	// Apply context-specific tag if specified
	if chosenInfo.HasTag && opts.UseContextTags {
		contextTag := NewContextSpecificTag(chosenInfo.Tag, isConstructedType(obj))
		wrapped := NewStructured(contextTag)
		wrapped.Add(obj)
		obj = wrapped
	}

	return obj, nil
}

// Helper functions for unmarshaling basic types
func unmarshalString(obj ASN1Object, v reflect.Value) error {
	switch s := obj.(type) {
	case *ASN1UTF8String:
		v.SetString(s.Value())
	case *ASN1PrintableString:
		v.SetString(s.Value())
	case *ASN1IA5String:
		v.SetString(s.Value())
	default:
		return fmt.Errorf("expected string type, got %T", obj)
	}
	return nil
}

func unmarshalInt(obj ASN1Object, v reflect.Value) error {
	integer, ok := obj.(*ASN1Integer)
	if !ok {
		return fmt.Errorf("expected ASN1Integer, got %T", obj)
	}
	val := integer.Value()
	if !val.IsInt64() {
		return fmt.Errorf("integer value too large for int64")
	}
	v.SetInt(val.Int64())
	return nil
}

func unmarshalUint(obj ASN1Object, v reflect.Value) error {
	integer, ok := obj.(*ASN1Integer)
	if !ok {
		return fmt.Errorf("expected ASN1Integer, got %T", obj)
	}
	val := integer.Value()
	if val.Sign() < 0 {
		return fmt.Errorf("cannot convert negative integer to unsigned")
	}
	if !val.IsUint64() {
		return fmt.Errorf("integer value too large for uint64")
	}
	v.SetUint(val.Uint64())
	return nil
}

func unmarshalBool(obj ASN1Object, v reflect.Value) error {
	boolean, ok := obj.(*ASN1Boolean)
	if !ok {
		return fmt.Errorf("expected ASN1Boolean, got %T", obj)
	}
	v.SetBool(boolean.Value())
	return nil
}

func unmarshalTime(obj ASN1Object, v reflect.Value) error {
	switch t := obj.(type) {
	case *ASN1UTCTime:
		v.Set(reflect.ValueOf(t.Time()))
	case *ASN1GeneralizedTime:
		v.Set(reflect.ValueOf(t.Time()))
	default:
		return fmt.Errorf("expected time type, got %T", obj)
	}
	return nil
}

// unmarshalInterface converts an ASN.1 object to an interface{} value
func unmarshalInterface(obj ASN1Object, v reflect.Value) error {
	// Convert ASN.1 object to appropriate Go type
	switch o := obj.(type) {
	case *ASN1Boolean:
		v.Set(reflect.ValueOf(o.Value()))
	case *ASN1Integer:
		val := o.Value()
		if val.IsInt64() {
			v.Set(reflect.ValueOf(val.Int64()))
		} else {
			v.Set(reflect.ValueOf(val)) // Keep as *big.Int for large values
		}
	case *ASN1UTF8String:
		v.Set(reflect.ValueOf(o.Value()))
	case *ASN1PrintableString:
		v.Set(reflect.ValueOf(o.Value()))
	case *ASN1IA5String:
		v.Set(reflect.ValueOf(o.Value()))
	case *ASN1OctetString:
		v.Set(reflect.ValueOf(o.Value()))
	case *ASN1UTCTime:
		v.Set(reflect.ValueOf(o.Time()))
	case *ASN1GeneralizedTime:
		v.Set(reflect.ValueOf(o.Time()))
	default:
		return fmt.Errorf("cannot unmarshal %T to interface{}", obj)
	}
	return nil
}

// isConstructedType determines if an ASN.1 object should use constructed encoding
func isConstructedType(obj ASN1Object) bool {
	switch obj.(type) {
	case *ASN1Structured:
		return true
	default:
		return false
	}
}

// replaceTag creates a new ASN.1 object with a context-specific tag replacing the original tag
// This implements IMPLICIT tagging
func replaceTag(obj ASN1Object, tagNum int) ASN1Object {
	// Get the raw encoded value
	encoded, err := obj.Encode()
	if err != nil {
		// If we can't encode, return original object
		return obj
	}

	// Decode to get the original TLV structure
	origValue, _, err := DecodeTLV(encoded)
	if err != nil {
		// If we can't decode, return original object
		return obj
	}

	// Create new tag with context-specific class, preserving constructed bit
	newTag := Tag{
		Class:       2, // Context-specific
		Constructed: origValue.Tag().Constructed,
		Number:      tagNum,
	}

	// Create new ASN1Value with the new tag but same content
	newValue := NewASN1Value(newTag, origValue.Value())

	// If it was constructed, return as ASN1Structured
	if newTag.Constructed {
		structured := NewStructured(newTag)
		// Parse the content to extract individual elements
		content := newValue.Value()
		offset := 0

		for offset < len(content) {
			elementValue, consumed, err := DecodeTLV(content[offset:])
			if err != nil {
				break
			}
			element := convertToHighLevelObject(elementValue)
			structured.Add(element)
			offset += consumed
		}
		return structured
	}

	// For primitive types, return the ASN1Value as-is
	// Don't try to convert it - context-specific tags should be preserved during marshaling
	return newValue
}

// restoreTag restores the original universal tag from an implicitly tagged object
// This is used during unmarshaling to convert context-specific tags back to universal tags
func restoreTag(obj ASN1Object, asn1Type string) ASN1Object {
	// Get the raw encoded value
	encoded, err := obj.Encode()
	if err != nil {
		return obj
	}

	// Decode to get the current TLV structure
	currentValue, _, err := DecodeTLV(encoded)
	if err != nil {
		return obj
	}

	// Map ASN.1 type name to universal tag number
	var tagNum int
	var constructed bool

	switch strings.ToLower(asn1Type) {
	case "boolean":
		tagNum = TagBoolean
	case "integer":
		tagNum = TagInteger
	case "octetstring":
		tagNum = TagOctetString
	case "utf8string":
		tagNum = TagUTF8String
	case "printablestring":
		tagNum = TagPrintableString
	case "ia5string":
		tagNum = TagIA5String
	case "utctime":
		tagNum = TagUTCTime
	case "generalizedtime":
		tagNum = TagGeneralizedTime
	case "sequence":
		tagNum = TagSequence
		constructed = true
	case "set":
		tagNum = TagSet
		constructed = true
	case "choice":
		// For CHOICE types, we need to restore to a SEQUENCE tag
		// since the choice struct is represented as a SEQUENCE with one alternative
		tagNum = TagSequence
		constructed = true
	default:
		// Unknown type, return as-is
		return obj
	}

	// Create new tag with universal class
	newTag := Tag{
		Class:       0, // Universal
		Constructed: constructed,
		Number:      tagNum,
	}

	// Create new ASN1Value with the restored tag and same content
	newValue := NewASN1Value(newTag, currentValue.Value())

	// Convert to appropriate high-level object
	if constructed {
		structured := NewStructured(newTag)
		content := newValue.Value()
		offset := 0

		for offset < len(content) {
			elementValue, consumed, err := DecodeTLV(content[offset:])
			if err != nil {
				break
			}
			element := convertToHighLevelObject(elementValue)
			structured.Add(element)
			offset += consumed
		}
		return structured
	}

	return convertPrimitiveValue(newValue)
}
