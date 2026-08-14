package wallet

import "testing"

// Regression test for finding #19 (CWE-400): Argon2 cost fields are
// attacker-controlled whenever a keystore file is imported. argon2Parameters
// must reject excessive values before they reach crypto.DeriveKey /
// argon2.IDKey, which allocates memoryCost KiB and loops timeCost times; a
// crafted key file with a huge memoryCost (or timeCost) OOM-kills or
// CPU-freezes any process that imports it.
func TestArgon2ParametersRejectExcessiveCostParameters(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Argon2Params)
	}{
		{"excessive memory cost", func(p *Argon2Params) { p.MemoryCost = 4294967295 }},
		{"excessive time cost", func(p *Argon2Params) { p.TimeCost = 4294967295 }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			file := validMalformedTestEnvelope()
			params := file.Crypto.Argon2Params
			params.TimeCost = 1
			params.MemoryCost = 64 * 1024
			params.HashLength = 32
			params.Parallelism = 4
			test.mutate(params)

			if _, err := file.argon2Parameters(); err == nil {
				t.Fatalf("argon2Parameters accepted %s without an upper-bound error", test.name)
			}
		})
	}

	file := validMalformedTestEnvelope()
	file.Crypto.Argon2Params.TimeCost = 1
	file.Crypto.Argon2Params.MemoryCost = 64 * 1024
	file.Crypto.Argon2Params.HashLength = 32
	file.Crypto.Argon2Params.Parallelism = 4
	if _, err := file.argon2Parameters(); err != nil {
		t.Fatalf("argon2Parameters rejected the default-cost parameters: %v", err)
	}
}
