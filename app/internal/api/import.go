package api

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/Kryvea/Kryvea/internal/burp"
	"github.com/Kryvea/Kryvea/internal/cvss"
	"github.com/Kryvea/Kryvea/internal/mongo"
	"github.com/Kryvea/Kryvea/internal/nessus"
	pocpkg "github.com/Kryvea/Kryvea/internal/poc"
	"github.com/Kryvea/Kryvea/internal/util"
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type importRequestData struct {
	Source string `json:"source"`
}

func (d *Driver) ImportVulnerabilities(c *fiber.Ctx) error {
	user := c.Locals("user").(*mongo.User)

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

	assessment, err := d.mongo.Assessment().GetByID(context.Background(), assessmentID)
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

	customer, err := d.mongo.Customer().GetByID(context.Background(), assessment.Customer.ID)
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
	case mongo.SourceBurp:
		parseErr = d.ParseBurp(data, *customer, *assessment, user.ID)
	case mongo.SourceNessus:
		parseErr = d.ParseNessus(data, *customer, *assessment, user.ID)
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

func (d *Driver) ParseBurp(data []byte, customer mongo.Customer, assessment mongo.Assessment, userID uuid.UUID) (err error) {
	burpData, err := burp.Parse(data)
	if err != nil {
		return err
	}

	session, err := d.mongo.NewSession()
	if err != nil {
		return err
	}
	defer session.End()

	_, err = session.WithTransaction(func(ctx context.Context) (any, error) {
		for _, issue := range burpData.Issues {
			target := &mongo.Target{
				IPv4: issue.Host.IP,
				FQDN: issue.Host.Name,
				Tag:  "burp",
			}
			targetID, _, err := d.mongo.Target().FirstOrInsert(ctx, target, customer.ID)
			if err != nil {
				return nil, err
			}

			err = d.mongo.Assessment().UpdateTargets(ctx, assessment.ID, targetID)
			if err != nil {
				return nil, err
			}

			category := &mongo.Category{
				Identifier:         strings.Trim(issue.Type, "\r\n "),
				Name:               strings.Trim(issue.Name, "\r\n "),
				Subcategory:        "",
				GenericDescription: map[string]string{"en": strings.Trim(issue.IssueBackground, "\r\n ")},
				GenericRemediation: map[string]string{"en": strings.Trim(issue.RemediationBackground, "\r\n ")},
				LanguagesOrder:     []string{"en"},
				References:         []string{},
				Source:             mongo.SourceBurp,
			}
			categoryID, _, err := d.mongo.Category().FirstOrInsert(ctx, category)
			if err != nil {
				return nil, err
			}

			vulnerability := &mongo.Vulnerability{
				Category: mongo.Category{
					Model: mongo.Model{
						ID: categoryID,
					},
				},
				CVSSv2:      cvss.InfoVector2,
				CVSSv3:      cvss.InfoVector3,
				CVSSv31:     cvss.InfoVector31,
				CVSSv4:      cvss.InfoVector4,
				Status:      strings.Trim(mongo.VulnerabilityStatusOpen, "\r\n "),
				References:  []string{strings.Trim(issue.References, "\r\n ")},
				Description: strings.Trim(issue.IssueDetail, "\r\n "),
				Remediation: strings.Trim(issue.RemediationDetail, "\r\n "),
				Target: mongo.Target{
					Model: mongo.Model{ID: targetID},
				},
				Assessment: mongo.Assessment{
					Model: mongo.Model{
						ID: assessment.ID,
					},
				},
				Customer: mongo.Customer{
					Model: mongo.Model{
						ID: customer.ID,
					},
				},
				User: mongo.User{
					Model: mongo.Model{
						ID: userID,
					},
				},
			}
			vulnerabilityID, err := d.mongo.Vulnerability().Insert(ctx, vulnerability)
			if err != nil {
				return nil, err
			}

			items := len(issue.RequestResponses) + len(issue.CollaboratorEvents) + len(issue.InfiltratorEvents)
			poc := mongo.Poc{
				VulnerabilityID: vulnerabilityID,
				Pocs:            make([]mongo.PocItem, 0, items),
			}
			i := 0
			for _, requestResponse := range issue.RequestResponses {
				var request, response []byte
				if requestResponse.Request != nil {
					request, err = base64.StdEncoding.DecodeString(requestResponse.Request.Base64)
					if err != nil {
						return nil, fmt.Errorf("cannot decode request: %w", err)
					}
				}
				if requestResponse.Response != nil {
					response, err = base64.StdEncoding.DecodeString(requestResponse.Response.Base64)
					if err != nil {
						return nil, fmt.Errorf("cannot decode response: %w", err)
					}
				}

				poc.Pocs = append(poc.Pocs, mongo.PocItem{
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
						request, err = base64.StdEncoding.DecodeString(collaboratorEvent.RequestResponse.Request.Base64)
						if err != nil {
							return nil, fmt.Errorf("cannot decode request: %w", err)
						}
					}
					if collaboratorEvent.RequestResponse.Response != nil {
						response, err = base64.StdEncoding.DecodeString(collaboratorEvent.RequestResponse.Response.Base64)
						if err != nil {
							return nil, fmt.Errorf("cannot decode response: %w", err)
						}
					}
				}

				poc.Pocs = append(poc.Pocs, mongo.PocItem{
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
				poc.Pocs = append(poc.Pocs, mongo.PocItem{
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

			err = d.mongo.Poc().Upsert(ctx, &poc)
			if err != nil {
				return nil, err
			}
		}

		return nil, nil
	})

	return err
}

func (d *Driver) ParseNessus(data []byte, customer mongo.Customer, assessment mongo.Assessment, userID uuid.UUID) (err error) {
	nessusData, err := nessus.Parse(data)
	if err != nil {
		return err
	}

	session, err := d.mongo.NewSession()
	if err != nil {
		return err
	}
	defer session.End()

	_, err = session.WithTransaction(func(ctx context.Context) (any, error) {
		if nessusData.Report == nil {
			return nil, errors.New("report data is empty")
		}

		for _, host := range nessusData.Report.ReportHosts {
			if host == nil {
				continue
			}
			var hostIP, hostFQDN, hostRDNS string
			if host.HostProperties == nil {
				continue
			}
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

			target := &mongo.Target{
				IPv4: hostIP,
				FQDN: hostFQDN,
				Tag:  "nessus",
			}

			targetID, _, err := d.mongo.Target().FirstOrInsert(ctx, target, customer.ID)
			if err != nil {
				return nil, err
			}

			err = d.mongo.Assessment().UpdateTargets(ctx, assessment.ID, targetID)
			if err != nil {
				return nil, err
			}

			for _, item := range host.ReportItems {
				if item == nil {
					continue
				}

				poc := mongo.Poc{
					Pocs: make([]mongo.PocItem, 0, 1),
				}
				category := &mongo.Category{
					Identifier:  strings.Trim(item.PluginID, "\r\n "),
					Name:        strings.Trim(item.PluginName, "\r\n "),
					Subcategory: "",
					GenericDescription: map[string]string{
						"en": strings.Trim(item.Description, "\r\n "),
					},
					GenericRemediation: map[string]string{
						"en": strings.Trim(item.Solution, "\r\n "),
					},
					LanguagesOrder: []string{"en"},
					References:     strings.Split(item.SeeAlso, "\n"),
					Source:         mongo.SourceNessus,
				}

				categoryID, _, err := d.mongo.Category().FirstOrInsert(ctx, category)
				if err != nil {
					return nil, err
				}

				vulnerability := &mongo.Vulnerability{
					Category: mongo.Category{
						Model: mongo.Model{
							ID: categoryID,
						},
					},
					CVSSv2:        cvss.InfoVector2,
					CVSSv3:        cvss.InfoVector3,
					CVSSv31:       cvss.InfoVector31,
					CVSSv4:        cvss.InfoVector4,
					DetailedTitle: "",
					Status:        mongo.VulnerabilityStatusOpen,
					References:    []string{},
					Description:   strings.Trim(item.Synopsis, "\r\n "),
					Remediation:   strings.Trim(item.Solution, "\r\n "),
					Target: mongo.Target{
						Model: mongo.Model{ID: targetID},
					},
					Assessment: mongo.Assessment{
						Model: mongo.Model{
							ID: assessment.ID,
						},
					},
					Customer: mongo.Customer{
						Model: mongo.Model{
							ID: customer.ID,
						},
					},
					User: mongo.User{
						Model: mongo.Model{
							ID: userID,
						},
					},
				}

				// Parse cvss2
				if item.CvssVector != "" {
					vector, err := cvss.ParseVector(item.CvssVector, cvss.Cvss2, assessment.Language)
					if err != nil {
						return nil, err
					}

					vulnerability.CVSSv2 = *vector
				}

				// Parse cvss3 as cvss31
				if item.Cvss3Vector != "" {
					vectorString := strings.Replace(item.Cvss3Vector, cvss.Cvss3, cvss.Cvss31, 1)
					vector, err := cvss.ParseVector(vectorString, cvss.Cvss31, assessment.Language)
					if err != nil {
						return nil, err
					}

					vulnerability.CVSSv31 = *vector
				}

				vulnerabilityID, err := d.mongo.Vulnerability().Insert(ctx, vulnerability)
				if err != nil {
					return nil, err
				}

				poc.Pocs = append(poc.Pocs, mongo.PocItem{
					Index:        0,
					Type:         "text",
					TextLanguage: "plaintext",
					TextData:     strings.Trim(item.PluginOutput, "\r\n "),
				})
				poc.VulnerabilityID = vulnerabilityID

				err = d.mongo.Poc().Upsert(ctx, &poc)
				if err != nil {
					return nil, err
				}
			}
		}

		return nil, nil
	})

	return err
}
