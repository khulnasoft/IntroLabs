package provenance

// Source describes the provenance of a CVE or advisory source.
type Source struct {
	CVE      string `json:"cve,omitempty"`
	Advisory string `json:"advisory,omitempty"`
	Upstream string `json:"upstream,omitempty"`
}

// Evidence describes a piece of verification evidence.
type Evidence struct {
	Type        string `json:"type,omitempty"` // e.g., "schema-validation", "smoke-test", "runtime-health"
	Description string `json:"description,omitempty"`
	Passed      bool   `json:"passed"`
	Timestamp   string `json:"timestamp,omitempty"`
	Reference   string `json:"reference,omitempty"` // URL or file path to evidence artifact
}

// Checksum describes a SHA-256 checksum for a packaged artifact.
type Checksum struct {
	Algorithm string `json:"algorithm"` // e.g., "sha256"
	Value     string `json:"value"`
	File      string `json:"file,omitempty"` // path to the checksummed file
}

// Provenance aggregates source, evidence, and checksum metadata for a lab template.
type Provenance struct {
	Source     *Source    `json:"source,omitempty"`
	Evidence   []Evidence `json:"evidence,omitempty"`
	Checksum   *Checksum  `json:"checksum,omitempty"`
	Generated  string     `json:"generated,omitempty"` // ISO-8601 timestamp of generation
	Generator  string     `json:"generator,omitempty"` // generator version
	Reviewed   bool       `json:"reviewed,omitempty"`
	Reviewer   string     `json:"reviewer,omitempty"`
	ReviewedAt string     `json:"reviewedAt,omitempty"`
}
