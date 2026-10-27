package cli

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/compose-spec/compose-go/loader"
	"github.com/compose-spec/compose-go/types"
	"github.com/labkit-project/vulnerable-lab-platform/internal/manifest"
	"github.com/labkit-project/vulnerable-lab-platform/internal/policy"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate <template-path>",
	Short: "Validate a lab template",
	Long:  `Validate a lab template against the schema and check required files.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		validateTemplate(args[0])
	},
}

func validateTemplate(templatePath string) {
	fmt.Printf("Validating template: %s\n", templatePath)
	fmt.Println()

	allPassed := true

	// Check required files. Supports both templates/<lab>/ and catalog/<vendor>/<product>/<cve>/ structures.
	requiredFiles := []string{
		"lab.yaml",
		"compose.yaml",
		"README.md",
	}

	// CVE catalog structure uses checks/smoke.sh, template structure uses tests/smoke.sh
	smokeTestPaths := []string{
		"checks/smoke.sh",
		"tests/smoke.sh",
	}

	// Template structure requires lessons/
	lessonFiles := []string{
		"lessons/objectives.md",
		"lessons/concepts.md",
		"lessons/remediation.md",
	}

	fmt.Println("Checking required files...")
	for _, file := range requiredFiles {
		filePath := filepath.Join(templatePath, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			fmt.Printf("  ✗ %s (missing)\n", file)
			allPassed = false
		} else {
			fmt.Printf("  ✓ %s\n", file)
		}
	}

	// Check smoke test exists in either location
	smokeFound := false
	for _, smokePath := range smokeTestPaths {
		if _, err := os.Stat(filepath.Join(templatePath, smokePath)); err == nil {
			fmt.Printf("  ✓ %s\n", smokePath)
			smokeFound = true
			break
		}
	}
	if !smokeFound {
		fmt.Printf("  ✗ smoke test not found (checked: %v)\n", smokeTestPaths)
		allPassed = false
	}

	// Check for CVE-specific references.yaml
	refsPath := filepath.Join(templatePath, "references.yaml")
	if _, err := os.Stat(refsPath); os.IsNotExist(err) {
		// references.yaml is optional for non-CVE labs
	} else {
		fmt.Printf("  ✓ references.yaml\n")
	}

	// Check lessons only if they would be expected (template structure)
	hasLessons := false
	for _, lessonFile := range lessonFiles {
		if _, err := os.Stat(filepath.Join(templatePath, lessonFile)); err == nil {
			hasLessons = true
			break
		}
	}
	if hasLessons {
		for _, lessonFile := range lessonFiles {
			filePath := filepath.Join(templatePath, lessonFile)
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				fmt.Printf("  ✗ %s (missing)\n", lessonFile)
				allPassed = false
			} else {
				fmt.Printf("  ✓ %s\n", lessonFile)
			}
		}
	}

	// Validate lab.yaml against JSON schema
	fmt.Println("\nValidating lab.yaml manifest...")
	manifestPath := filepath.Join(templatePath, "lab.yaml")
	var labManifest *manifest.Manifest
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		fmt.Printf("  ✗ lab.yaml not found\n")
		allPassed = false
	} else {
		loaded, err := manifestLoader.Load(manifestPath)
		if err != nil {
			fmt.Printf("  ✗ Manifest validation failed: %v\n", err)
			allPassed = false
		} else {
			labManifest = loaded
			fmt.Println("  ✓ lab.yaml is valid against schema")

			// CVE-specific validation
			if labManifest.Metadata.CVE != "" {
				fmt.Println("\n  Validating CVE catalog fields...")
				cvePassed := true

				// Required CVE fields
				if labManifest.Metadata.AffectedVersions == nil || len(labManifest.Metadata.AffectedVersions) == 0 {
					fmt.Printf("    ✗ CVE specified but affected_versions is empty\n")
					allPassed = false
					cvePassed = false
				} else {
					fmt.Printf("    ✓ affected_versions declared: %v\n", labManifest.Metadata.AffectedVersions)
				}
				if labManifest.Metadata.FixedVersions == nil || len(labManifest.Metadata.FixedVersions) == 0 {
					fmt.Printf("    ✗ CVE specified but fixed_versions is empty\n")
					allPassed = false
					cvePassed = false
				} else {
					fmt.Printf("    ✓ fixed_versions declared: %v\n", labManifest.Metadata.FixedVersions)
				}
				if labManifest.CoverageState == "" {
					fmt.Printf("    ✗ coverage_state is not set\n")
					allPassed = false
					cvePassed = false
				} else {
					fmt.Printf("    ✓ coverage_state: %s\n", labManifest.CoverageState)
				}

				// Validate severity vector consistency (calculates score from vector)
				if labManifest.Vulnerability != nil && labManifest.Vulnerability.Severity != "" {
					cvssOK, cvssMsg := validateCVSSVector(labManifest.Vulnerability.Severity)
					if !cvssOK {
						fmt.Printf("    ✗ %s\n", cvssMsg)
						allPassed = false
						cvePassed = false
					} else {
						fmt.Printf("    ✓ %s\n", cvssMsg)
					}

					// Validate provenance - required for reproducible/published, optional otherwise
					if labManifest.Vulnerability.Provenance == nil {
						if labManifest.CoverageState == manifest.CoveragePublished || labManifest.CoverageState == manifest.CoverageReproducible {
							fmt.Printf("    ✗ coverage_state is '%s' but vulnerability section missing provenance\n", labManifest.CoverageState)
							allPassed = false
							cvePassed = false
						} else {
							fmt.Printf("    ✓ provenance not required for '%s' coverage_state\n", labManifest.CoverageState)
						}
					} else {
						provOK, provMsg := validateProvenance(labManifest.Vulnerability.Provenance)
						if !provOK {
							fmt.Printf("    ✗ %s\n", provMsg)
							allPassed = false
							cvePassed = false
						} else {
							fmt.Printf("    ✓ %s\n", provMsg)
						}
					}

					// Validate verification claims against coverage_state
					if labManifest.Vulnerability.Verification != nil {
						verOK, verMsg := verifyClaimsAgainstCoverage(
							labManifest.CoverageState,
							labManifest.Vulnerability.Verification,
						)
						if !verOK {
							fmt.Printf("    ✗ %s\n", verMsg)
							allPassed = false
							cvePassed = false
						} else {
							fmt.Printf("    ✓ %s\n", verMsg)
						}
					} else if labManifest.CoverageState == manifest.CoveragePublished {
						fmt.Printf("    ✗ coverage_state is 'published' but verification section is missing\n")
						allPassed = false
						cvePassed = false
					}
				}

				if labManifest.Vulnerability == nil {
					fmt.Printf("    ✗ CVE specified but vulnerability section is missing\n")
					allPassed = false
					cvePassed = false
				}

				// Check references.yaml exists for CVE labs
				refsPath := filepath.Join(templatePath, "references.yaml")
				if _, err := os.Stat(refsPath); os.IsNotExist(err) {
					fmt.Printf("    ✗ references.yaml missing for CVE lab\n")
					allPassed = false
					cvePassed = false
				} else {
					fmt.Printf("    ✓ references.yaml present\n")
				}

				if cvePassed {
					fmt.Printf("    ✓ All CVE catalog field checks passed\n")
				}
			}
		}
	}

	// Validate compose.yaml
	fmt.Println("\nValidating compose.yaml...")
	composePath := filepath.Join(templatePath, "compose.yaml")
	if _, err := os.Stat(composePath); os.IsNotExist(err) {
		fmt.Printf("  ✗ compose.yaml not found\n")
		allPassed = false
	} else {
		fmt.Println("  ✓ compose.yaml exists")

		// Validate the resolved Compose configuration against safety policies.
		// This inspects the effective configuration after interpolation,
		// overrides, and defaults have been applied (equivalent to
		// `docker compose config --format json`).
		data, err := os.ReadFile(composePath)
		if err != nil {
			fmt.Printf("  ✗ compose.yaml unreadable: %v\n", err)
			allPassed = false
		} else {
			project, err := loader.Load(types.ConfigDetails{
				ConfigFiles: []types.ConfigFile{
					{Filename: composePath, Content: data},
				},
				WorkingDir: filepath.Dir(composePath),
			}, func(o *loader.Options) {
				projectName := filepath.Base(filepath.Dir(composePath))
				projectName = strings.Map(func(r rune) rune {
					if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
						return r
					}
					return '-'
				}, projectName)
				if len(projectName) > 0 && projectName[0] < '0' {
					projectName = "lab" + projectName
				}
				o.SetProjectName(projectName, true)
			})
			if err != nil {
				fmt.Printf("  ✗ compose.yaml failed to resolve: %v\n", err)
				allPassed = false
			} else {
				policyResult := policy.NewValidator().ValidateComposeProject(project, composePath)
				if !policyResult.Passed {
					allPassed = false
					fmt.Println("\n  Security policy violations:")
					for _, errMsg := range policyResult.Errors {
						fmt.Printf("    ✗ %s\n", errMsg)
					}
				} else {
					fmt.Println("  ✓ compose.yaml passes safety policies")
				}

				// Check agreement between manifest endpoints and resolved Compose configuration
				if labManifest != nil && len(labManifest.Endpoints) > 0 {
					fmt.Println("\n  Checking endpoint/compose agreement...")
					endpointOK := checkEndpointComposeAgreement(labManifest, project)
					if !endpointOK {
						allPassed = false
					}
				}

				// Check agreement between declared network restrictions and effective Compose config
				if labManifest != nil {
					fmt.Println("\n  Checking network restriction agreement...")
					netOK := checkNetworkAgreement(labManifest, project)
					if !netOK {
						allPassed = false
					}
				}
			}
		}
	}

	// Check smoke test is executable (either checks/smoke.sh or tests/smoke.sh)
	smokeFound = false
	for _, smokePath := range smokeTestPaths {
		smokeTestPath := filepath.Join(templatePath, smokePath)
		if info, err := os.Stat(smokeTestPath); err == nil {
			if info.Mode().Perm()&0111 != 0 {
				fmt.Printf("\n✓ %s is executable\n", smokePath)
			} else {
				fmt.Printf("\n⚠ %s is not executable (run: chmod +x %s)\n", smokePath, smokePath)
			}
			smokeFound = true
			break
		}
	}
	if !smokeFound {
		fmt.Printf("\n⚠ smoke test not found (checked: %v)\n", smokeTestPaths)
	}

	fmt.Println()
	if allPassed {
		fmt.Println("✓ Template validation passed")
	} else {
		fmt.Println("✗ Template validation failed")
		os.Exit(1)
	}
}

// cvssMetricValues maps CVSS v3.1 metric values to their numerical coefficients.
var cvssMetricValues = map[string]map[string]float64{
	"AV": {"N": 0.85, "A": 0.62, "L": 0.55, "P": 0.20},
	"AC": {"L": 0.77, "H": 0.44},
	"PR": {"U": 0.85, "N": 0.85, "L": 0.62, "H": 0.27},
	"UI": {"N": 0.85, "R": 0.62},
	"C":  {"N": 0.00, "L": 0.22, "H": 0.56},
	"I":  {"N": 0.00, "L": 0.22, "H": 0.56},
	"A":  {"N": 0.00, "L": 0.22, "H": 0.56},
}

// validCVSSMetrics lists the allowed values for each required CVSS v3.1 metric.
var validCVSSMetrics = map[string][]string{
	"AV": {"N", "A", "L", "P"},
	"AC": {"L", "H"},
	"PR": {"U", "N", "L", "H"},
	"UI": {"N", "R"},
	"S":  {"U", "C"},
	"C":  {"N", "L", "H"},
	"I":  {"N", "L", "H"},
	"A":  {"N", "L", "H"},
}

// validateCVSSVector parses a CVSS v3.1 vector, calculates the base score, and
// compares it against the declared score. Returns an error if the vector is
// malformed, uses invalid metric values, or the calculated score mismatches
// the declared score by more than 0.1.
func validateCVSSVector(severity string) (bool, string) {
	// Extract CVSS vector from severity string like "7.5 (CVSS:3.1/AV:N/AC:L/...)"
	vectorMatch := regexp.MustCompile(`CVSS:3\.1/(.+)\)`).FindStringSubmatch(severity)
	if len(vectorMatch) < 2 {
		return false, "severity does not contain a valid CVSS v3.1 vector"
	}
	vector := vectorMatch[1]

	// Extract declared score
	scoreMatch := regexp.MustCompile(`^([0-9]+(?:\.[0-9]+)?)`).FindStringSubmatch(severity)
	if len(scoreMatch) < 2 {
		return false, "severity does not include a numeric score"
	}
	declaredScore, err := strconv.ParseFloat(scoreMatch[1], 64)
	if err != nil {
		return false, fmt.Sprintf("invalid CVSS score '%s': %v", scoreMatch[1], err)
	}

	// Parse metrics
	metrics := make(map[string]string)
	for _, part := range strings.Split(vector, "/") {
		kv := strings.SplitN(part, ":", 2)
		if len(kv) == 2 {
			metrics[kv[0]] = kv[1]
		}
	}

	// Validate all required metrics are present with valid values
	for metric, validValues := range validCVSSMetrics {
		val, ok := metrics[metric]
		if !ok {
			return false, fmt.Sprintf("missing required CVSS metric '%s'", metric)
		}
		valid := false
		for _, v := range validValues {
			if val == v {
				valid = true
				break
			}
		}
		if !valid {
			return false, fmt.Sprintf("invalid %s value '%s' in CVSS vector (valid: %v)", metric, val, validValues)
		}
	}

	// Calculate CVSS v3.1 base score
	calculatedScore := calculateCVSSBaseScore(metrics)

	// Compare with declared score (tolerance: 0.1 due to rounding)
	if math.Abs(calculatedScore-declaredScore) > 0.1 {
		return false, fmt.Sprintf("CVSS score mismatch: declared %.1f, vector calculates %.1f", declaredScore, calculatedScore)
	}

	return true, fmt.Sprintf("CVSS vector valid (score %.1f)", calculatedScore)
}

// calculateCVSSBaseScore computes the CVSS v3.1 base score from parsed metrics.
func calculateCVSSBaseScore(metrics map[string]string) float64 {
	av := cvssMetricValues["AV"][metrics["AV"]]
	ac := cvssMetricValues["AC"][metrics["AC"]]
	ui := cvssMetricValues["UI"][metrics["UI"]]
	scope := metrics["S"]
	c := cvssMetricValues["C"][metrics["C"]]
	i := cvssMetricValues["I"][metrics["I"]]
	a := cvssMetricValues["A"][metrics["A"]]

	// Impact Sub-Score
	isc := 1 - ((1 - c) * (1 - i) * (1 - a))

	// Impact
	var impact float64
	if scope == "U" {
		impact = 6.42 * isc
	} else {
		impact = 7.52*(isc-0.029) - 3.25*math.Pow(isc-0.02, 15)
	}

	// Exploitability
	// PR metric depends on Scope
	prValue := cvssMetricValues["PR"][metrics["PR"]]
	if scope == "C" && metrics["PR"] == "U" {
		prValue = 0.68
	}
	exploitability := 8.22 * av * ac * prValue * ui

	// Base Score
	var baseScore float64
	if scope == "U" {
		if impact <= 0 {
			baseScore = 0
		} else {
			baseScore = math.Min(impact+exploitability, 10)
		}
	} else {
		if impact <= 0 {
			baseScore = 0
		} else {
			baseScore = math.Min(1.08*(impact+exploitability), 10)
		}
	}

	// Round up to 1 decimal place
	return roundUp(baseScore, 1)
}

// roundUp rounds a value up to the specified number of decimal places.
func roundUp(val float64, places int) float64 {
	shift := math.Pow(10, float64(places))
	return math.Ceil(val*shift) / shift
}

// validateProvenance rejects placeholder or zero-value provenance data.
func validateProvenance(prov *manifest.Provenance) (bool, string) {
	if prov.ImageDigest == "" || prov.ImageDigest == "sha256:placeholder" || prov.ImageDigest == "sha256:unverified" {
		return false, "provenance.image_digest is a placeholder or empty"
	}
	if !regexp.MustCompile(`^sha256:[a-f0-9]{64}$`).MatchString(prov.ImageDigest) {
		return false, "provenance.image_digest is not a valid SHA-256 digest"
	}
	if prov.SourceRevision == "" || prov.SourceRevision == "placeholder" || prov.SourceRevision == "unverified" {
		return false, "provenance.source_revision is a placeholder or empty"
	}
	if prov.VerificationDate == "" {
		return false, "provenance.verification_date is empty"
	}
	if prov.VerifiedBy == "" {
		return false, "provenance.verified_by is empty"
	}
	return true, "provenance data is present and non-placeholder"
}

// verifyClaimsAgainstCoverage checks that verification claims are consistent
// with the declared coverage_state.
func verifyClaimsAgainstCoverage(state manifest.CoverageState, ver *manifest.Verification) (bool, string) {
	if state == manifest.CoveragePublished {
		if !ver.AffectedVersionConfirmed {
			return false, "coverage_state is 'published' but affected_version_confirmed is false"
		}
		if !ver.FixedVersionConfirmed {
			return false, "coverage_state is 'published' but fixed_version_confirmed is false"
		}
		if !ver.ReproductionStepsDocumented {
			return false, "coverage_state is 'published' but reproduction_steps_documented is false"
		}
		return true, "verification claims consistent with 'published' coverage_state"
	}
	if state == manifest.CoverageReproducible {
		if !ver.AffectedVersionConfirmed {
			return false, "coverage_state is 'reproducible' but affected_version_confirmed is false"
		}
		return true, "verification claims consistent with 'reproducible' coverage_state"
	}
	if state == manifest.CoverageTriaged {
		return true, "verification claims consistent with 'triaged' coverage_state"
	}
	if state == manifest.CoverageDiscovered {
		return true, "verification claims consistent with 'discovered' coverage_state"
	}
	if state == manifest.CoverageRetired {
		return true, "verification claims consistent with 'retired' coverage_state"
	}
	return false, fmt.Sprintf("unknown coverage_state '%s'", state)
}

// checkEndpointComposeAgreement verifies that declared endpoints match
// the resolved Compose service configuration, including host binding.
func checkEndpointComposeAgreement(manifest *manifest.Manifest, project *types.Project) bool {
	allOK := true
	for _, ep := range manifest.Endpoints {
		var foundService *types.ServiceConfig
		for i := range project.Services {
			if project.Services[i].Name == ep.Service {
				foundService = &project.Services[i]
				break
			}
		}
		if foundService == nil {
			fmt.Printf("    ✗ endpoint '%s' references unknown service '%s'\n", ep.ID, ep.Service)
			allOK = false
			continue
		}
		found := false
		for _, port := range foundService.Ports {
			if uint32(ep.ContainerPort) == port.Target {
				found = true
				// Check host binding agreement
				hostIP := port.HostIP
				if hostIP == "" {
					hostIP = "0.0.0.0" // Docker Compose default
				}
				if ep.HostBinding != "" && ep.HostBinding != hostIP && ep.HostBinding != "localhost" {
					if hostIP == "0.0.0.0" && ep.HostBinding == "127.0.0.1" {
						fmt.Printf("    ✗ endpoint '%s' declares hostBinding 127.0.0.1 but Compose binds to 0.0.0.0\n", ep.ID)
						allOK = false
					} else if hostIP == "127.0.0.1" && ep.HostBinding == "0.0.0.0" {
						fmt.Printf("    ✗ endpoint '%s' declares hostBinding 0.0.0.0 but Compose binds to 127.0.0.1\n", ep.ID)
						allOK = false
					} else if ep.HostBinding != hostIP {
						fmt.Printf("    ✗ endpoint '%s' declares hostBinding %s but Compose binds to %s\n", ep.ID, ep.HostBinding, hostIP)
						allOK = false
					}
				}
				break
			}
		}
		if !found {
			fmt.Printf("    ✗ endpoint '%s' declares containerPort %d but service '%s' does not publish it\n",
				ep.ID, ep.ContainerPort, ep.Service)
			allOK = false
		}
	}
	if allOK {
		fmt.Println("    ✓ endpoints agree with resolved Compose configuration")
	}
	return allOK
}

// checkNetworkAgreement verifies that declared network restrictions match
// the effective Compose configuration.
func checkNetworkAgreement(manifest *manifest.Manifest, project *types.Project) bool {
	if manifest.Runtime.Network.Outbound != "restricted" {
		fmt.Println("    ✓ no restricted outbound requirement; skipping network agreement check")
		return true
	}

	// For restricted outbound, Compose networks should be internal
	for _, net := range project.Networks {
		if net.Internal {
			fmt.Println("    ✓ restricted outbound matches internal Compose network")
			return true
		}
	}
	fmt.Println("    ✗ manifest declares restricted outbound but no Compose network is internal")
	return false
}
