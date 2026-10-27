package generator

import (
	"fmt"
	"strings"
)

// NormalizedCVE is the structured representation of a CVE used for template generation.
type NormalizedCVE struct {
	CVEID               string
	AdvisoryURL         string
	AffectedProduct     string
	AffectedVersions    []string
	FixedVersions       []string
	Component           string
	TechnologyStack     string
	RequiredRuntime     string
	PocReferences       []string
	CVSSVector          string
	CVSSScore           float64
	Actionable          bool
	NonActionableReason string
}

// Normalizer converts raw NVD CVE data into normalized metadata.
type Normalizer struct{}

// NewNormalizer creates a Normalizer.
func NewNormalizer() *Normalizer {
	return &Normalizer{}
}

// Normalize converts an NVDCVEItem into a NormalizedCVE.
func (n *Normalizer) Normalize(cve *NVDCVEItem, advisoryURL string) *NormalizedCVE {
	nc := &NormalizedCVE{
		CVEID:       cve.ID,
		AdvisoryURL: advisoryURL,
		CVSSVector:  "",
		CVSSScore:   0,
		Actionable:  false,
	}

	vector, score := cve.CVSSv31()
	nc.CVSSVector = vector
	nc.CVSSScore = score

	// Extract description for context
	desc := cve.Description()
	nc.Component = extractComponent(desc, cve.ID)
	nc.TechnologyStack = inferTechnologyStack(desc)
	nc.RequiredRuntime = inferRuntime(nc.TechnologyStack, nc.Component)

	// Extract affected versions from CPE match nodes
	versions := extractAffectedVersions(cve.Configurations)
	nc.AffectedVersions = versions

	// Determine actionability
	if !IsValidCVE(cve.ID) {
		nc.NonActionableReason = "invalid CVE identifier"
		return nc
	}
	if len(versions) == 0 {
		nc.NonActionableReason = "no affected versions found in NVD data"
		return nc
	}
	if nc.RequiredRuntime == "" {
		nc.NonActionableReason = "no identifiable runtime or service from CVE data"
		return nc
	}

	nc.Actionable = true
	return nc
}

// extractComponent attempts to identify the affected component from the CVE description.
func extractComponent(desc, cveID string) string {
	lower := strings.ToLower(desc)
	knownProducts := []string{
		"apache tomcat", "tomcat", "nginx", "apache httpd", "httpd",
		"mysql", "postgresql", "redis", "mongodb", "elasticsearch",
		"spring", "log4j", "struts", "docker", "kubernetes",
		"openssl", "gnutls", "mbedtls",
	}
	for _, p := range knownProducts {
		if strings.Contains(lower, p) {
			return p
		}
	}
	return ""
}

// inferTechnologyStack guesses the technology stack from the CVE description.
func inferTechnologyStack(desc string) string {
	lower := strings.ToLower(desc)
	if strings.Contains(lower, "tomcat") || strings.Contains(lower, "apache") {
		return "java"
	}
	if strings.Contains(lower, "nginx") {
		return "c"
	}
	if strings.Contains(lower, "mysql") || strings.Contains(lower, "postgresql") || strings.Contains(lower, "mongodb") || strings.Contains(lower, "redis") {
		return "database"
	}
	if strings.Contains(lower, "openssl") || strings.Contains(lower, "tls") || strings.Contains(lower, "ssl") {
		return "cryptography"
	}
	return "unknown"
}

// inferRuntime infers the required runtime from the technology stack and component.
func inferRuntime(stack, component string) string {
	switch stack {
	case "java":
		if component == "apache tomcat" || component == "tomcat" {
			return "tomcat"
		}
		return "java"
	case "c":
		if component == "nginx" {
			return "nginx"
		}
		return "c"
	case "database":
		return "database"
	case "cryptography":
		return "library"
	default:
		if component != "" {
			return component
		}
		return ""
	}
}

// extractAffectedVersions pulls version ranges from CPE match nodes.
func extractAffectedVersions(configs []struct {
	Nodes []struct {
		CpeMatch []struct {
			Vulnerable       bool   `json:"vulnerable"`
			Criteria         string `json:"criteria"`
			VersionStartIncl string `json:"versionStartIncluding,omitempty"`
			VersionEndIncl   string `json:"versionEndIncluding,omitempty"`
			VersionStartExcl string `json:"versionStartExcluding,omitempty"`
			VersionEndExcl   string `json:"versionEndExcluding,omitempty"`
		} `json:"cpeMatch,omitempty"`
	} `json:"nodes"`
}) []string {
	seen := make(map[string]bool)
	var versions []string

	for _, cfg := range configs {
		for _, node := range cfg.Nodes {
			for _, m := range node.CpeMatch {
				if !m.Vulnerable {
					continue
				}
				// Build a version range string
				parts := []string{}
				if m.VersionStartIncl != "" {
					parts = append(parts, ">="+m.VersionStartIncl)
				}
				if m.VersionStartExcl != "" {
					parts = append(parts, ">"+m.VersionStartExcl)
				}
				if m.VersionEndIncl != "" {
					parts = append(parts, "<="+m.VersionEndIncl)
				}
				if m.VersionEndExcl != "" {
					parts = append(parts, "<"+m.VersionEndExcl)
				}
				verStr := strings.Join(parts, " ")
				if verStr == "" {
					verStr = "any"
				}
				if !seen[verStr] {
					seen[verStr] = true
					versions = append(versions, verStr)
				}
			}
		}
	}
	return versions
}

// String returns a human-readable summary of the normalized CVE.
func (nc *NormalizedCVE) String() string {
	return fmt.Sprintf("CVE=%s product=%s runtime=%s stack=%s versions=%v actionable=%v",
		nc.CVEID, nc.AffectedProduct, nc.RequiredRuntime, nc.TechnologyStack, nc.AffectedVersions, nc.Actionable)
}
