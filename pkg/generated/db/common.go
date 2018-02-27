package db

import (
	"database/sql"
	"github.com/Juniper/contrail/pkg/generated/models"
)

type DB struct {
	models.BaseService
	DB *sql.DB
}

func NewService(db *sql.DB) models.Service {
	return &DB{
		BaseService: models.BaseService{},
		DB:          db,
	}
}

//SetDB sets db object.
func (db *DB) SetDB(sqlDB *sql.DB) {
	db.DB = sqlDB
}
