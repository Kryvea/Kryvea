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
	pocpkg "github.com/Kryvea/Kryvea/internal/poc"
	"github.com/Kryvea/Kryvea/internal/util"
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

var multiNewline = regexp.MustCompile(`\n{3,}`)
var whitespaceRun = regexp.MustCompile(`\s+`)

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
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "Assessment ID is required",
		})
	}

	assessmentID, err := util.ParseUUID(assessmentParam)
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "Invalid assessment ID",
		})
	}

	assessment, err := d.db.Assessment().GetByID(c.UserContext(), assessmentID)
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "Invalid assessment ID",
		})
	}

	if !user.CanAccessCustomer(assessment.Customer.ID) {
		c.Status(fiber.StatusForbidden)
		return c.JSON(fiber.Map{
			"error": "Forbidden",
		})
	}

	customer, err := d.db.Customer().GetByID(c.UserContext(), assessment.Customer.ID)
	if err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "Cannot get customer",
		})
	}

	// parse request body
	importData := &importRequestData{}
	err = sonic.Unmarshal([]byte(c.FormValue("import_data")), &importData)
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	data, _, err := d.formDataReadFile(c, "file")
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "Cannot read file",
		})
	}

	var parseErr error
	switch importData.Source {
	case model.SourceBurp:
		parseErr = d.ParseBurp(c.UserContext(), data, *customer, *assessment, user.ID)
	case model.SourceNessus:
		parseErr = d.ParseNessus(c.UserContext(), data, *customer, *assessment, user.ID)
	default:
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "Unsupported source",
		})
	}
	if parseErr != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": fmt.Sprintf("Cannot parse: %v", parseErr),
		})
	}

	c.Status(fiber.StatusCreated)
	return c.JSON(fiber.Map{
		"message": "File parsed",
	})
}

func (d *Driver) ParseBurp(ctx context.Context, data []byte, customer model.Customer, assessment model.Assessment, userID uuid.UUID) (err error) {
	burpData, err := burp.Parse(data)
	if err != nil {
		return err
	}

	_, err = d.db.RunInTx(ctx, func(ctx context.Context) (any, error) {
		for _, issue := range burpData.Issues {
			var hostFQDN string
			var port int
			if u, err := url.Parse(issue.Host.Name); err == nil && u.Host != "" {
				if hostname := u.Hostname(); net.ParseIP(hostname) == nil {
					hostFQDN = hostname
				}
				if p, err := strconv.Atoi(u.Port()); err == nil {
					port = p
				}
			}
			target := &model.Target{
				FQDN: hostFQDN,
				Port: port,
				Tag:  "burp",
			}
			if ip := net.ParseIP(issue.Host.IP); ip != nil {
				if ip.To4() != nil {
					target.IPv4 = issue.Host.IP
				} else {
					target.IPv6 = issue.Host.IP
				}
			}
			targetID, _, err := d.db.Target().FirstOrInsert(ctx, target, customer.ID)
			if err != nil {
				return nil, err
			}

			err = d.db.Assessment().UpdateTargets(ctx, assessment.ID, targetID)
			if err != nil {
				return nil, err
			}

			category := &model.Category{
				Identifier:         strings.Trim(issue.Type, "\r\n "),
				Name:               strings.Trim(issue.Name, "\r\n "),
				Subcategory:        "",
				GenericDescription: map[string]string{"en": htmlToText(issue.IssueBackground)},
				GenericRemediation: map[string]string{"en": htmlToText(issue.RemediationBackground)},
				LanguagesOrder:     []string{"en"},
				References:         htmlToRefs(issue.VulnerabilityClassifications),
				Source:             model.SourceBurp,
			}
			categoryID, _, err := d.db.Category().FirstOrInsert(ctx, category)
			if err != nil {
				return nil, err
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
				Status:      strings.Trim(model.VulnerabilityStatusOpen, "\r\n "),
				References:  htmlToRefs(issue.References),
				Description: htmlToText(issue.IssueDetail),
				Remediation: htmlToText(issue.RemediationDetail),
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
			vulnerabilityID, err := d.db.Vulnerability().Insert(ctx, vulnerability)
			if err != nil {
				return nil, err
			}

			items := len(issue.RequestResponses) + len(issue.CollaboratorEvents) + len(issue.InfiltratorEvents)
			poc := model.Poc{
				VulnerabilityID: vulnerabilityID,
				Pocs:            make([]model.PocItem, 0, items),
			}
			i := 0
			for _, requestResponse := range issue.RequestResponses {
				var request, response []byte
				if requestResponse.Request != nil {
					if requestResponse.Request.Base64 == "true" {
						request, err = base64.StdEncoding.DecodeString(strings.TrimSpace(requestResponse.Request.Body))
						if err != nil {
							return nil, fmt.Errorf("cannot decode request: %w", err)
						}
					} else {
						request = []byte(requestResponse.Request.Body)
					}
				}
				if requestResponse.Response != nil {
					if requestResponse.Response.Base64 == "true" {
						response, err = base64.StdEncoding.DecodeString(strings.TrimSpace(requestResponse.Response.Body))
						if err != nil {
							return nil, fmt.Errorf("cannot decode response: %w", err)
						}
					} else {
						response = []byte(requestResponse.Response.Body)
					}
				}

				poc.Pocs = append(poc.Pocs, model.PocItem{
					Index:    i,
					Type:     pocpkg.PocTypeRequest,
					Request:  strings.Trim(string(request), "\r\n "),
					Response: strings.Trim(string(response), "\r\n "),
				})

				i++
			}
			for _, collaboratorEvent := range issue.CollaboratorEvents {
				var request, response []byte
				if collaboratorEvent.RequestResponse != nil {
					if collaboratorEvent.RequestResponse.Request != nil {
						if collaboratorEvent.RequestResponse.Request.Base64 == "true" {
							request, err = base64.StdEncoding.DecodeString(strings.TrimSpace(collaboratorEvent.RequestResponse.Request.Body))
							if err != nil {
								return nil, fmt.Errorf("cannot decode request: %w", err)
							}
						} else {
							request = []byte(collaboratorEvent.RequestResponse.Request.Body)
						}
					}
					if collaboratorEvent.RequestResponse.Response != nil {
						if collaboratorEvent.RequestResponse.Response.Base64 == "true" {
							response, err = base64.StdEncoding.DecodeString(strings.TrimSpace(collaboratorEvent.RequestResponse.Response.Body))
							if err != nil {
								return nil, fmt.Errorf("cannot decode response: %w", err)
							}
						} else {
							response = []byte(collaboratorEvent.RequestResponse.Response.Body)
						}
					}
				}

				poc.Pocs = append(poc.Pocs, model.PocItem{
					Index: i,
					Type:  pocpkg.PocTypeText,
					TextData: strings.Trim(fmt.Sprintf(`Interaction Type: %s
Origin IP: %s
Time: %s
Lookup Type: %s
Lookup Host: %s`,
						collaboratorEvent.InteractionType,
						collaboratorEvent.OriginIP,
						collaboratorEvent.Time,
						collaboratorEvent.LookupType,
						collaboratorEvent.LookupHost,
					), "\r\n "),
					Request:  strings.Trim(string(request), "\r\n "),
					Response: strings.Trim(string(response), "\r\n "),
				})

				i++
			}
			for _, infiltratorEvent := range issue.InfiltratorEvents {
				poc.Pocs = append(poc.Pocs, model.PocItem{
					Index: i,
					Type:  pocpkg.PocTypeText,
					TextData: strings.Trim(fmt.Sprintf(`Parameter Name: %s
Platform: %s
Signature: %s
Stack Trace: %s
Parameter Value: %s`,
						infiltratorEvent.ParameterName,
						infiltratorEvent.Platform,
						infiltratorEvent.Signature,
						infiltratorEvent.StackTrace,
						infiltratorEvent.ParameterValue,
					), "\r\n "),
				})

				i++
			}

			err = d.db.Poc().Upsert(ctx, &poc)
			if err != nil {
				return nil, err
			}
		}

		return nil, nil
	})

	return err
}

func (d *Driver) ParseNessus(ctx context.Context, data []byte, customer model.Customer, assessment model.Assessment, userID uuid.UUID) (err error) {
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
					hostIP = property.Text
				case "host-fqdn":
					hostFQDN = property.Text
				case "host-rdns":
					hostRDNS = property.Text
				}
			}
			if hostFQDN == hostRDNS {
				hostFQDN = ""
			}

			targetKey := hostIP + "|" + hostFQDN
			targetID, ok := targetCache[targetKey]
			if !ok {
				target := &model.Target{IPv4: hostIP, FQDN: hostFQDN, Tag: "nessus"}
				targetID, _, err = d.db.Target().FirstOrInsert(ctx, target, customer.ID)
				if err != nil {
					return nil, err
				}
				targetCache[targetKey] = targetID
			}

			for _, item := range host.ReportItems {
				if item == nil {
					continue
				}

				catKey := item.PluginID
				categoryID, ok := categoryCache[catKey]
				if !ok {
					category := &model.Category{
						Identifier:         strings.Trim(item.PluginID, "\r\n "),
						Name:               strings.Trim(item.PluginName, "\r\n "),
						GenericDescription: map[string]string{"en": strings.Trim(item.Description, "\r\n ")},
						GenericRemediation: map[string]string{"en": strings.Trim(item.Solution, "\r\n ")},
						LanguagesOrder:     []string{"en"},
						References:         strings.Split(item.SeeAlso, "\n"),
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
					Description: strings.Trim(item.Synopsis, "\r\n "),
					Remediation: strings.Trim(item.Solution, "\r\n "),
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
						TextData:     strings.Trim(item.PluginOutput, "\r\n "),
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
