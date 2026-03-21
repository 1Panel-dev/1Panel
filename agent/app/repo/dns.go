package repo

import (
	"context"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/global"
	"gorm.io/gorm"
)

type DnsZoneRepo struct{}

type IDnsZoneRepo interface {
	Page(page, size int, opts ...DBOption) (int64, []model.DnsZone, error)
	List(opts ...DBOption) ([]model.DnsZone, error)
	GetFirst(opts ...DBOption) (model.DnsZone, error)
	Create(ctx context.Context, zone *model.DnsZone) error
	Save(ctx context.Context, zone *model.DnsZone) error
	Update(id uint, vars map[string]interface{}) error
	DeleteBy(ctx context.Context, opts ...DBOption) error
	WithName(name string) DBOption
	WithLikeName(name string) DBOption
}

func NewIDnsZoneRepo() IDnsZoneRepo {
	return &DnsZoneRepo{}
}

func (d *DnsZoneRepo) Page(page, size int, opts ...DBOption) (int64, []model.DnsZone, error) {
	var zones []model.DnsZone
	db := global.DB.Model(&model.DnsZone{})
	for _, opt := range opts {
		db = opt(db)
	}
	count := int64(0)
	db = db.Count(&count)
	if err := db.Limit(size).Offset(size * (page - 1)).Find(&zones).Error; err != nil {
		return count, zones, err
	}
	return count, zones, nil
}

func (d *DnsZoneRepo) List(opts ...DBOption) ([]model.DnsZone, error) {
	var zones []model.DnsZone
	db := global.DB.Model(&model.DnsZone{})
	for _, opt := range opts {
		db = opt(db)
	}
	if err := db.Find(&zones).Error; err != nil {
		return zones, err
	}
	return zones, nil
}

func (d *DnsZoneRepo) GetFirst(opts ...DBOption) (model.DnsZone, error) {
	var zone model.DnsZone
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	if err := db.First(&zone).Error; err != nil {
		return zone, err
	}
	return zone, nil
}

func (d *DnsZoneRepo) Create(ctx context.Context, zone *model.DnsZone) error {
	return getTx(ctx).Create(zone).Error
}

func (d *DnsZoneRepo) Save(ctx context.Context, zone *model.DnsZone) error {
	return getTx(ctx).Save(zone).Error
}

func (d *DnsZoneRepo) Update(id uint, vars map[string]interface{}) error {
	return global.DB.Model(&model.DnsZone{}).Where("id = ?", id).Updates(vars).Error
}

func (d *DnsZoneRepo) DeleteBy(ctx context.Context, opts ...DBOption) error {
	return getTx(ctx, opts...).Delete(&model.DnsZone{}).Error
}

func (d *DnsZoneRepo) WithName(name string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("name = ?", name)
	}
}

func (d *DnsZoneRepo) WithLikeName(name string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		if len(name) == 0 {
			return g
		}
		return g.Where("name like ?", "%"+name+"%")
	}
}

// DnsRecord repo

type DnsRecordRepo struct{}

type IDnsRecordRepo interface {
	Page(page, size int, opts ...DBOption) (int64, []model.DnsRecord, error)
	List(opts ...DBOption) ([]model.DnsRecord, error)
	GetFirst(opts ...DBOption) (model.DnsRecord, error)
	Create(ctx context.Context, record *model.DnsRecord) error
	Save(ctx context.Context, record *model.DnsRecord) error
	DeleteBy(ctx context.Context, opts ...DBOption) error
	WithZoneID(zoneID uint) DBOption
	WithType(recordType string) DBOption
	WithLikeName(name string) DBOption
}

func NewIDnsRecordRepo() IDnsRecordRepo {
	return &DnsRecordRepo{}
}

func (d *DnsRecordRepo) Page(page, size int, opts ...DBOption) (int64, []model.DnsRecord, error) {
	var records []model.DnsRecord
	db := global.DB.Model(&model.DnsRecord{})
	for _, opt := range opts {
		db = opt(db)
	}
	count := int64(0)
	db = db.Count(&count)
	if err := db.Limit(size).Offset(size * (page - 1)).Find(&records).Error; err != nil {
		return count, records, err
	}
	return count, records, nil
}

func (d *DnsRecordRepo) List(opts ...DBOption) ([]model.DnsRecord, error) {
	var records []model.DnsRecord
	db := global.DB.Model(&model.DnsRecord{})
	for _, opt := range opts {
		db = opt(db)
	}
	if err := db.Find(&records).Error; err != nil {
		return records, err
	}
	return records, nil
}

func (d *DnsRecordRepo) GetFirst(opts ...DBOption) (model.DnsRecord, error) {
	var record model.DnsRecord
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	if err := db.First(&record).Error; err != nil {
		return record, err
	}
	return record, nil
}

func (d *DnsRecordRepo) Create(ctx context.Context, record *model.DnsRecord) error {
	return getTx(ctx).Create(record).Error
}

func (d *DnsRecordRepo) Save(ctx context.Context, record *model.DnsRecord) error {
	return getTx(ctx).Save(record).Error
}

func (d *DnsRecordRepo) DeleteBy(ctx context.Context, opts ...DBOption) error {
	return getTx(ctx, opts...).Delete(&model.DnsRecord{}).Error
}

func (d *DnsRecordRepo) WithZoneID(zoneID uint) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("zone_id = ?", zoneID)
	}
}

func (d *DnsRecordRepo) WithType(recordType string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		if len(recordType) == 0 {
			return g
		}
		return g.Where("`type` = ?", recordType)
	}
}

func (d *DnsRecordRepo) WithLikeName(name string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		if len(name) == 0 {
			return g
		}
		return g.Where("name like ?", "%"+name+"%")
	}
}
