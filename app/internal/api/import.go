package api

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	xhtml "golang.org/x/net/html"

	"github.com/Kryvea/Kryvea/internal/burp"
	"github.com/Kryvea/Kryvea/internal/cvss"
	"github.com/Kryvea/Kryvea/internal/model"
	"github.com/Kryvea/Kryvea/internal/nessus"
	"github.com/Kryvea/Kryvea/internal/util"
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

var multiNewline = regexp.MustCompile(`\n{3,}`)
var whitespaceRun = regexp.MustCompile(`\s+`)
var nessusBullet = regexp.MustCompile(`^\s*[-*]\s+`)
var inlineSpaces = regexp.MustCompile(`[ \t]+`)

func dropNA(s string) string {
	if strings.EqualFold(strings.TrimSpace(s), "n/a") {
		return ""
	}
	return s
}

var burpSeverityVector = map[string]string{
	"Low":      "CVSS:3.1/AV:N/AC:H/PR:N/UI:N/S:U/C:L/I:N/A:N",
	"Medium":   "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:L/I:N/A:N",
	"High":     "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H",
	"Critical": "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:C/C:H/I:H/A:H",
}

type importRequestData struct {
	Source string `json:"source"`
}

func (d *Driver) ImportVulnerabilities(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.User)

	assessmentParam := c.Params("assessment")
	if assessmentParam == "" {
		return jsonError(c, fiber.StatusBadRequest, "Assessment ID is required")
	}

	assessmentID, err := util.ParseUUID(assessmentParam)
	if err != nil {
		return jsonError(c, fiber.StatusBadRequest, "Invalid assessment ID")
	}

	assessment, err := d.db.Assessment().GetByID(c.UserContext(), assessmentID)
	if err != nil {
		return jsonError(c, fiber.StatusBadRequest, "Invalid assessment ID")
	}

	if !user.CanAccessCustomer(assessment.Customer.ID) {
		return jsonError(c, fiber.StatusForbidden, "Forbidden")
	}

	// the assessment is fetched with its customer relation hydrated
	customer := assessment.Customer

	importData := &importRequestData{}
	err = sonic.Unmarshal([]byte(c.FormValue("import_data")), &importData)
	if err != nil {
		return jsonError(c, fiber.StatusBadRequest, "Cannot parse JSON")
	}

	data, _, err := d.formDataReadFile(c, "file")
	if err != nil {
		return jsonError(c, fiber.StatusBadRequest, "Cannot read file")
	}

	var parseErr error
	switch importData.Source {
	case model.SourceBurp:
		parseErr = d.parseBurp(c.UserContext(), data, customer, *assessment, user.ID)
	case model.SourceNessus:
		parseErr = d.parseNessus(c.UserContext(), data, customer, *assessment, user.ID)
	default:
		return jsonError(c, fiber.StatusBadRequest, "Unsupported source")
	}
	if parseErr != nil {
		return jsonError(c, fiber.StatusBadRequest, fmt.Sprintf("Cannot parse: %v", parseErr))
	}

	c.Status(fiber.StatusCreated)
	return c.JSON(fiber.Map{
		"message": "File parsed",
	})
}

// decodeBurpBody returns the raw body of a burp message part, decoding it
// from base64 when the export flags it as encoded.
func decodeBurpBody(content *burp.Base64Content, what string) ([]byte, error) {
	if content == nil {
		return nil, nil
	}
	if content.Base64 == "true" {
		decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(content.Body))
		if err != nil {
			return nil, fmt.Errorf("cannot decode %s: %w", what, err)
		}
		return decoded, nil
	}
	return []byte(content.Body), nil
}

func (d *Driver) parseBurp(ctx context.Context, data []byte, customer model.Customer, assessment model.Assessment, userID uuid.UUID) (err error) {
	burpData, err := burp.Parse(data)
	if err != nil {
		return err
	}

	_, err = d.db.RunInTx(ctx, func(ctx context.Context) (any, error) {
		targetCache := make(map[string]uuid.UUID)
		categoryCache := make(map[string]uuid.UUID)

		vulns := make([]*model.Vulnerability, 0, len(burpData.Issues))
		pocs := make([]model.Poc, 0, len(burpData.Issues))

		for _, issue := range burpData.Issues {
			var hostFQDN, protocol string
			var port int
			if u, err := url.Parse(issue.Host.Name); err == nil && u.Host != "" {
				if hostname := u.Hostname(); net.ParseIP(hostname) == nil {
					hostFQDN = hostname
				}
				if p, err := strconv.Atoi(u.Port()); err == nil {
					port = p
				}
				protocol = u.Scheme
			}
			target := &model.Target{
				FQDN:     dropNA(hostFQDN),
				Port:     port,
				Protocol: dropNA(protocol),
				Tag:      "burp",
			}
			if ip := net.ParseIP(dropNA(issue.Host.IP)); ip != nil {
				target.IPv4 = ""
				target.IPv6 = ip.String()
				if ip.To4() != nil {
					target.IPv4 = ip.String()
					target.IPv6 = ""
				}
			}

			// cache targets on the same fields FirstOrInsert matches on
			targetKey := fmt.Sprintf("%s|%s|%s|%d|%s", target.IPv4, target.IPv6, target.FQDN, target.Port, target.Protocol)
			targetID, ok := targetCache[targetKey]
			if !ok {
				targetID, _, err = d.db.Target().FirstOrInsert(ctx, target, customer.ID)
				if err != nil {
					return nil, err
				}
				targetCache[targetKey] = targetID
			}

			category := &model.Category{
				Identifier:         dropNA(strings.Trim(issue.Type, trimCutset)),
				Name:               dropNA(strings.Trim(issue.Name, trimCutset)),
				Subcategory:        "",
				GenericDescription: map[string]string{"en": dropNA(htmlToText(issue.IssueBackground))},
				GenericRemediation: map[string]string{"en": dropNA(htmlToText(issue.RemediationBackground))},
				LanguagesOrder:     []string{"en"},
				References:         htmlToRefs(issue.VulnerabilityClassifications),
				Source:             model.SourceBurp,
			}

			// cache categories on the same fields FirstOrInsert matches on
			catKey := fmt.Sprintf("%s|%s|%s", category.Identifier, category.Name, category.Subcategory)
			categoryID, ok := categoryCache[catKey]
			if !ok {
				categoryID, _, err = d.db.Category().FirstOrInsert(ctx, category)
				if err != nil {
					return nil, err
				}
				categoryCache[catKey] = categoryID
			}

			vulnerability := &model.Vulnerability{
				Category: model.Category{
					Model: model.Model{
						ID: categoryID,
					},
				},
				CVSSv2:      cvss.InfoVector2,
				CVSSv3:      cvss.InfoVector3,
				CVSSv31:     cvss.InfoVector31,
				CVSSv4:      cvss.InfoVector4,
				Status:      model.VulnerabilityStatusOpen,
				References:  htmlToRefs(issue.References),
				Description: dropNA(htmlToText(issue.IssueDetail)),
				Remediation: dropNA(htmlToText(issue.RemediationDetail)),
				GenericRemediation: model.VulnerabilityGeneric{
					Enabled: true,
				},
				Target: model.Target{
					Model: model.Model{ID: targetID},
				},
				Assessment: model.Assessment{
					Model: model.Model{
						ID: assessment.ID,
					},
				},
				Customer: model.Customer{
					Model: model.Model{
						ID: customer.ID,
					},
				},
				User: model.User{
					Model: model.Model{
						ID: userID,
					},
				},
			}
			if vectorStr, ok := burpSeverityVector[issue.Severity]; ok {
				vector, err := cvss.ParseVector(vectorStr, cvss.Cvss31, assessment.Language)
				if err != nil {
					return nil, err
				}
				vulnerability.CVSSv31 = *vector
			}

			items := len(issue.RequestResponses) + len(issue.CollaboratorEvents) + len(issue.InfiltratorEvents)
			poc := model.Poc{
				Pocs: make([]model.PocItem, 0, items),
			}
			i := 0
			for _, requestResponse := range issue.RequestResponses {
				request, err := decodeBurpBody(requestResponse.Request, "request")
				if err != nil {
					return nil, err
				}
				response, err := decodeBurpBody(requestResponse.Response, "response")
				if err != nil {
					return nil, err
				}

				poc.Pocs = append(poc.Pocs, model.PocItem{
					Index:    i,
					Type:     model.PocTypeRequest,
					Request:  strings.Trim(string(request), trimCutset),
					Response: strings.Trim(string(response), trimCutset),
				})

				i++
			}
			for _, collaboratorEvent := range issue.CollaboratorEvents {
				var request, response []byte
				if collaboratorEvent.RequestResponse != nil {
					request, err = decodeBurpBody(collaboratorEvent.RequestResponse.Request, "request")
					if err != nil {
						return nil, err
					}
					response, err = decodeBurpBody(collaboratorEvent.RequestResponse.Response, "response")
					if err != nil {
						return nil, err
					}
				}

				poc.Pocs = append(poc.Pocs, model.PocItem{
					Index: i,
					Type:  model.PocTypeText,
					TextData: fmt.Sprintf(`Interaction Type: %s
Origin IP: %s
Time: %s
Lookup Type: %s
Lookup Host: %s`,
						collaboratorEvent.InteractionType,
						collaboratorEvent.OriginIP,
						collaboratorEvent.Time,
						collaboratorEvent.LookupType,
						collaboratorEvent.LookupHost,
					),
					Request:  strings.Trim(string(request), trimCutset),
					Response: strings.Trim(string(response), trimCutset),
				})

				i++
			}
			for _, infiltratorEvent := range issue.InfiltratorEvents {
				poc.Pocs = append(poc.Pocs, model.PocItem{
					Index: i,
					Type:  model.PocTypeText,
					TextData: fmt.Sprintf(`Parameter Name: %s
Platform: %s
Signature: %s
Stack Trace: %s
Parameter Value: %s`,
						infiltratorEvent.ParameterName,
						infiltratorEvent.Platform,
						infiltratorEvent.Signature,
						infiltratorEvent.StackTrace,
						infiltratorEvent.ParameterValue,
					),
				})

				i++
			}

			vulns = append(vulns, vulnerability)
			pocs = append(pocs, poc)
		}

		if err := d.db.Vulnerability().BulkInsert(ctx, vulns); err != nil {
			return nil, err
		}

		for i := range pocs {
			pocs[i].VulnerabilityID = vulns[i].ID
		}

		if err := d.db.Poc().BulkInsertNew(ctx, pocs); err != nil {
			return nil, err
		}

		uniqueTargetIDs := make([]uuid.UUID, 0, len(targetCache))
		for _, id := range targetCache {
			uniqueTargetIDs = append(uniqueTargetIDs, id)
		}
		return nil, d.db.Assessment().BulkUpdateTargets(ctx, assessment.ID, uniqueTargetIDs)
	})

	return err
}

func (d *Driver) parseNessus(ctx context.Context, data []byte, customer model.Customer, assessment model.Assessment, userID uuid.UUID) (err error) {
	nessusData, err := nessus.Parse(data)
	if err != nil {
		return err
	}

	_, err = d.db.RunInTx(ctx, func(ctx context.Context) (any, error) {
		if nessusData.Report == nil {
			return nil, errors.New("report data is empty")
		}

		categoryCache := make(map[string]uuid.UUID)
		targetCache := make(map[string]uuid.UUID)

		totalItems := 0
		for _, host := range nessusData.Report.ReportHosts {
			if host != nil {
				totalItems += len(host.ReportItems)
			}
		}
		vulns := make([]*model.Vulnerability, 0, totalItems)
		pocs := make([]model.Poc, 0, totalItems)

		for _, host := range nessusData.Report.ReportHosts {
			if host == nil || host.HostProperties == nil {
				continue
			}

			var hostIP, hostFQDN, hostRDNS string
			for _, property := range host.HostProperties.Tag {
				switch property.Name {
				case "host-ip":
					hostIP = dropNA(property.Text)
				case "host-fqdn":
					hostFQDN = dropNA(property.Text)
				case "host-rdns":
					hostRDNS = dropNA(property.Text)
				}
			}
			if hostFQDN == hostRDNS {
				hostFQDN = ""
			}

			for _, item := range host.ReportItems {
				if item == nil {
					continue
				}

				itemProtocol := dropNA(item.Protocol)
				targetKey := fmt.Sprintf("%s|%s|%d|%s", hostIP, hostFQDN, item.Port, itemProtocol)
				targetID, ok := targetCache[targetKey]
				if !ok {
					target := &model.Target{
						IPv4:     hostIP,
						FQDN:     hostFQDN,
						Port:     item.Port,
						Protocol: itemProtocol,
						Tag:      "nessus",
					}
					targetID, _, err = d.db.Target().FirstOrInsert(ctx, target, customer.ID)
					if err != nil {
						return nil, err
					}
					targetCache[targetKey] = targetID
				}

				// cache categories on the same fields FirstOrInsert matches on
				// (identifier, name, subcategory - always empty for nessus)
				identifier := dropNA(strings.Trim(item.PluginID, trimCutset))
				name := dropNA(strings.Trim(item.PluginName, trimCutset))
				catKey := identifier + "|" + name + "|"
				categoryID, ok := categoryCache[catKey]
				if !ok {
					category := &model.Category{
						Identifier:         identifier,
						Name:               name,
						GenericDescription: map[string]string{"en": dropNA(nessusToText(item.Description))},
						GenericRemediation: map[string]string{"en": dropNA(nessusToText(item.Solution))},
						LanguagesOrder:     []string{"en"},
						References:         splitNessusRefs(item.SeeAlso),
						Source:             model.SourceNessus,
					}
					categoryID, _, err = d.db.Category().FirstOrInsert(ctx, category)
					if err != nil {
						return nil, err
					}
					categoryCache[catKey] = categoryID
				}

				vuln := &model.Vulnerability{
					Category:    model.Category{Model: model.Model{ID: categoryID}},
					CVSSv2:      cvss.InfoVector2,
					CVSSv3:      cvss.InfoVector3,
					CVSSv31:     cvss.InfoVector31,
					CVSSv4:      cvss.InfoVector4,
					Status:      model.VulnerabilityStatusOpen,
					References:  []string{},
					Description: dropNA(nessusToText(item.Synopsis)),
					GenericRemediation: model.VulnerabilityGeneric{
						Enabled: true,
					},
					Target:     model.Target{Model: model.Model{ID: targetID}},
					Assessment: model.Assessment{Model: model.Model{ID: assessment.ID}},
					Customer:   model.Customer{Model: model.Model{ID: customer.ID}},
					User:       model.User{Model: model.Model{ID: userID}},
				}

				if item.CvssVector != "" {
					vector, err := cvss.ParseVector(item.CvssVector, cvss.Cvss2, assessment.Language)
					if err != nil {
						return nil, err
					}
					vuln.CVSSv2 = *vector
				}
				if item.Cvss3Vector != "" {
					vectorString := strings.Replace(item.Cvss3Vector, cvss.Cvss3, cvss.Cvss31, 1)
					vector, err := cvss.ParseVector(vectorString, cvss.Cvss31, assessment.Language)
					if err != nil {
						return nil, err
					}
					vuln.CVSSv31 = *vector
				}

				pocs = append(pocs, model.Poc{
					Pocs: []model.PocItem{{
						Type:         "text",
						TextLanguage: "plaintext",
						TextData:     dropNA(strings.Trim(item.PluginOutput, trimCutset)),
					}},
				})
				vulns = append(vulns, vuln)
			}
		}

		if err := d.db.Vulnerability().BulkInsert(ctx, vulns); err != nil {
			return nil, err
		}

		for i := range pocs {
			pocs[i].VulnerabilityID = vulns[i].ID
		}

		if err := d.db.Poc().BulkInsertNew(ctx, pocs); err != nil {
			return nil, err
		}

		uniqueTargetIDs := make([]uuid.UUID, 0, len(targetCache))
		for _, id := range targetCache {
			uniqueTargetIDs = append(uniqueTargetIDs, id)
		}
		return nil, d.db.Assessment().BulkUpdateTargets(ctx, assessment.ID, uniqueTargetIDs)
	})

	return err
}

func htmlToText(s string) string {
	if s == "" {
		return ""
	}
	doc, err := xhtml.Parse(strings.NewReader(s))
	if err != nil {
		return s
	}
	var parts []string
	var current strings.Builder

	flush := func() {
		if text := strings.TrimSpace(current.String()); text != "" {
			parts = append(parts, text)
		}
		current.Reset()
	}

	var walk func(*xhtml.Node)
	walk = func(n *xhtml.Node) {
		if n.Type == xhtml.TextNode {
			current.WriteString(whitespaceRun.ReplaceAllString(n.Data, " "))
			return
		}
		if n.Type != xhtml.ElementNode {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				walk(c)
			}
			return
		}
		switch n.Data {
		case "p", "div", "h1", "h2", "h3", "h4", "h5", "h6":
			flush()
			parts = append(parts, "")
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				walk(c)
			}
			flush()
		case "li":
			flush()
			current.WriteString("• ")
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				walk(c)
			}
			flush()
		case "br":
			flush()
		case "ul", "ol":
			flush()
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				walk(c)
			}
			flush()
			parts = append(parts, "")
		default:
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				walk(c)
			}
		}
	}
	walk(doc)
	flush()
	return multiNewline.ReplaceAllString(strings.TrimSpace(strings.Join(parts, "\n")), "\n\n")
}

func htmlToRefs(s string) []string {
	if s == "" {
		return nil
	}
	doc, err := xhtml.Parse(strings.NewReader(s))
	if err != nil {
		return nil
	}
	var refs []string
	var walk func(*xhtml.Node)
	walk = func(n *xhtml.Node) {
		if n.Type == xhtml.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" && attr.Val != "" {
					refs = append(refs, attr.Val)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return refs
}

func nessusToText(s string) string {
	if s == "" {
		return ""
	}
	s = strings.ReplaceAll(s, "\r\n", "\n")
	var out []string
	for _, p := range strings.Split(s, "\n\n") {
		var blocks []string
		var current strings.Builder
		flush := func() {
			if t := strings.TrimSpace(inlineSpaces.ReplaceAllString(current.String(), " ")); t != "" {
				blocks = append(blocks, t)
			}
			current.Reset()
		}
		for _, line := range strings.Split(p, "\n") {
			if nessusBullet.MatchString(line) {
				flush()
				current.WriteString("• ")
				current.WriteString(strings.TrimSpace(nessusBullet.ReplaceAllString(line, "")))
			} else {
				if current.Len() > 0 {
					current.WriteString(" ")
				}
				current.WriteString(strings.TrimSpace(line))
			}
		}
		flush()
		if len(blocks) > 0 {
			out = append(out, strings.Join(blocks, "\n"))
		}
	}
	return multiNewline.ReplaceAllString(strings.TrimSpace(strings.Join(out, "\n\n")), "\n\n")
}

func splitNessusRefs(s string) []string {
	if s == "" {
		return []string{}
	}
	var refs []string
	for _, ref := range strings.Split(s, "\n") {
		if ref = strings.TrimSpace(ref); ref != "" {
			refs = append(refs, ref)
		}
	}
	return refs
}
