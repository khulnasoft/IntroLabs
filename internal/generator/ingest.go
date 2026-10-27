package generator

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// cvePattern validates CVE identifiers (CVE-YYYY-NNNN or longer).
var cvePattern = regexp.MustCompile(`^CVE-\d{4}-\d{4,}$`)

// IsValidCVE reports whether id is a well-formed CVE identifier.
func IsValidCVE(id string) bool {
	return cvePattern.MatchString(strings.TrimSpace(id))
}

// NVDCVEItem represents a single CVE item from the NVD 2.0 API.
type NVDCVEItem struct {
	ID           string    `json:"id"`
	Published    time.Time `json:"published"`
	LastModified time.Time `json:"lastModified"`
	VulnStatus   string    `json:"vulnStatus"`
	Descriptions []struct {
		Lang  string `json:"lang"`
		Value string `json:"value"`
	} `json:"descriptions"`
	Metrics struct {
		CVSSMetricV31 []struct {
			CvssData struct {
				Version             string  `json:"version"`
				VectorString        string  `json:"vectorString"`
				BaseScore           float64 `json:"baseScore"`
				BaseSeverity        string  `json:"baseSeverity"`
				ExploitabilityScore float64 `json:"exploitabilityScore"`
				ImpactScore         float64 `json:"impactScore"`
			} `json:"cvssData"`
		} `json:"cvssMetricV31"`
	} `json:"metrics"`
	Configurations []struct {
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
	} `json:"configurations"`
	References []struct {
		URL    string   `json:"url"`
		Source string   `json:"source"`
		Tags   []string `json:"tags"`
	} `json:"references"`
}

// NVDResponse represents the top-level NVD 2.0 API response.
type NVDResponse struct {
	ResultsPerPage  int `json:"resultsPerPage"`
	TotalResults    int `json:"totalResults"`
	Vulnerabilities []struct {
		CVE NVDCVEItem `json:"cve"`
	} `json:"vulnerabilities"`
}

// Ingestor fetches CVE metadata from the NVD API.
type Ingestor struct {
	client  *http.Client
	baseURL string
}

// NewIngestor creates an Ingestor with a default HTTP client.
func NewIngestor() *Ingestor {
	return &Ingestor{
		client:  &http.Client{Timeout: 30 * time.Second},
		baseURL: "https://services.nvd.nist.gov/rest/json/cves/2.0",
	}
}

// NewIngestorWithClient creates an Ingestor with a custom HTTP client (useful for tests).
func NewIngestorWithClient(client *http.Client, baseURL string) *Ingestor {
	return &Ingestor{
		client:  client,
		baseURL: baseURL,
	}
}

// Fetch retrieves CVE metadata from the NVD API for the given CVE ID.
func (i *Ingestor) Fetch(cveID string) (*NVDCVEItem, error) {
	if !IsValidCVE(cveID) {
		return nil, fmt.Errorf("invalid CVE identifier: %s", cveID)
	}

	reqURL := i.baseURL + "?cveId=" + url.QueryEscape(cveID)
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := i.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching CVE data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("CVE %s not found in NVD", cveID)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("NVD API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	var nvdResp NVDResponse
	if err := json.Unmarshal(body, &nvdResp); err != nil {
		return nil, fmt.Errorf("parsing NVD response: %w", err)
	}

	if nvdResp.TotalResults == 0 || len(nvdResp.Vulnerabilities) == 0 {
		return nil, fmt.Errorf("CVE %s not found in NVD", cveID)
	}

	return &nvdResp.Vulnerabilities[0].CVE, nil
}

// FetchFromReader reads CVE data from a reader (useful for offline tests with fixtures).
func FetchFromReader(r io.Reader) (*NVDCVEItem, error) {
	body, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading fixture: %w", err)
	}

	var nvdResp NVDResponse
	if err := json.Unmarshal(body, &nvdResp); err != nil {
		return nil, fmt.Errorf("parsing fixture: %w", err)
	}

	if nvdResp.TotalResults == 0 || len(nvdResp.Vulnerabilities) == 0 {
		return nil, fmt.Errorf("no CVE data in fixture")
	}

	return &nvdResp.Vulnerabilities[0].CVE, nil
}

// Description returns the English description of the CVE, or empty string if none.
func (c *NVDCVEItem) Description() string {
	for _, d := range c.Descriptions {
		if d.Lang == "en" {
			return d.Value
		}
	}
	if len(c.Descriptions) > 0 {
		return c.Descriptions[0].Value
	}
	return ""
}

// CVSSv31 returns the CVSS v3.1 vector string and base score, or empty/0 if not available.
func (c *NVDCVEItem) CVSSv31() (string, float64) {
	if len(c.Metrics.CVSSMetricV31) > 0 {
		return c.Metrics.CVSSMetricV31[0].CvssData.VectorString, c.Metrics.CVSSMetricV31[0].CvssData.BaseScore
	}
	return "", 0
}
