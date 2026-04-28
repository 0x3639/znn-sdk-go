package embedded

import (
	"math/big"
	"testing"
)

func TestPlasmaApi_GetPlasmaByQsr(t *testing.T) {
	api := NewPlasmaApi(nil)

	tests := []struct {
		name string
		qsr  *big.Int
		want *big.Int
	}{
		{name: "zero", qsr: big.NewInt(0), want: big.NewInt(0)},
		{name: "1 QSR", qsr: big.NewInt(1), want: big.NewInt(2100)},
		{name: "10 QSR", qsr: big.NewInt(10), want: big.NewInt(21000)},
		{name: "nil treated as zero", qsr: nil, want: big.NewInt(0)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := api.GetPlasmaByQsr(tc.qsr)
			if got.Cmp(tc.want) != 0 {
				t.Errorf("GetPlasmaByQsr(%v) = %v, want %v", tc.qsr, got, tc.want)
			}
		})
	}
}
