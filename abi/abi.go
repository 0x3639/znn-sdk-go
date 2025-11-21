package abi

import (
	"bytes"
	"crypto/sha3"
	"encoding/json"
	"fmt"
	"strings"
)

// =============================================================================
// TypeEnum - Entry Type Enumeration
// =============================================================================

// TypeEnum represents the type of an ABI entry
type TypeEnum int

const (
	// Function represents a function entry
	Function TypeEnum = iota
)

func (te TypeEnum) String() string {
	switch te {
	case Function:
		return "function"
	default:
		return "unknown"
	}
}

// =============================================================================
// Param - Function Parameter
// =============================================================================

// Param represents a function parameter
type Param struct {
	Indexed bool
	Name    string
	Type    AbiType
}

// NewParam creates a new parameter
func NewParam(name string, typeName string) (*Param, error) {
	abiType, err := GetType(typeName)
	if err != nil {
		return nil, fmt.Errorf("failed to create param type: %w", err)
	}

	return &Param{
		Indexed: false,
		Name:    name,
		Type:    abiType,
	}, nil
}

// DecodeList decodes a list of encoded values according to parameter types
func DecodeList(params []Param, encoded []byte) ([]interface{}, error) {
	result := make([]interface{}, 0, len(params))

	offset := 0
	for _, param := range params {
		var decoded interface{}
		var err error

		if param.Type.IsDynamicType() {
			// For dynamic types, read the offset pointer
			offsetBig, decodeErr := DecodeInt(encoded, offset)
			if decodeErr != nil {
				return nil, fmt.Errorf("failed to decode offset for param %s: %w", param.Name, decodeErr)
			}
			dataOffset := int(offsetBig.Int64())

			// Decode from the pointed location
			decoded, err = param.Type.Decode(encoded, dataOffset)
			if err != nil {
				return nil, fmt.Errorf("failed to decode param %s: %w", param.Name, err)
			}
		} else {
			// For static types, decode directly
			decoded, err = param.Type.Decode(encoded, offset)
			if err != nil {
				return nil, fmt.Errorf("failed to decode param %s: %w", param.Name, err)
			}
		}

		result = append(result, decoded)
		offset += param.Type.GetFixedSize()
	}

	return result, nil
}

// =============================================================================
// Entry - ABI Entry (Function, Event, etc.)
// =============================================================================

// Entry represents an ABI entry (function, event, etc.)
type Entry struct {
	Name   string
	Inputs []Param
	Type   TypeEnum
}

// NewEntry creates a new ABI entry
func NewEntry(name string, inputs []Param, entryType TypeEnum) *Entry {
	return &Entry{
		Name:   name,
		Inputs: inputs,
		Type:   entryType,
	}
}

// FormatSignature formats the entry signature as "name(type1,type2,...)"
func (e *Entry) FormatSignature() string {
	paramTypes := make([]string, len(e.Inputs))
	for i, param := range e.Inputs {
		paramTypes[i] = param.Type.GetCanonicalName()
	}

	return e.Name + "(" + strings.Join(paramTypes, ",") + ")"
}

// FingerprintSignature returns the SHA3-256 hash of the signature
func (e *Entry) FingerprintSignature() []byte {
	signature := e.FormatSignature()
	hash := sha3.New256()
	// #nosec G104 -- hash.Write never returns an error
	hash.Write([]byte(signature)) //nolint:errcheck
	return hash.Sum(nil)
}

// EncodeSignature returns the full signature hash
func (e *Entry) EncodeSignature() []byte {
	return e.FingerprintSignature()
}

// EncodeArguments encodes function arguments with proper head/tail separation for dynamic types
func (e *Entry) EncodeArguments(args []interface{}) ([]byte, error) {
	if len(args) > len(e.Inputs) {
		return nil, fmt.Errorf("too many arguments: got %d, expected %d", len(args), len(e.Inputs))
	}

	// Calculate static size and count dynamic parameters
	staticSize := 0
	dynamicCount := 0
	for i := 0; i < len(args); i++ {
		paramType := e.Inputs[i].Type
		if paramType.IsDynamicType() {
			dynamicCount++
			// Dynamic types use 32 bytes in the head for offset pointer
			staticSize += 32
		} else {
			staticSize += paramType.GetFixedSize()
		}
	}

	// Create buffer for head (static) and tail (dynamic) parts
	// bb array holds: [static part for each arg] + [dynamic data for each dynamic arg]
	bb := make([][]byte, len(args)+dynamicCount)
	for i := range bb {
		bb[i] = []byte{}
	}

	// Encode each argument
	curDynamicPtr := staticSize
	curDynamicCount := 0
	for i := 0; i < len(args); i++ {
		paramType := e.Inputs[i].Type
		if paramType.IsDynamicType() {
			// For dynamic types: encode offset pointer in head, data in tail
			dynData, err := paramType.Encode(args[i])
			if err != nil {
				return nil, fmt.Errorf("failed to encode arg %d: %w", i, err)
			}

			bb[i] = EncodeInt(curDynamicPtr)
			bb[len(args)+curDynamicCount] = dynData
			curDynamicCount++
			curDynamicPtr += len(dynData)
		} else {
			// For static types: encode directly in head
			encoded, err := paramType.Encode(args[i])
			if err != nil {
				return nil, fmt.Errorf("failed to encode arg %d: %w", i, err)
			}
			bb[i] = encoded
		}
	}

	// Merge all parts
	return bytes.Join(bb, nil), nil
}

// =============================================================================
// AbiFunction - ABI Function Entry
// =============================================================================

const (
	// EncodedSignLength is the length of the encoded function signature (4 bytes)
	EncodedSignLength = 4
)

// AbiFunction represents an ABI function entry
type AbiFunction struct {
	Entry
}

// NewAbiFunction creates a new ABI function
func NewAbiFunction(name string, inputs []Param) *AbiFunction {
	return &AbiFunction{
		Entry: Entry{
			Name:   name,
			Inputs: inputs,
			Type:   Function,
		},
	}
}

// Decode decodes the encoded function call data (skipping the 4-byte signature)
func (af *AbiFunction) Decode(encoded []byte) ([]interface{}, error) {
	if len(encoded) < EncodedSignLength {
		return nil, fmt.Errorf("encoded data too short: %d bytes", len(encoded))
	}

	// Skip the first 4 bytes (function signature) and decode the rest
	return DecodeList(af.Inputs, encoded[EncodedSignLength:])
}

// Encode encodes the function call with signature and arguments
func (af *AbiFunction) Encode(args []interface{}) ([]byte, error) {
	// Encode arguments
	encodedArgs, err := af.EncodeArguments(args)
	if err != nil {
		return nil, err
	}

	// Combine signature (first 4 bytes) with encoded arguments
	signature := af.EncodeSignature()
	return append(signature, encodedArgs...), nil
}

// EncodeSignature returns the first 4 bytes of the signature hash
func (af *AbiFunction) EncodeSignature() []byte {
	fullSignature := af.Entry.EncodeSignature()
	return extractSignature(fullSignature)
}

// extractSignature extracts the first 4 bytes from the signature hash
func extractSignature(data []byte) []byte {
	if len(data) < EncodedSignLength {
		return data
	}
	return data[:EncodedSignLength]
}

// =============================================================================
// Abi - ABI Container
// =============================================================================

// Abi represents a collection of ABI entries (functions, events)
type Abi struct {
	Entries []Entry
}

// NewAbi creates a new ABI container from a list of entries
func NewAbi(entries []Entry) *Abi {
	return &Abi{
		Entries: entries,
	}
}

// parseEntries parses JSON ABI into entries
func parseEntries(jsonStr string) ([]Entry, error) {
	var rawEntries []map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &rawEntries); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	entries := make([]Entry, 0, len(rawEntries))
	for _, raw := range rawEntries {
		// Get entry name
		name, ok := raw["name"].(string)
		if !ok {
			return nil, fmt.Errorf("entry missing 'name' field")
		}

		// Check entry type (only functions supported for now)
		entryType, ok := raw["type"].(string)
		if !ok {
			return nil, fmt.Errorf("entry missing 'type' field")
		}
		if entryType != "function" {
			return nil, fmt.Errorf("only ABI functions supported, got: %s", entryType)
		}

		// Parse inputs
		inputs := []Param{}
		if rawInputs, ok := raw["inputs"].([]interface{}); ok {
			for _, rawInput := range rawInputs {
				inputMap, ok := rawInput.(map[string]interface{})
				if !ok {
					return nil, fmt.Errorf("invalid input format")
				}

				//nolint:errcheck // Name is optional in ABI, defaults to empty string if not present
				paramName, _ := inputMap["name"].(string)
				paramType, ok := inputMap["type"].(string)
				if !ok {
					return nil, fmt.Errorf("input missing 'type' field")
				}

				param, err := NewParam(paramName, paramType)
				if err != nil {
					return nil, fmt.Errorf("failed to create param '%s': %w", paramName, err)
				}

				inputs = append(inputs, *param)
			}
		}

		// Create ABI function entry
		entry := Entry{
			Name:   name,
			Inputs: inputs,
			Type:   Function,
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// FromJson creates a new ABI container from JSON string
func FromJson(jsonStr string) (*Abi, error) {
	entries, err := parseEntries(jsonStr)
	if err != nil {
		return nil, err
	}

	return &Abi{
		Entries: entries,
	}, nil
}

// EncodeFunction encodes a function call by name
func (a *Abi) EncodeFunction(name string, args []interface{}) ([]byte, error) {
	// Find function by name
	var foundEntry *Entry
	for i := range a.Entries {
		if a.Entries[i].Name == name {
			foundEntry = &a.Entries[i]
			break
		}
	}

	if foundEntry == nil {
		return nil, fmt.Errorf("function '%s' not found in ABI", name)
	}

	// Create AbiFunction and encode
	fn := &AbiFunction{
		Entry: *foundEntry,
	}

	return fn.Encode(args)
}

// DecodeFunction decodes a function call by matching signature
func (a *Abi) DecodeFunction(encoded []byte) ([]interface{}, error) {
	if len(encoded) < EncodedSignLength {
		return nil, fmt.Errorf("encoded data too short: %d bytes", len(encoded))
	}

	// Extract signature from encoded data
	signature := extractSignature(encoded)

	// Find matching function by signature
	var foundEntry *Entry
	for i := range a.Entries {
		entrySignature := extractSignature(a.Entries[i].EncodeSignature())
		if bytes.Equal(signature, entrySignature) {
			foundEntry = &a.Entries[i]
			break
		}
	}

	if foundEntry == nil {
		return nil, fmt.Errorf("no matching function found for signature: %x", signature)
	}

	// Create AbiFunction and decode
	fn := &AbiFunction{
		Entry: *foundEntry,
	}

	return fn.Decode(encoded)
}
