package nessus

type NessusData struct {
	Policy *Policy `xml:"Policy,omitempty"`
	Report *Report `xml:"Report,omitempty"`
}

type Policy struct {
	FamilySelection           *FamilySelection           `xml:"FamilySelection,omitempty"`
	IndividualPluginSelection *IndividualPluginSelection `xml:"IndividualPluginSelection,omitempty"`
	PolicyName                string                     `xml:"policyName,omitempty"`
	Preferences               *Preferences               `xml:"Preferences,omitempty"`
}

type Preferences struct {
	PluginsPreferences *PluginsPreferences `xml:"PluginsPreferences,omitempty"`
	ServerPreferences  *ServerPreferences  `xml:"ServerPreferences,omitempty"`
}

type ServerPreferences struct {
	Preference []struct {
		Name  string `xml:"name,omitempty"`
		Value string `xml:"value,omitempty"`
	} `xml:"preference,omitempty"`
}

type PluginsPreferences struct {
	Item []struct {
		FullName         string `xml:"fullName,omitempty"`
		PluginID         string `xml:"pluginId,omitempty"`
		PluginName       string `xml:"pluginName,omitempty"`
		PreferenceName   string `xml:"preferenceName,omitempty"`
		PreferenceType   string `xml:"preferenceType,omitempty"`
		PreferenceValues string `xml:"preferenceValues,omitempty"`
		SelectedValue    string `xml:"selectedValue,omitempty"`
	} `xml:"item,omitempty"`
}

type FamilySelection struct {
	FamilyItem []struct {
		FamilyName string `xml:"FamilyName,omitempty"`
		Status     string `xml:"Status,omitempty"`
	} `xml:"FamilyItem,omitempty"`
}

type IndividualPluginSelection struct {
	PluginItem []struct {
		Family     string `xml:"Family,omitempty"`
		PluginID   string `xml:"PluginId,omitempty"`
		PluginName string `xml:"PluginName,omitempty"`
		Status     string `xml:"Status,omitempty"`
	} `xml:"PluginItem,omitempty"`
}

type Report struct {
	Xmlnscm     string        `xml:"xmlns cm,attr"`
	Name        string        `xml:"name,attr"`
	ReportHosts []*ReportHost `xml:"ReportHost,omitempty"`
}

type ReportHost struct {
	Name           string          `xml:"name,attr"`
	HostProperties *HostProperties `xml:"HostProperties,omitempty"`
	ReportItems    []*ReportItem   `xml:"ReportItem,omitempty"`
}

type HostProperties struct {
	Tag []struct {
		Text string `xml:",chardata"`
		Name string `xml:"name,attr"`
	} `xml:"tag,omitempty"`
}

type ReportItem struct {
	Port                              int      `xml:"port,attr"`
	SvcName                           string   `xml:"svc_name,attr"`
	Protocol                          string   `xml:"protocol,attr"`
	Severity                          int      `xml:"severity,attr"`
	PluginID                          string   `xml:"pluginID,attr"`
	AttrPluginName                    string   `xml:"pluginName,attr"`
	PluginFamily                      string   `xml:"pluginFamily,attr"`
	Description                       string   `xml:"description,omitempty"`
	Fname                             string   `xml:"fname,omitempty"`
	PluginModificationDate            string   `xml:"plugin_modification_date,omitempty"`
	PluginName                        string   `xml:"plugin_name,omitempty"`
	PluginPublicationDate             string   `xml:"plugin_publication_date,omitempty"`
	PluginType                        string   `xml:"plugin_type,omitempty"`
	RiskFactor                        string   `xml:"risk_factor,omitempty"`
	ScriptVersion                     string   `xml:"script_version,omitempty"`
	Solution                          string   `xml:"solution,omitempty"`
	Synopsis                          string   `xml:"synopsis,omitempty"`
	PluginOutput                      string   `xml:"plugin_output,omitempty"`
	Agent                             string   `xml:"agent,omitempty"`
	SeeAlso                           string   `xml:"see_also,omitempty"`
	CvssBaseScore                     float64  `xml:"cvss_base_score,omitempty"`
	CvssScoreRationale                string   `xml:"cvss_score_rationale,omitempty"`
	CvssScoreSource                   string   `xml:"cvss_score_source,omitempty"`
	CvssVector                        string   `xml:"cvss_vector,omitempty"`
	ExploitedByNessus                 string   `xml:"exploited_by_nessus,omitempty"`
	AssetInventory                    string   `xml:"asset_inventory,omitempty"`
	Iavt                              string   `xml:"iavt,omitempty"`
	XREF                              []string `xml:"xref,omitempty"`
	AlwaysRun                         string   `xml:"always_run,omitempty"`
	Iavb                              string   `xml:"iavb,omitempty"`
	ThoroughTests                     string   `xml:"thorough_tests,omitempty"`
	Cpe                               string   `xml:"cpe,omitempty"`
	Iava                              []string `xml:"iava,omitempty"`
	Compliance                        bool     `xml:"compliance,omitempty"`
	ComplianceCheckType               string   `xml:"compliance_check_type,omitempty"`
	ComplianceSupportsParseValidation bool     `xml:"compliance_supports_parse_validation,omitempty"`
	ComplianceSupportsReplacement     bool     `xml:"compliance_supports_replacement,omitempty"`
	ComplianceBenchmarkVersion        string   `xml:"compliance-benchmark-version,omitempty"`
	ComplianceCheckName               string   `xml:"compliance-check-name,omitempty"`
	ComplianceCheckID                 string   `xml:"compliance-check-id,omitempty"`
	ComplianceActualValue             string   `xml:"compliance-actual-value,omitempty"`
	ComplianceSource                  string   `xml:"compliance-source,omitempty"`
	ComplianceAuditFile               string   `xml:"compliance-audit-file,omitempty"`
	CompliancePolicyValue             string   `xml:"compliance-policy-value,omitempty"`
	ComplianceFunctionalID            string   `xml:"compliance-functional-id,omitempty"`
	ComplianceUname                   string   `xml:"compliance-uname,omitempty"`
	ComplianceInfo                    string   `xml:"compliance-info,omitempty"`
	ComplianceResult                  string   `xml:"compliance-result,omitempty"`
	ComplianceInformationalID         string   `xml:"compliance-informational-id,omitempty"`
	ComplianceReference               string   `xml:"compliance-reference,omitempty"`
	ComplianceSolution                string   `xml:"compliance-solution,omitempty"`
	ComplianceBenchmarkName           string   `xml:"compliance-benchmark-name,omitempty"`
	ComplianceControlID               string   `xml:"compliance-control-id,omitempty"`
	ComplianceSeeAlso                 string   `xml:"compliance-see-also,omitempty"`
	ComplianceFullID                  string   `xml:"compliance-full-id,omitempty"`
	ComplianceError                   string   `xml:"compliance-error,omitempty"`
	AgeOfVuln                         string   `xml:"age_of_vuln,omitempty"`
	CVE                               []string `xml:"cve,omitempty"`
	BID                               []string `xml:"bid,omitempty"`
	Cvss3BaseScore                    float64  `xml:"cvss3_base_score,omitempty"`
	Cvss3TemporalScore                float64  `xml:"cvss3_temporal_score,omitempty"`
	Cvss3TemporalVector               string   `xml:"cvss3_temporal_vector,omitempty"`
	Cvss3Vector                       string   `xml:"cvss3_vector,omitempty"`
	CvssV3ImpactScore                 float64  `xml:"cvssV3_impactScore,omitempty"`
	CvssTemporalScore                 float64  `xml:"cvss_temporal_score,omitempty"`
	CvssTemporalVector                string   `xml:"cvss_temporal_vector,omitempty"`
	CWE                               []string `xml:"cwe,omitempty"`
	ExploitAvailable                  bool     `xml:"exploit_available,omitempty"`
	ExploitFrameworkCanvas            bool     `xml:"exploit_framework_canvas,omitempty"`
	ExploitFrameworkMetasploit        bool     `xml:"exploit_framework_metasploit,omitempty"`
	ExploitFrameworkCore              bool     `xml:"exploit_framework_core,omitempty"`
	ExploitCodeMaturity               string   `xml:"exploit_code_maturity,omitempty"`
	ExploitabilityEase                string   `xml:"exploitability_ease,omitempty"`
	PatchPublicationDate              string   `xml:"patch_publication_date,omitempty"`
	ProductCoverage                   string   `xml:"product_coverage,omitempty"`
	StigSeverity                      string   `xml:"stig_severity,omitempty"`
	ThreatIntensityLast28             string   `xml:"threat_intensity_last_28,omitempty"`
	ThreatRecency                     string   `xml:"threat_recency,omitempty"`
	ThreatSourcesLast28               string   `xml:"threat_sources_last_28,omitempty"`
	VprScore                          string   `xml:"vpr_score,omitempty"`
	VulnPublicationDate               string   `xml:"vuln_publication_date,omitempty"`
	CisaKnownExploited                string   `xml:"cisa-known-exploited,omitempty"`
	Cvss3ScoreSource                  string   `xml:"cvss3_score_source,omitempty"`
	AssetCategories                   string   `xml:"asset_categories,omitempty"`
	HardwareInventory                 string   `xml:"hardware_inventory,omitempty"`
	OsIdentification                  string   `xml:"os_identification,omitempty"`
}
