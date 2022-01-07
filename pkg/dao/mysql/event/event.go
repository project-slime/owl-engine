package event

import (
	"owl-engine/pkg/client/database"
	"owl-engine/pkg/model/dbModel"
)

type event struct{}

var EventDto = new(event)

func (e *event) Insert(record *dbModel.Alert) error {
	return database.DB.Model(&dbModel.Alert{}).Create(record).Error
}
