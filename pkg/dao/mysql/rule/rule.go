package rule

import (
	"errors"
	"strings"

	"owl-engine/pkg/client/database"
	"owl-engine/pkg/model/apiModel"
	"owl-engine/pkg/model/dbModel"
)

type rule struct{}

var RuleDto = new(rule)

func (r *rule) SelectByCondition(condition *apiModel.MathRuleCondition) (*[]dbModel.Rule, int64, error) {
	db := database.DB.Model(&dbModel.Rule{})

	if condition.Id > 0 {
		db = db.Where("id = ?", condition.Id)
	}

	if strings.Compare(condition.Name, "") != 0 {
		db = db.Where("name like ?", "%"+condition.Name+"%")
	}

	if strings.Compare(condition.Creator, "") != 0 {
		db = db.Where("creator = ?", condition.Creator)
	}

	if strings.Compare(condition.ResponsiblePeople, "") != 0 {
		db = db.Where("responsible_people = ?", condition.ResponsiblePeople)
	}

	if strings.Compare(condition.Origin, "") != 0 {
		db = db.Where("origin = ?", condition.Origin)
	}

	if strings.Compare(condition.Type, "") != 0 {
		db = db.Where("business_type = ?", condition.Type)
	}

	if condition.Category != 0 {
		db = db.Where("category = ?", condition.Category)
	}

	if condition.Switch > 0 {
		db = db.Where("switch = ?", condition.Switch)
	}

	if condition.Inuse > 0 {
		db = db.Where("inuse = ?", condition.Inuse)
	}

	var count int64
	db.Count(&count)

	var record = make([]dbModel.Rule, 0, condition.Size)
	offset := (condition.Page - 1) * condition.Size

	// 按照更新时间进行排序
	return &record, count, db.Offset(int(offset)).Limit(int(condition.Size)).Order("updated_at desc").Scan(&record).Error
}

func (r *rule) SelectByIds(ids []int) (*[]dbModel.Rule, int64, error) {
	db := database.DB.Model(&dbModel.Rule{}).Where("id in (?)", ids)

	var count int64
	db.Count(&count)

	var record = make([]dbModel.Rule, 0)
	return &record, count, db.Order("created_at desc").Scan(&record).Error
}

func (r *rule) Insert(data *dbModel.Rule) error {
	work := database.NewWork()
	db := work.Begin()
	defer work.Rollback()

	err := db.Create(data).Error
	if err == nil {
		work.Commit()
	}

	return err
}

// gorm 使用 struct 类型对象作为参数时，
// struct会首先转化为map对象，然后再生成SQL语句，但是转化为map的过程中，对于零值字段会被忽略
// 致使其零值不会被更新
func (r *rule) Update(data *dbModel.Rule) error {
	work := database.NewWork()
	db := work.Begin()
	defer work.Rollback()

	var count int64
	db = db.Model(&dbModel.Rule{}).Where("id = ?", data.ID).Count(&count)
	if count == 0 {
		return errors.New("update error, because record not found")
	}

	err := db.Updates(data).Error
	if err == nil {
		work.Commit()
	}
	return err
}

func (r *rule) Save(data *dbModel.Rule) error {
	work := database.NewWork()
	db := work.Begin()
	defer work.Rollback()

	var err error
	var record dbModel.Rule
	err = db.Model(&dbModel.Rule{}).Where("id = ?", data.ID).First(&record).Error
	if err == nil {
		record.Name = data.Name
		record.CalculateType = data.CalculateType
		record.Express = data.Express
		record.MetricList = data.MetricList
		record.Threshold = data.Threshold
		record.Unit = data.Unit
		record.TimeWindow = data.TimeWindow
		record.Duration = data.Duration
		record.Origin = data.Origin
		record.BusinessType = data.BusinessType
		record.Category = data.Category
		record.ExtensionCondition = data.ExtensionCondition
		record.Level = data.Level
		record.Creator = data.Creator
		record.Updater = data.Updater
		record.ResponsiblePeople = data.ResponsiblePeople
		record.Crontab = data.Crontab
		record.Switch = data.Switch
		record.Inuse = data.Inuse
		record.GroupIp = data.GroupIp
		record.WebHooks = data.WebHooks
		record.Description = data.Description
		record.UpdatedAt = data.UpdatedAt

		err = db.Save(&record).Error
	}

	if err == nil {
		work.Commit()
	}

	return err
}

func (r *rule) Delete(updater string, ids []int) (err error) {
	work := database.NewWork()
	db := work.Begin()
	defer work.Rollback()

	err = db.Model(&dbModel.Rule{}).Where("id in (?)", ids).UpdateColumn("updater", updater).Error
	err = db.Model(&dbModel.Rule{}).Where("id in (?)", ids).Delete(&dbModel.Rule{}).Error
	if err == nil {
		work.Commit()
	}

	return
}
