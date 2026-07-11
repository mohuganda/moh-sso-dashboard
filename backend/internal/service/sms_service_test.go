package service

import "testing"

func TestNormalizePhoneNumber(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		phone       string
		countryCode string
		want        string
		wantErr     bool
	}{
		{
			name:        "local uganda number",
			phone:       "0772 123 456",
			countryCode: "+256",
			want:        "+256772123456",
		},
		{
			name:        "already international",
			phone:       "+256772123456",
			countryCode: "+256",
			want:        "+256772123456",
		},
		{
			name:        "international with 00 prefix",
			phone:       "00256772123456",
			countryCode: "+256",
			want:        "+256772123456",
		},
		{
			name:        "missing country code",
			phone:       "772123456",
			countryCode: "",
			wantErr:     true,
		},
		{
			name:        "empty phone",
			phone:       "",
			countryCode: "+256",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := NormalizePhoneNumber(tt.phone, tt.countryCode)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("NormalizePhoneNumber() = %q, want %q", got, tt.want)
			}
		})
	}
}
