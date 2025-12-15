package asn1

import (
	"encoding/hex"
	"fmt"
	"testing"
)

// This file contains the exact reproduction from the issue to verify the fix

// Custom type with marshaler (from the issue)
type IssuePhoneNumber struct {
	Digits string
}

func (p IssuePhoneNumber) MarshalASN1() ([]byte, error) {
	// Custom encoding: prefix with 0xAA
	return append([]byte{0xAA}, []byte(p.Digits)...), nil
}

func (p *IssuePhoneNumber) UnmarshalASN1(data []byte) error {
	if len(data) < 1 || data[0] != 0xAA {
		return fmt.Errorf("invalid prefix")
	}
	p.Digits = string(data[1:])
	return nil
}

type IssueMessage struct {
	Numbers []IssuePhoneNumber `asn1:"sequence,tag:0"`
}

// TestIssueReproduction reproduces the exact issue from the bug report
func TestIssueReproduction(t *testing.T) {
	msg := &IssueMessage{
		Numbers: []IssuePhoneNumber{
			{Digits: "123"},
		},
	}
	
	encoded, err := Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	
	t.Logf("Encoded: %s", hex.EncodeToString(encoded))
	
	// Verify custom marshaler was called (should have 0xAA prefix)
	if !contains(encoded, []byte{0xAA, '1', '2', '3'}) {
		t.Errorf("Custom marshaler not invoked: expected 0xAA prefix in %s", hex.EncodeToString(encoded))
	}
	
	// This should now work (it failed before the fix)
	decoded := &IssueMessage{}
	err = Unmarshal(encoded, decoded)
	if err != nil {
		t.Fatalf("Decoding failed (this was the bug): %v", err)
	}
	
	// Verify round-trip
	if len(decoded.Numbers) != 1 {
		t.Errorf("Expected 1 number, got %d", len(decoded.Numbers))
	}
	if decoded.Numbers[0].Digits != "123" {
		t.Errorf("Number mismatch: got %q, want %q", decoded.Numbers[0].Digits, "123")
	}
	
	t.Log("✓ Issue reproduction test passed - custom marshalers work for slice elements!")
}

// TestRealWorldCAP reproduces the real-world CAP protocol issue
type IssueISDNAddress struct {
	Nature        byte
	NumberingPlan byte
	Digits        string
}

func (a *IssueISDNAddress) MarshalASN1() ([]byte, error) {
	// Encode first byte: nature (bits 4-6) | numbering plan (bits 0-3) | extension bit
	firstByte := (a.Nature << 4) | a.NumberingPlan | 0x80
	
	// Encode digits as TBCD (Binary Coded Decimal with nibble swap)
	tbcdDigits := make([]byte, (len(a.Digits)+1)/2)
	for i := 0; i < len(a.Digits); i++ {
		digit := a.Digits[i] - '0'
		byteIdx := i / 2
		if i%2 == 0 {
			tbcdDigits[byteIdx] = digit
		} else {
			tbcdDigits[byteIdx] |= digit << 4
		}
	}
	if len(a.Digits)%2 == 1 {
		tbcdDigits[len(tbcdDigits)-1] |= 0xF0
	}
	
	return append([]byte{firstByte}, tbcdDigits...), nil
}

func (a *IssueISDNAddress) UnmarshalASN1(data []byte) error {
	if len(data) < 1 {
		return fmt.Errorf("ISDN address too short")
	}
	
	a.Nature = (data[0] >> 4) & 0x07
	a.NumberingPlan = data[0] & 0x0F
	
	// Decode TBCD
	var digits []byte
	for _, b := range data[1:] {
		lowNibble := b & 0x0F
		if lowNibble <= 9 {
			digits = append(digits, '0'+lowNibble)
		} else if lowNibble == 0xF {
			break
		}
		
		highNibble := (b >> 4) & 0x0F
		if highNibble <= 9 {
			digits = append(digits, '0'+highNibble)
		} else if highNibble == 0xF {
			break
		}
	}
	
	a.Digits = string(digits)
	return nil
}

type IssueConnectArg struct {
	DestinationRoutingAddress []IssueISDNAddress `asn1:"sequence,tag:0"`
}

func TestRealWorldCAPProtocol(t *testing.T) {
	arg := &IssueConnectArg{
		DestinationRoutingAddress: []IssueISDNAddress{
			{
				Nature:        0x01,
				NumberingPlan: 0x01,
				Digits:        "467011111",
			},
		},
	}
	
	encoded, err := Marshal(arg)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	
	t.Logf("Encoded: %s", hex.EncodeToString(encoded))
	
	// Check that TBCD encoding is used, not ASCII
	// ASCII "467011111" would be: 34 36 37 30 31 31 31 31 31
	asciiBytes := []byte("467011111")
	if contains(encoded, asciiBytes) {
		t.Errorf("BUG: Found ASCII encoding %s instead of TBCD", hex.EncodeToString(asciiBytes))
		t.Errorf("Full encoding: %s", hex.EncodeToString(encoded))
		t.Errorf("This was the original bug - custom marshaler was not invoked!")
		t.FailNow()
	}
	
	// Verify TBCD encoding is present (first byte 0x91 = nature/plan, then TBCD digits)
	expectedFirstByte := byte(0x91) // (0x01 << 4) | 0x01 | 0x80
	if !contains(encoded, []byte{expectedFirstByte}) {
		t.Errorf("Expected TBCD first byte 0x%02X not found", expectedFirstByte)
	}
	
	// Decode
	decoded := &IssueConnectArg{}
	err = Unmarshal(encoded, decoded)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	
	// Verify
	if len(decoded.DestinationRoutingAddress) != 1 {
		t.Fatalf("Expected 1 address, got %d", len(decoded.DestinationRoutingAddress))
	}
	
	if decoded.DestinationRoutingAddress[0].Digits != "467011111" {
		t.Errorf("Digits mismatch: got %q, want %q", 
			decoded.DestinationRoutingAddress[0].Digits, "467011111")
	}
	
	t.Log("✓ Real-world CAP protocol test passed - TBCD encoding works!")
}

// Helper function to check if haystack contains needle
func contains(haystack, needle []byte) bool {
	if len(needle) == 0 {
		return true
	}
	if len(needle) > len(haystack) {
		return false
	}
	for i := 0; i <= len(haystack)-len(needle); i++ {
		match := true
		for j := 0; j < len(needle); j++ {
			if haystack[i+j] != needle[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
