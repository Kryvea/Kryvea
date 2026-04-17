package burp

type BurpData struct {
	BurpVersion string  `xml:"burpVersion,attr"`
	ExportTime  string  `xml:"exportTime,attr"`
	Issues      []Issue `xml:"issue"`
}

type Issue struct {
	SerialNumber                 string              `xml:"serialNumber"`
	Type                         string              `xml:"type"`
	Name                         string              `xml:"name"`
	Host                         Host                `xml:"host"`
	Path                         string              `xml:"path"`
	Location                     string              `xml:"location"`
	Severity                     string              `xml:"severity"`
	Confidence                   string              `xml:"confidence"`
	IssueBackground              string              `xml:"issueBackground,omitempty"`
	RemediationBackground        string              `xml:"remediationBackground,omitempty"`
	References                   string              `xml:"references,omitempty"`
	VulnerabilityClassifications string              `xml:"vulnerabilityClassifications,omitempty"`
	IssueDetail                  string              `xml:"issueDetail,omitempty"`
	IssueDetailItems             *IssueDetailItems   `xml:"issueDetailItems,omitempty"`
	RemediationDetail            string              `xml:"remediationDetail,omitempty"`
	RequestResponses             []RequestResponse   `xml:"requestresponse,omitempty"`
	CollaboratorEvents           []CollaboratorEvent `xml:"collaboratorEvent,omitempty"`
	InfiltratorEvents            []InfiltratorEvent  `xml:"infiltratorEvent,omitempty"`
	StaticAnalysis               *StaticAnalysis     `xml:"staticAnalysis,omitempty"`
	DynamicAnalysis              *DynamicAnalysis    `xml:"dynamicAnalysis,omitempty"`
	PrototypePollution           *PrototypePollution `xml:"prototypePollution,omitempty"`
}

type Host struct {
	IP   string `xml:"ip,attr"`
	Name string `xml:",chardata"`
}

type IssueDetailItems struct {
	Items []string `xml:"issueDetailItem"`
}

type RequestResponse struct {
	Request            *Base64Content `xml:"request"`
	Response           *Base64Content `xml:"response"`
	ResponseRedirected string         `xml:"responseRedirected,omitempty"`
}

type Base64Content struct {
	Method string `xml:"method,attr,omitempty"`
	Base64 string `xml:"base64,attr,omitempty"`
	Body   string `xml:",chardata"`
}

type CollaboratorEvent struct {
	InteractionType string           `xml:"interactionType"`
	OriginIP        string           `xml:"originIp"`
	Time            string           `xml:"time"`
	LookupType      string           `xml:"lookupType,omitempty"`
	LookupHost      string           `xml:"lookupHost,omitempty"`
	RequestResponse *RequestResponse `xml:"requestresponse,omitempty"`
	SMTP            *SMTP            `xml:"smtp,omitempty"`
}

type SMTP struct {
	Sender       string     `xml:"sender"`
	Recipients   Recipients `xml:"recipients"`
	Message      string     `xml:"message"`
	Conversation string     `xml:"conversation"`
}

type Recipients struct {
	Recipient []string `xml:"recipient"`
}

type InfiltratorEvent struct {
	ParameterName     string            `xml:"parameterName"`
	Platform          string            `xml:"platform"`
	Signature         string            `xml:"signature"`
	StackTrace        string            `xml:"stackTrace,omitempty"`
	ParameterValue    string            `xml:"parameterValue,omitempty"`
	CollaboratorEvent CollaboratorEvent `xml:"collaboratorEvent"`
}

type StaticAnalysis struct {
	Source       string        `xml:"source"`
	Sink         string        `xml:"sink"`
	CodeSnippets *CodeSnippets `xml:"codeSnippets"`
}

type CodeSnippets struct {
	CodeSnippet []string `xml:"codeSnippet"`
}

type DynamicAnalysis struct {
	Source                      string `xml:"source"`
	Sink                        string `xml:"sink"`
	SourceStackTrace            string `xml:"sourceStackTrace"`
	SinkStackTrace              string `xml:"sinkStackTrace"`
	EventListenerStackTrace     string `xml:"eventListenerStackTrace"`
	SourceValue                 string `xml:"sourceValue"`
	SinkValue                   string `xml:"sinkValue"`
	EventHandlerData            string `xml:"eventHandlerData"`
	EventHandlerDataType        string `xml:"eventHandlerDataType"`
	EventHandlerManipulatedData string `xml:"eventHandlerManipulatedData"`
	POC                         string `xml:"poc"`
	Origin                      string `xml:"origin"`
	IsOriginChecked             string `xml:"isOriginChecked"`
	SourceElementId             string `xml:"sourceElementId"`
	SourceElementName           string `xml:"sourceElementName"`
	EventFiredEventName         string `xml:"eventFiredEventName"`
	EventFiredElementId         string `xml:"eventFiredElementId"`
	EventFiredElementName       string `xml:"eventFiredElementName"`
	EventFiredOuterHtml         string `xml:"eventFiredOuterHtml"`
}

type PrototypePollution struct {
	POC                string `xml:"poc"`
	PollutionTechnique string `xml:"pollutionTechnique"`
	PollutionType      string `xml:"pollutionType"`
}
