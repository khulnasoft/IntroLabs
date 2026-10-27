package cli

import "testing"

func TestValidateCVSSVectorCorrectScore(t *testing.T) {
	cases := []struct {
		name     string
		severity string
	}{
		{"CVE-2026-34486 high", "7.5 (CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:N/A:N)"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ok, msg := validateCVSSVector(tc.severity)
			if !ok {
				t.Fatalf("expected valid CVSS vector, got invalid: %s", msg)
			}
		})
	}
}

func TestValidateCVSSVectorWrongScore(t *testing.T) {
	// Declared 7.5 but vector calculates a different score (AV:L instead of AV:N)
	severity := "7.5 (CVSS:3.1/AV:L/AC:L/PR:N/UI:N/S:U/C:H/I:N/A:N)"
	ok, msg := validateCVSSVector(severity)
	if ok {
		t.Fatalf("expected invalid CVSS (score mismatch), got valid: %s", msg)
	}
}

func TestValidateCVSSVectorMissingMetric(t *testing.T) {
	severity := "7.5 (CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:N)" // missing A
	ok, msg := validateCVSSVector(severity)
	if ok {
		t.Fatalf("expected invalid CVSS (missing metric), got valid: %s", msg)
	}
}

func TestValidateCVSSVectorInvalidValue(t *testing.T) {
	severity := "7.5 (CVSS:3.1/AV:X/AC:L/PR:N/UI:N/S:U/C:H/I:N/A:N)" // invalid AV:X
	ok, msg := validateCVSSVector(severity)
	if ok {
		t.Fatalf("expected invalid CVSS (bad metric value), got valid: %s", msg)
	}
}

func TestValidateCVSSVectorNoVector(t *testing.T) {
	severity := "7.5"
	ok, msg := validateCVSSVector(severity)
	if ok {
		t.Fatalf("expected invalid CVSS (no vector), got valid: %s", msg)
	}
}

func TestValidateCVSSVectorNoScore(t *testing.T) {
	severity := "(CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:N/A:N)"
	ok, msg := validateCVSSVector(severity)
	if ok {
		t.Fatalf("expected invalid CVSS (no score), got valid: %s", msg)
	}
}
