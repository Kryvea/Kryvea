package db

import (
	"github.com/Kryvea/Kryvea/internal/store"
)

var (
	_ store.Store              = (*Driver)(nil)
	_ store.AssessmentStore    = (*AssessmentIndex)(nil)
	_ store.CategoryStore      = (*CategoryIndex)(nil)
	_ store.CustomerStore      = (*CustomerIndex)(nil)
	_ store.FileReferenceStore = (*FileReferenceIndex)(nil)
	_ store.PocStore           = (*PocIndex)(nil)
	_ store.SettingStore       = (*SettingIndex)(nil)
	_ store.TargetStore        = (*TargetIndex)(nil)
	_ store.TemplateStore      = (*TemplateIndex)(nil)
	_ store.UserStore          = (*UserIndex)(nil)
	_ store.VulnerabilityStore = (*VulnerabilityIndex)(nil)
)
