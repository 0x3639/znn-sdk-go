package wallet

import "testing"

func TestCoinType(t *testing.T) {
	if CoinType != "73404" {
		t.Errorf("CoinType = %s, want 73404", CoinType)
	}
}

func TestDerivationPath(t *testing.T) {
	expected := "m/44'/73404'"
	if DerivationPath != expected {
		t.Errorf("DerivationPath = %s, want %s", DerivationPath, expected)
	}
}

func TestGetDerivationAccount_Zero(t *testing.T) {
	path := GetDerivationAccount(0)
	expected := "m/44'/73404'/0'"

	if path != expected {
		t.Errorf("GetDerivationAccount(0) = %s, want %s", path, expected)
	}
}

func TestGetDerivationAccount_Five(t *testing.T) {
	path := GetDerivationAccount(5)
	expected := "m/44'/73404'/5'"

	if path != expected {
		t.Errorf("GetDerivationAccount(5) = %s, want %s", path, expected)
	}
}

func TestGetDerivationAccount_Large(t *testing.T) {
	path := GetDerivationAccount(100)
	expected := "m/44'/73404'/100'"

	if path != expected {
		t.Errorf("GetDerivationAccount(100) = %s, want %s", path, expected)
	}
}
