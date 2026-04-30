package db

import (
	"time"

	"github.com/Kryvea/Kryvea/internal/crypto"
	"github.com/Kryvea/Kryvea/internal/model"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type dbSetting struct {
	bun.BaseModel `bun:"table:setting,alias:s"`

	ID                      uuid.UUID `bun:"id,pk,type:uuid"`
	CreatedAt               time.Time `bun:"created_at,nullzero,notnull,default:now()"`
	UpdatedAt               time.Time `bun:"updated_at,nullzero,notnull,default:now()"`
	MaxImageSize            int64     `bun:"max_image_size,notnull,default:5242880"`
	DefaultCategoryLanguage string    `bun:"default_category_language,notnull,default:'en'"`
}

func (r *dbSetting) toModel() model.Setting {
	out := model.Setting{
		MaxImageSize:            r.MaxImageSize,
		DefaultCategoryLanguage: r.DefaultCategoryLanguage,
	}
	out.ID = r.ID
	out.CreatedAt = r.CreatedAt
	out.UpdatedAt = r.UpdatedAt
	return out
}

type dbFileReference struct {
	bun.BaseModel `bun:"table:file_reference,alias:fr"`

	ID        uuid.UUID `bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:now()"`
	UpdatedAt time.Time `bun:"updated_at,nullzero,notnull,default:now()"`
	Checksum  []byte    `bun:"checksum,notnull,unique"`
	MimeType  string    `bun:"mime_type,notnull"`
	SizeBytes int64     `bun:"size_bytes,notnull,default:0"`
}

func (r *dbFileReference) toModel() model.FileReference {
	out := model.FileReference{
		MimeType: r.MimeType,
		UsedBy:   []uuid.UUID{},
	}
	out.ID = r.ID
	out.CreatedAt = r.CreatedAt
	out.UpdatedAt = r.UpdatedAt
	if len(r.Checksum) == 16 {
		copy(out.Checksum[:], r.Checksum)
	}
	return out
}

type dbCustomer struct {
	bun.BaseModel `bun:"table:customer,alias:c"`

	ID            uuid.UUID  `bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	CreatedAt     time.Time  `bun:"created_at,nullzero,notnull,default:now()"`
	UpdatedAt     time.Time  `bun:"updated_at,nullzero,notnull,default:now()"`
	Name          string     `bun:"name,notnull,unique"`
	Language      string     `bun:"language,notnull,default:''"`
	LogoID        *uuid.UUID `bun:"logo_id,nullzero,type:uuid"`
	LogoMimeType  string     `bun:"logo_mime_type,notnull,default:''"`
	LogoReference string     `bun:"logo_reference,notnull,default:''"`

	Templates []dbTemplate `bun:"rel:has-many,join:id=customer_id"`
}

func (r *dbCustomer) toModel() model.Customer {
	out := model.Customer{
		Name:          r.Name,
		Language:      r.Language,
		LogoMimeType:  r.LogoMimeType,
		LogoReference: r.LogoReference,
		Templates:     []model.Template{},
	}
	out.ID = r.ID
	out.CreatedAt = r.CreatedAt
	out.UpdatedAt = r.UpdatedAt
	if r.LogoID != nil {
		out.LogoID = *r.LogoID
	}
	return out
}

type dbUser struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	ID             uuid.UUID  `bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	CreatedAt      time.Time  `bun:"created_at,nullzero,notnull,default:now()"`
	UpdatedAt      time.Time  `bun:"updated_at,nullzero,notnull,default:now()"`
	DisabledAt     *time.Time `bun:"disabled_at,nullzero"`
	Username       string     `bun:"username,notnull,unique"`
	Password       []byte     `bun:"password,notnull"`
	PasswordExpiry time.Time  `bun:"password_expiry,notnull"`
	Token          []byte     `bun:"token,nullzero"`
	TokenExpiry    *time.Time `bun:"token_expiry,nullzero"`
	Role           string     `bun:"role,notnull"`

	Customers   []dbCustomer   `bun:"m2m:user_customer,join:User=Customer"`
	Assessments []dbAssessment `bun:"m2m:user_assessment,join:User=Assessment"`
}

func (r *dbUser) toModel() model.User {
	out := model.User{
		Username:       r.Username,
		Password:       r.Password,
		PasswordExpiry: r.PasswordExpiry,
		Role:           r.Role,
	}
	out.ID = r.ID
	out.CreatedAt = r.CreatedAt
	out.UpdatedAt = r.UpdatedAt
	if r.DisabledAt != nil {
		out.DisabledAt = *r.DisabledAt
	}
	if len(r.Token) > 0 {
		out.Token = crypto.Token(r.Token)
	}
	if r.TokenExpiry != nil {
		out.TokenExpiry = *r.TokenExpiry
	}
	return out
}

type dbUserCustomer struct {
	bun.BaseModel `bun:"table:user_customer,alias:uc"`

	UserID     uuid.UUID `bun:"user_id,pk,type:uuid"`
	CustomerID uuid.UUID `bun:"customer_id,pk,type:uuid"`

	User     *dbUser     `bun:"rel:belongs-to,join:user_id=id"`
	Customer *dbCustomer `bun:"rel:belongs-to,join:customer_id=id"`
}

type dbUserAssessment struct {
	bun.BaseModel `bun:"table:user_assessment,alias:ua"`

	UserID       uuid.UUID `bun:"user_id,pk,type:uuid"`
	AssessmentID uuid.UUID `bun:"assessment_id,pk,type:uuid"`

	User       *dbUser       `bun:"rel:belongs-to,join:user_id=id"`
	Assessment *dbAssessment `bun:"rel:belongs-to,join:assessment_id=id"`
}

type dbCategory struct {
	bun.BaseModel `bun:"table:category,alias:cat"`

	ID                 uuid.UUID         `bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	CreatedAt          time.Time         `bun:"created_at,nullzero,notnull,default:now()"`
	UpdatedAt          time.Time         `bun:"updated_at,nullzero,notnull,default:now()"`
	Identifier         string            `bun:"identifier,notnull,default:'',unique:category_natural_key"`
	Name               string            `bun:"name,notnull,default:'',unique:category_natural_key"`
	Subcategory        string            `bun:"subcategory,notnull,default:'',unique:category_natural_key"`
	GenericDescription map[string]string `bun:"generic_description,type:jsonb,notnull,default:'{}'"`
	GenericRemediation map[string]string `bun:"generic_remediation,type:jsonb,notnull,default:'{}'"`
	LanguagesOrder     []string          `bun:"languages_order,array,notnull,default:'{}'"`
	References         []string          `bun:"refs,array,notnull,default:'{}'"`
	Source             string            `bun:"source,notnull,default:'generic'"`
}

func (r *dbCategory) toModel() model.Category {
	out := model.Category{
		Identifier:         r.Identifier,
		Name:               r.Name,
		Subcategory:        r.Subcategory,
		GenericDescription: r.GenericDescription,
		GenericRemediation: r.GenericRemediation,
		LanguagesOrder:     r.LanguagesOrder,
		References:         r.References,
		Source:             r.Source,
	}
	out.ID = r.ID
	out.CreatedAt = r.CreatedAt
	out.UpdatedAt = r.UpdatedAt
	if out.GenericDescription == nil {
		out.GenericDescription = map[string]string{}
	}
	if out.GenericRemediation == nil {
		out.GenericRemediation = map[string]string{}
	}
	if out.LanguagesOrder == nil {
		out.LanguagesOrder = []string{}
	}
	if out.References == nil {
		out.References = []string{}
	}
	return out
}

type dbTarget struct {
	bun.BaseModel `bun:"table:target,alias:t"`

	ID         uuid.UUID  `bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	CreatedAt  time.Time  `bun:"created_at,nullzero,notnull,default:now()"`
	UpdatedAt  time.Time  `bun:"updated_at,nullzero,notnull,default:now()"`
	IPv4       string     `bun:"ipv4,notnull,default:'',unique:target_natural_key"`
	IPv6       string     `bun:"ipv6,notnull,default:'',unique:target_natural_key"`
	FQDN       string     `bun:"fqdn,notnull,default:'',unique:target_natural_key"`
	Tag        string     `bun:"tag,notnull,default:'',unique:target_natural_key"`
	Protocol   string     `bun:"protocol,notnull,default:''"`
	Port       int        `bun:"port,notnull,default:0"`
	CustomerID *uuid.UUID `bun:"customer_id,nullzero,type:uuid,unique:target_natural_key"`

	Customer *dbCustomer `bun:"rel:belongs-to,join:customer_id=id"`
}

func (r *dbTarget) toModel() model.Target {
	out := model.Target{
		IPv4:     r.IPv4,
		IPv6:     r.IPv6,
		FQDN:     r.FQDN,
		Tag:      r.Tag,
		Protocol: r.Protocol,
		Port:     r.Port,
	}
	out.ID = r.ID
	out.CreatedAt = r.CreatedAt
	out.UpdatedAt = r.UpdatedAt
	if r.CustomerID != nil {
		out.Customer.ID = *r.CustomerID
	}
	if r.Customer != nil {
		out.Customer = r.Customer.toModel()
	}
	return out
}

type dbAssessment struct {
	bun.BaseModel `bun:"table:assessment,alias:a"`

	ID                 uuid.UUID       `bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	CreatedAt          time.Time       `bun:"created_at,nullzero,notnull,default:now()"`
	UpdatedAt          time.Time       `bun:"updated_at,nullzero,notnull,default:now()"`
	CustomerID         uuid.UUID       `bun:"customer_id,notnull,type:uuid,unique:assessment_natural_key"`
	Name               string          `bun:"name,notnull,unique:assessment_natural_key"`
	Language           string          `bun:"language,notnull,unique:assessment_natural_key"`
	StartDateTime      *time.Time      `bun:"start_date_time,nullzero"`
	EndDateTime        *time.Time      `bun:"end_date_time,nullzero"`
	KickoffDateTime    *time.Time      `bun:"kickoff_date_time,nullzero"`
	Status             string          `bun:"status,notnull,default:''"`
	TypeShort          string          `bun:"type_short,notnull,default:''"`
	TypeFull           string          `bun:"type_full,notnull,default:''"`
	CVSSVersions       map[string]bool `bun:"cvss_versions,type:jsonb,notnull,default:'{}'"`
	Environment        string          `bun:"environment,notnull,default:''"`
	TestingType        string          `bun:"testing_type,notnull,default:''"`
	OSSTMMVector       string          `bun:"osstmm_vector,notnull,default:''"`
	VulnerabilityCount int             `bun:"vulnerability_count,notnull,default:0"`

	Customer *dbCustomer `bun:"rel:belongs-to,join:customer_id=id"`
	Targets  []dbTarget  `bun:"m2m:assessment_target,join:Assessment=Target"`
}

func (r *dbAssessment) toModel() model.Assessment {
	out := model.Assessment{
		Name:               r.Name,
		Language:           r.Language,
		Status:             r.Status,
		Type:               model.AssessmentType{Short: r.TypeShort, Full: r.TypeFull},
		Environment:        r.Environment,
		TestingType:        r.TestingType,
		OSSTMMVector:       r.OSSTMMVector,
		VulnerabilityCount: r.VulnerabilityCount,
		CVSSVersions:       r.CVSSVersions,
	}
	out.ID = r.ID
	out.CreatedAt = r.CreatedAt
	out.UpdatedAt = r.UpdatedAt
	if r.StartDateTime != nil {
		out.StartDateTime = *r.StartDateTime
	}
	if r.EndDateTime != nil {
		out.EndDateTime = *r.EndDateTime
	}
	if r.KickoffDateTime != nil {
		out.KickoffDateTime = *r.KickoffDateTime
	}
	out.Customer.ID = r.CustomerID
	if r.Customer != nil {
		out.Customer = r.Customer.toModel()
	}
	return out
}

type dbAssessmentTarget struct {
	bun.BaseModel `bun:"table:assessment_target,alias:at"`

	AssessmentID uuid.UUID `bun:"assessment_id,pk,type:uuid"`
	TargetID     uuid.UUID `bun:"target_id,pk,type:uuid"`

	Assessment *dbAssessment `bun:"rel:belongs-to,join:assessment_id=id"`
	Target     *dbTarget     `bun:"rel:belongs-to,join:target_id=id"`
}

type dbTemplate struct {
	bun.BaseModel `bun:"table:template,alias:tpl"`

	ID           uuid.UUID  `bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	CreatedAt    time.Time  `bun:"created_at,nullzero,notnull,default:now()"`
	UpdatedAt    time.Time  `bun:"updated_at,nullzero,notnull,default:now()"`
	Name         string     `bun:"name,notnull,unique:template_natural_key"`
	Filename     string     `bun:"filename,notnull,default:'',unique:template_natural_key"`
	Language     string     `bun:"language,notnull,default:'',unique:template_natural_key"`
	TemplateType string     `bun:"template_type,notnull,unique:template_natural_key"`
	MimeType     string     `bun:"mime_type,notnull"`
	Identifier   string     `bun:"identifier,notnull,default:'',unique:template_natural_key"`
	FileID       uuid.UUID  `bun:"file_id,notnull,type:uuid"`
	CustomerID   *uuid.UUID `bun:"customer_id,nullzero,type:uuid"`

	Customer *dbCustomer `bun:"rel:belongs-to,join:customer_id=id"`
}

func (r *dbTemplate) toModelBare() model.Template {
	out := model.Template{
		Name:         r.Name,
		Filename:     r.Filename,
		Language:     r.Language,
		TemplateType: r.TemplateType,
		MimeType:     r.MimeType,
		Identifier:   r.Identifier,
		FileID:       r.FileID,
	}
	out.ID = r.ID
	out.CreatedAt = r.CreatedAt
	out.UpdatedAt = r.UpdatedAt
	return out
}

func (r *dbTemplate) toModel() model.Template {
	out := r.toModelBare()
	switch {
	case r.Customer != nil:
		c := r.Customer.toModel()
		out.Customer = &c
	case r.CustomerID != nil:
		out.Customer = &model.Customer{}
		out.Customer.ID = *r.CustomerID
	default:
		out.Customer = &model.Customer{}
	}
	return out
}

type dbVulnerability struct {
	bun.BaseModel `bun:"table:vulnerability,alias:v"`

	ID                        uuid.UUID  `bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	CreatedAt                 time.Time  `bun:"created_at,nullzero,notnull,default:now()"`
	UpdatedAt                 time.Time  `bun:"updated_at,nullzero,notnull,default:now()"`
	AssessmentID              uuid.UUID  `bun:"assessment_id,notnull,type:uuid"`
	CustomerID                uuid.UUID  `bun:"customer_id,notnull,type:uuid"`
	TargetID                  uuid.UUID  `bun:"target_id,notnull,type:uuid"`
	UserID                    *uuid.UUID `bun:"user_id,nullzero,type:uuid"`
	CategoryID                uuid.UUID  `bun:"category_id,notnull,type:uuid"`
	DetailedTitle             string     `bun:"detailed_title,notnull,default:''"`
	Status                    string     `bun:"status,notnull,default:''"`
	CVSSv2Vector              string     `bun:"cvssv2_vector,notnull,default:''"`
	CVSSv2Score               float64    `bun:"cvssv2_score,notnull,default:0"`
	CVSSv3Vector              string     `bun:"cvssv3_vector,notnull,default:''"`
	CVSSv3Score               float64    `bun:"cvssv3_score,notnull,default:0"`
	CVSSv31Vector             string     `bun:"cvssv31_vector,notnull,default:''"`
	CVSSv31Score              float64    `bun:"cvssv31_score,notnull,default:0"`
	CVSSv4Vector              string     `bun:"cvssv4_vector,notnull,default:''"`
	CVSSv4Score               float64    `bun:"cvssv4_score,notnull,default:0"`
	References                []string   `bun:"refs,array,notnull,default:'{}'"`
	Description               string     `bun:"description,notnull,default:''"`
	Remediation               string     `bun:"remediation,notnull,default:''"`
	GenericDescriptionEnabled bool       `bun:"generic_description_enabled,notnull,default:true"`
	GenericRemediationEnabled bool       `bun:"generic_remediation_enabled,notnull,default:false"`
	GenericRemediationText    string     `bun:"generic_remediation_text,notnull,default:''"`

	Assessment *dbAssessment `bun:"rel:belongs-to,join:assessment_id=id"`
	Customer   *dbCustomer   `bun:"rel:belongs-to,join:customer_id=id"`
	Target     *dbTarget     `bun:"rel:belongs-to,join:target_id=id"`
	User       *dbUser       `bun:"rel:belongs-to,join:user_id=id"`
	Category   *dbCategory   `bun:"rel:belongs-to,join:category_id=id"`
}

func (r *dbVulnerability) toModel() model.Vulnerability {
	v := model.Vulnerability{
		DetailedTitle: r.DetailedTitle,
		Status:        r.Status,
		Description:   r.Description,
		Remediation:   r.Remediation,
		References:    r.References,
	}
	v.ID = r.ID
	v.CreatedAt = r.CreatedAt
	v.UpdatedAt = r.UpdatedAt
	v.CVSSv2.Vector = r.CVSSv2Vector
	v.CVSSv2.Score = r.CVSSv2Score
	v.CVSSv3.Vector = r.CVSSv3Vector
	v.CVSSv3.Score = r.CVSSv3Score
	v.CVSSv31.Vector = r.CVSSv31Vector
	v.CVSSv31.Score = r.CVSSv31Score
	v.CVSSv4.Vector = r.CVSSv4Vector
	v.CVSSv4.Score = r.CVSSv4Score
	v.GenericDescription.Enabled = r.GenericDescriptionEnabled
	v.GenericRemediation.Enabled = r.GenericRemediationEnabled
	v.GenericRemediation.Text = r.GenericRemediationText

	v.Assessment.ID = r.AssessmentID
	v.Customer.ID = r.CustomerID
	v.Target.ID = r.TargetID
	v.Category.ID = r.CategoryID
	if r.UserID != nil {
		v.User.ID = *r.UserID
	}
	if r.Assessment != nil {
		v.Assessment.Name = r.Assessment.Name
		v.Assessment.Language = r.Assessment.Language
		v.Assessment.CVSSVersions = r.Assessment.CVSSVersions
	}

	if r.Customer != nil {
		v.Customer.Name = r.Customer.Name
	}
	if r.Target != nil {
		v.Target.IPv4 = r.Target.IPv4
		v.Target.IPv6 = r.Target.IPv6
		v.Target.FQDN = r.Target.FQDN
		v.Target.Port = r.Target.Port
		v.Target.Protocol = r.Target.Protocol
		v.Target.Tag = r.Target.Tag
		if r.Target.CustomerID != nil {
			v.Target.Customer.ID = *r.Target.CustomerID
		}
	}
	if r.User != nil {
		v.User.Username = r.User.Username
	}
	if r.Category != nil {
		v.Category.Identifier = r.Category.Identifier
		v.Category.Name = r.Category.Name
		v.Category.Subcategory = r.Category.Subcategory
		v.Category.LanguagesOrder = r.Category.LanguagesOrder
		v.Category.References = r.Category.References
		v.Category.Source = r.Category.Source
		if v.GenericDescription.Enabled {
			v.GenericDescription.Text = r.Category.GenericDescription[v.Assessment.Language]
		}
		if v.GenericRemediation.Enabled && v.GenericRemediation.Text == "" {
			v.GenericRemediation.Text = r.Category.GenericRemediation[v.Assessment.Language]
		}
	}
	v.Category.GenericDescription = map[string]string{}
	v.Category.GenericRemediation = map[string]string{}
	return v
}

type dbPoc struct {
	bun.BaseModel `bun:"table:poc,alias:p"`

	ID              uuid.UUID       `bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	CreatedAt       time.Time       `bun:"created_at,nullzero,notnull,default:now()"`
	UpdatedAt       time.Time       `bun:"updated_at,nullzero,notnull,default:now()"`
	VulnerabilityID uuid.UUID       `bun:"vulnerability_id,notnull,type:uuid,unique"`
	Items           []model.PocItem `bun:"items,type:jsonb,notnull,default:'[]'"`
}

func (r *dbPoc) toModel() model.Poc {
	out := model.Poc{
		VulnerabilityID: r.VulnerabilityID,
		Pocs:            r.Items,
	}
	out.ID = r.ID
	out.CreatedAt = r.CreatedAt
	out.UpdatedAt = r.UpdatedAt
	if out.Pocs == nil {
		out.Pocs = []model.PocItem{}
	}
	return out
}

type dbPocImage struct {
	bun.BaseModel `bun:"table:poc_image,alias:pi"`

	PocID           uuid.UUID `bun:"poc_id,pk,type:uuid"`
	PocItemID       uuid.UUID `bun:"poc_item_id,pk,type:uuid"`
	FileReferenceID uuid.UUID `bun:"file_reference_id,notnull,type:uuid"`
}
