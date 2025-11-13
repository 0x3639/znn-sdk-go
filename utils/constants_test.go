package utils

import (
	"testing"
	"time"
)

func TestCoinDecimals(t *testing.T) {
	if CoinDecimals != 8 {
		t.Errorf("CoinDecimals = %d, want 8", CoinDecimals)
	}
}

func TestOneZnn(t *testing.T) {
	if OneZnn != 100000000 {
		t.Errorf("OneZnn = %d, want 100000000", OneZnn)
	}
}

func TestOneQsr(t *testing.T) {
	if OneQsr != 100000000 {
		t.Errorf("OneQsr = %d, want 100000000", OneQsr)
	}
}

func TestIntervalBetweenMomentums(t *testing.T) {
	expected := 10 * time.Second
	if IntervalBetweenMomentums != expected {
		t.Errorf("IntervalBetweenMomentums = %v, want %v", IntervalBetweenMomentums, expected)
	}
}
