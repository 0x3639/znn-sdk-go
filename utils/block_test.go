package utils

import "testing"

func TestIsSendBlock(t *testing.T) {
	testCases := []struct {
		blockType int
		expected  bool
	}{
		{BlockTypeUserSend, true},
		{BlockTypeContractSend, true},
		{BlockTypeUserReceive, false},
		{BlockTypeGenesisReceive, false},
		{BlockTypeContractReceive, false},
		{BlockTypeUnknown, false},
	}

	for _, tc := range testCases {
		result := IsSendBlock(tc.blockType)
		if result != tc.expected {
			t.Errorf("IsSendBlock(%d) = %v, want %v", tc.blockType, result, tc.expected)
		}
	}
}

func TestIsReceiveBlock(t *testing.T) {
	testCases := []struct {
		blockType int
		expected  bool
	}{
		{BlockTypeUserReceive, true},
		{BlockTypeGenesisReceive, true},
		{BlockTypeContractReceive, true},
		{BlockTypeUserSend, false},
		{BlockTypeContractSend, false},
		{BlockTypeUnknown, false},
	}

	for _, tc := range testCases {
		result := IsReceiveBlock(tc.blockType)
		if result != tc.expected {
			t.Errorf("IsReceiveBlock(%d) = %v, want %v", tc.blockType, result, tc.expected)
		}
	}
}
