package generator

import (
	"os"
	"path/filepath"

	"github.com/labkit-project/vulnerable-lab-platform/internal/manifest"
	"github.com/labkit-project/vulnerable-lab-platform/internal/policy"
)

// GateResult is the result of a single verification gate check.
type GateResult struct {
	Check   string
	Passed  bool
	Message string
}

// GateResultSlice is a slice of GateResult.
type GateResultSlice []GateResult

// AllPassed reports whether all gate checks passed.
func (g GateResultSlice) AllPassed() bool {
	for _, r := range g {
		if !r.Passed {
			return false
		}
	}
	return true
}

// FailedChecks returns the checks that did not pass.
func (g GateResultSlice) FailedChecks() []GateResult {
	var failed []GateResult
	for _, r := range g {
		if !r.Passed {
			failed = append(failed, r)
		}
	}
	return failed
}

// Verifier runs the verification gate checks for a generated template.
type Verifier struct {
	manifestLoader  *manifest.Loader
	policyValidator *policy.Validator
}

// NewVerifier creates a Verifier.
func NewVerifier(ml *manifest.Loader, pv *policy.Validator) *Verifier {
	return &Verifier{
		manifestLoader:  ml,
		policyValidator: pv,
	}
}

// Verify runs all six gate checks and returns the results.
func (v *Verifier) Verify(templatePath string) GateResultSlice {
	return GateResultSlice{
		v.checkSchema(templatePath),
		v.checkCompose(templatePath),
		v.checkPolicy(templatePath),
		v.checkHealthReadiness(templatePath),
		v.checkExpectedOutput(templatePath),
		v.checkCleanup(templatePath),
	}
}

// checkSchema validates lab.yaml against the JSON schema.
func (v *Verifier) checkSchema(templatePath string) GateResult {
	manifestPath := filepath.Join(templatePath, "lab.yaml")
	_, err := v.manifestLoader.Load(manifestPath)
	if err != nil {
		return GateResult{Check: "schema validation", Passed: false, Message: err.Error()}
	}
	return GateResult{Check: "schema validation", Passed: true, Message: "lab.yaml is valid against schema"}
}

// checkCompose validates that compose.yaml exists.
func (v *Verifier) checkCompose(templatePath string) GateResult {
	composePath := filepath.Join(templatePath, "compose.yaml")
	if _, err := os.Stat(composePath); err != nil {
		return GateResult{Check: "compose validation", Passed: false, Message: "compose.yaml not found: " + err.Error()}
	}
	return GateResult{Check: "compose validation", Passed: true, Message: "compose.yaml exists"}
}

// checkPolicy runs safety policy checks against the compose configuration.
func (v *Verifier) checkPolicy(templatePath string) GateResult {
	composePath := filepath.Join(templatePath, "compose.yaml")
	result := v.policyValidator.ValidateCompose(composePath)
	if !result.Passed {
		return GateResult{Check: "policy checks", Passed: false, Message: "policy violations: " + joinErrors(result.Errors)}
	}
	return GateResult{Check: "policy checks", Passed: true, Message: "compose.yaml passes safety policies"}
}

// checkHealthReadiness verifies endpoints are bound to loopback.
func (v *Verifier) checkHealthReadiness(templatePath string) GateResult {
	manifestPath := filepath.Join(templatePath, "lab.yaml")
	m, err := v.manifestLoader.Load(manifestPath)
	if err != nil {
		return GateResult{Check: "health/readiness", Passed: false, Message: "cannot load manifest: " + err.Error()}
	}
	if len(m.Endpoints) == 0 {
		return GateResult{Check: "health/readiness", Passed: false, Message: "no endpoints declared"}
	}
	for _, ep := range m.Endpoints {
		if ep.HostBinding != "127.0.0.1" {
			return GateResult{Check: "health/readiness", Passed: false, Message: "endpoint " + ep.ID + " not bound to loopback (hostBinding=" + ep.HostBinding + ")"}
		}
	}
	return GateResult{Check: "health/readiness", Passed: true, Message: "all endpoints bound to loopback"}
}

// checkExpectedOutput verifies smoke test exists.
func (v *Verifier) checkExpectedOutput(templatePath string) GateResult {
	smokePath := filepath.Join(templatePath, "tests", "smoke.sh")
	if _, err := os.Stat(smokePath); err != nil {
		return GateResult{Check: "expected output", Passed: false, Message: "tests/smoke.sh not found: " + err.Error()}
	}
	return GateResult{Check: "expected output", Passed: true, Message: "smoke test present"}
}

// checkCleanup verifies deterministic cleanup behavior is documented.
func (v *Verifier) checkCleanup(templatePath string) GateResult {
	manifestPath := filepath.Join(templatePath, "lab.yaml")
	m, err := v.manifestLoader.Load(manifestPath)
	if err != nil {
		return GateResult{Check: "cleanup behavior", Passed: false, Message: "cannot load manifest: " + err.Error()}
	}
	if m.Vulnerability == nil || m.Vulnerability.Safety == nil || m.Vulnerability.Safety.CleanupBehavior == "" {
		return GateResult{Check: "cleanup behavior", Passed: false, Message: "no cleanup behavior documented"}
	}
	return GateResult{Check: "cleanup behavior", Passed: true, Message: "cleanup behavior documented: " + m.Vulnerability.Safety.CleanupBehavior}
}

func joinErrors(errs []string) string {
	result := ""
	for i, e := range errs {
		if i > 0 {
			result += "; "
		}
		result += e
	}
	return result
}
