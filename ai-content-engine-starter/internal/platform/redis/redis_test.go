package redis

import "testing"

func TestValidateAddr(t *testing.T) {
	tests := []struct {
		name    string
		addr    string
		wantErr bool
	}{
		{name: "valid host port", addr: "localhost:6379", wantErr: false},
		{name: "valid ip port", addr: "127.0.0.1:6379", wantErr: false},
		{name: "missing port", addr: "localhost", wantErr: true},
		{name: "missing host", addr: ":6379", wantErr: true},
		{name: "non-numeric port", addr: "localhost:abcd", wantErr: true},
		{name: "port out of range", addr: "localhost:70000", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAddr(tt.addr)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateAddr() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
