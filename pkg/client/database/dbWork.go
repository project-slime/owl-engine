package database

import "gorm.io/gorm"

type DBWork struct {
	db  *gorm.DB
	end bool
}

func NewWork() *DBWork {
	return &DBWork{
		db:  DB,
		end: false,
	}
}

func (d *DBWork) Begin() *gorm.DB {
	d.db = d.db.Begin()
	d.end = false
	return d.db
}

func (d *DBWork) Commit() {
	if !d.end {
		d.db.Commit()
		d.end = true
	}
}

func (d *DBWork) Rollback() {
	if !d.end {
		d.db.Rollback()
		d.end = true
	}
}
