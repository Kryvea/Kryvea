package db

import (
	"context"

	"github.com/Kryvea/Kryvea/internal/model"
	"github.com/Kryvea/Kryvea/internal/store"
)

type SettingIndex struct{ driver *Driver }

func (si *SettingIndex) Get(ctx context.Context) (*model.Setting, error) {
	var row dbSetting
	err := idbFrom(ctx, si.driver.db).NewSelect().
		Model(&row).
		Where("id = ?", model.SettingID).
		Scan(ctx)
	if err != nil {
		return nil, mapErr(err)
	}
	out := row.toModel()
	return &out, nil
}

func (si *SettingIndex) Update(ctx context.Context, setting *model.Setting) error {
	_, err := idbFrom(ctx, si.driver.db).NewUpdate().
		Model((*dbSetting)(nil)).
		Set("max_image_size = ?", setting.MaxImageSize).
		Set("default_category_language = ?", setting.DefaultCategoryLanguage).
		Where("id = ?", model.SettingID).
		Exec(ctx)
	return mapErr(err)
}

func (si *SettingIndex) ValidateImageSize(ctx context.Context, size int64) error {
	s, err := si.Get(ctx)
	if err != nil {
		return err
	}
	if size > s.MaxImageSize {
		return store.ErrFileSizeTooLarge
	}
	return nil
}
