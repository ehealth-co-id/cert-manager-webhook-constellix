package main

import (
	"strings"
	"testing"

	"github.com/cert-manager/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
)

func TestNormalizeDNSName(t *testing.T) {
	t.Parallel()
	if got := normalizeDNSName(" Example.COM. "); got != "example.com" {
		t.Fatalf("normalizeDNSName: got %q", got)
	}
}

func TestChallengeApex(t *testing.T) {
	t.Parallel()
	ch := &v1alpha1.ChallengeRequest{
		ResolvedZone: "Zone.EXAMPLE.com.",
	}
	if got := challengeApex(ch, "other.org"); got != "zone.example.com" {
		t.Fatalf("challengeApex with ResolvedZone: got %q", got)
	}
	ch2 := &v1alpha1.ChallengeRequest{ResolvedZone: ""}
	if got := challengeApex(ch2, "fallback.io"); got != "fallback.io" {
		t.Fatalf("challengeApex fallback: got %q", got)
	}
}

func TestResolveZoneID(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		cfg     constellixDNSProviderConfig
		apex    string
		want    int
		wantErr string
	}{
		{
			name: "legacy zoneId only",
			cfg: constellixDNSProviderConfig{
				ZoneId: 42,
			},
			apex: "example.com",
			want: 42,
		},
		{
			name: "zones overrides legacy when both set",
			cfg: constellixDNSProviderConfig{
				ZoneId: 999,
				Zones: []zoneMapping{
					{DNSName: "example.com", ZoneID: 111},
				},
			},
			apex: "example.com",
			want: 111,
		},
		{
			name: "longest suffix wins",
			cfg: constellixDNSProviderConfig{
				Zones: []zoneMapping{
					{DNSName: "example.com", ZoneID: 1},
					{DNSName: "sub.example.com", ZoneID: 2},
				},
			},
			apex: "sub.example.com",
			want: 2,
		},
		{
			name: "suffix match to parent zone",
			cfg: constellixDNSProviderConfig{
				Zones: []zoneMapping{
					{DNSName: "example.com", ZoneID: 10},
				},
			},
			apex: "foo.bar.example.com",
			want: 10,
		},
		{
			name: "exact match in zones",
			cfg: constellixDNSProviderConfig{
				Zones: []zoneMapping{
					{DNSName: "other.org", ZoneID: 5},
				},
			},
			apex: "other.org",
			want: 5,
		},
		{
			name: "empty apex",
			cfg: constellixDNSProviderConfig{
				ZoneId: 1,
			},
			apex:    "",
			wantErr: "empty DNS zone apex",
		},
		{
			name:    "no zoneId and no zones",
			cfg:     constellixDNSProviderConfig{},
			apex:    "example.com",
			wantErr: "solver config must set zoneId or non-empty zones",
		},
		{
			name: "zones non-empty no match",
			cfg: constellixDNSProviderConfig{
				ZoneId: 100,
				Zones: []zoneMapping{
					{DNSName: "other.org", ZoneID: 1},
				},
			},
			apex:    "example.com",
			wantErr: "no zones entry matches DNS apex",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.cfg.resolveZoneID(tt.apex)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatal("expected error")
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error %q should contain %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("resolveZoneID: got %d want %d", got, tt.want)
			}
		})
	}
}
