package rpcvalidation

import (
	"strings"
	"testing"
)

func TestValidateLimit(t *testing.T) {
	t.Parallel()
	if err := ValidateLimit("test.method", "pageSize", MaxPageSize, MaxPageSize); err != nil {
		t.Fatalf("inclusive maximum was rejected: %v", err)
	}
	err := ValidateLimit("test.method", "pageSize", MaxPageSize+1, MaxPageSize)
	if err == nil {
		t.Fatal("value above maximum was accepted")
	}
	for _, fragment := range []string{"test.method", "pageSize", "1025", "1024"} {
		if !strings.Contains(err.Error(), fragment) {
			t.Fatalf("error %q does not contain %q", err, fragment)
		}
	}
}
