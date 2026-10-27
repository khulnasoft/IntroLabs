package generator

import (
	"strings"
	"testing"
)

func TestIsValidCVE(t *testing.T) {
	valid := []string{"CVE-2026-34486", "CVE-2024-12345", "CVE-1999-0001"}
	invalid := []string{"cve-2026-34486", "CVE-26-34486", "CVE-2026-344", "NOT-A-CVE", ""}

	for _, v := range valid {
		if !IsValidCVE(v) {
			t.Errorf("IsValidCVE(%q) = false, want true", v)
		}
	}
	for _, inv := range invalid {
		if IsValidCVE(inv) {
			t.Errorf("IsValidCVE(%q) = true, want false", inv)
		}
	}
}

func TestCanTransition(t *testing.T) {
	tests := []struct {
		from, to TemplateState
		want     bool
	}{
		{StateDraft, StateMetadataOnly, true},
		{StateMetadataOnly, StateReproductionPending, true},
		{StateReproductionPending, StateNeedsVerification, true},
		{StateNeedsVerification, StatePOCVerified, true},
		{StatePOCVerified, StateVerified, true},
		{StateVerified, StatePublished, true},
		{StateDraft, StateReproductionPending, false}, // skip
		{StateDraft, StateVerified, false},            // skip
		{StateVerified, StateMetadataOnly, false},     // backward
		{StatePublished, StateVerified, false},        // backward
		{StateDraft, StateDraft, true},                // idempotent
		{StateDeprecated, StatePublished, false},      // cannot un-deprecate
	}

	for _, tc := range tests {
		got := CanTransition(tc.from, tc.to)
		if got != tc.want {
			t.Errorf("CanTransition(%s, %s) = %v, want %v", tc.from, tc.to, got, tc.want)
		}
	}
}

func TestTemplateStateMethods(t *testing.T) {
	if !StateDraft.IsValid() {
		t.Error("StateDraft.IsValid() = false, want true")
	}
	if (TemplateState("invalid")).IsValid() {
		t.Error("invalid.IsValid() = true, want false")
	}
	if StateDraft.String() != "draft" {
		t.Errorf("StateDraft.String() = %q, want %q", StateDraft.String(), "draft")
	}
	if !StateVerified.IsPublishable() {
		t.Error("StateVerified.IsPublishable() = false, want true")
	}
	if StateDraft.IsPublishable() {
		t.Error("StateDraft.IsPublishable() = true, want false")
	}
	if !StateVerified.IsRunnable() {
		t.Error("StateVerified.IsRunnable() = false, want true")
	}
	if StateDraft.IsRunnable() {
		t.Error("StateDraft.IsRunnable() = true, want false")
	}
	if !StatePOCVerified.IsPOCVerified() {
		t.Error("StatePOCVerified.IsPOCVerified() = false, want true")
	}
	if !StateVerified.IsPOCVerified() {
		t.Error("StateVerified.IsPOCVerified() = false, want true")
	}
	if StateNeedsVerification.IsPOCVerified() {
		t.Error("StateNeedsVerification.IsPOCVerified() = true, want false")
	}
}

func TestNextState(t *testing.T) {
	if NextState(StateDraft) != StateMetadataOnly {
		t.Errorf("NextState(draft) = %s, want %s", NextState(StateDraft), StateMetadataOnly)
	}
	if NextState(StatePublished) != StateDeprecated {
		t.Errorf("NextState(published) = %s, want %s", NextState(StatePublished), StateDeprecated)
	}
	if NextState(StateDeprecated) != StateDeprecated {
		t.Errorf("NextState(deprecated) = %s, want %s", NextState(StateDeprecated), StateDeprecated)
	}
}

func TestGateResultSlice(t *testing.T) {
	allPass := GateResultSlice{
		{Check: "a", Passed: true, Message: "ok"},
		{Check: "b", Passed: true, Message: "ok"},
	}
	if !allPass.AllPassed() {
		t.Error("AllPassed() = false, want true")
	}
	if len(allPass.FailedChecks()) != 0 {
		t.Errorf("FailedChecks() = %d, want 0", len(allPass.FailedChecks()))
	}

	oneFail := GateResultSlice{
		{Check: "a", Passed: true, Message: "ok"},
		{Check: "b", Passed: false, Message: "fail"},
	}
	if oneFail.AllPassed() {
		t.Error("AllPassed() = true, want false")
	}
	if len(oneFail.FailedChecks()) != 1 {
		t.Errorf("FailedChecks() = %d, want 1", len(oneFail.FailedChecks()))
	}
	if oneFail.FailedChecks()[0].Check != "b" {
		t.Errorf("FailedChecks()[0].Check = %q, want %q", oneFail.FailedChecks()[0].Check, "b")
	}
}

func TestNormalizeNonActionable(t *testing.T) {
	// Invalid CVE
	cve := &NVDCVEItem{ID: "NOT-A-CVE"}
	nc := NewNormalizer().Normalize(cve, "")
	if nc.Actionable {
		t.Error("nc.Actionable = true, want false for invalid CVE")
	}
	if nc.NonActionableReason == "" {
		t.Error("NonActionableReason should be set for invalid CVE")
	}

	// CVE with no affected versions
	cve2 := &NVDCVEItem{ID: "CVE-2026-0001"}
	nc2 := NewNormalizer().Normalize(cve2, "")
	if nc2.Actionable {
		t.Error("nc2.Actionable = true, want false when no versions")
	}
	if !strings.Contains(nc2.NonActionableReason, "versions") {
		t.Errorf("NonActionableReason = %q, should mention versions", nc2.NonActionableReason)
	}
}
