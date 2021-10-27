package image

import (
	"gorm.io/gorm"
)

const (
	imagesTableName = "images"
)

type entity struct {
	gorm.Model
	OriginalFilename string
	ContentType      string
	Filename         string
	Type             string
	Width            int
	Height           int
}

func (*entity) TableName() string {
	return imagesTableName
}

type Database struct {
	db *gorm.DB
}

func NewDatabase(db *gorm.DB) *Database {
	_ = db.AutoMigrate(&entity{})
	return &Database{
		db: db,
	}
}

func (d Database) Create(originalFilename string, image Image) error {
	e := entity{
		OriginalFilename: originalFilename,
		ContentType:      image.ContentType,
		Filename:         image.Filename,
		Type:             string(image.Type),
		Width:            image.Dimension.Width,
		Height:           image.Dimension.Height,
	}

	return d.db.Create(&e).Error
}

func (d Database) FindByOriginalFilename(originalFilename string) ([]Image, bool, error) {
	return d.runQuery("original_filename = ?", originalFilename)
}

func (d *Database) FindByFilenameAndType(filename string, imageType Type) ([]Image, bool, error) {
	return d.runQuery("original_filename = ? AND type = ?", filename, imageType)
}

func (d Database) runQuery(where string, params ...interface{}) ([]Image, bool, error) {
	var entities []entity

	if query := d.db.Where(where, params...).Find(&entities); query.Error != nil {
		if query.Error == gorm.ErrRecordNotFound {
			return nil, false, nil
		}

		return nil, false, query.Error
	}

	return parseFromEntities(entities), true, nil
}

func parseFromEntities(entities []entity) []Image {
	output := make([]Image, len(entities))
	for index, e := range entities {
		output[index] = Image{
			Type:     Type(e.Type),
			Filename: e.Filename,
			Metadata: Metadata{
				ContentType: e.ContentType,
				Dimension: Dimension{
					Width:  e.Width,
					Height: e.Height,
				},
			},
		}
	}

	return output
}
