package postgres

import "testing"

func TestValidateDSN(t *testing.T) {
	tests := []struct {
		name    string
		dsn     string
		wantErr bool
	}{
		{name: "valid", dsn: "postgres://user:pass@localhost:5432/app", wantErr: false},
		{name: "invalid scheme", dsn: "http://localhost:5432/app", wantErr: true},
		{name: "missing host", dsn: "postgres:///app", wantErr: true},
		{name: "empty hostname with port", dsn: "postgres://:5432/app", wantErr: true},
		{name: "missing db", dsn: "postgres://localhost:5432", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDSN(tt.dsn)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateDSN() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
