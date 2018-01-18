package models

import (
	"fmt"

	_ "github.com/lib/pq"
	"github.com/jinzhu/gorm"
	"github.com/astaxie/beego"
)

var (
	gormDB *gorm.DB
)

func init() {
	dataSource := fmt.Sprintf(
		"dbname=%s user=%s password=%s host=%s port=%d sslmode=disable",
		beego.AppConfig.DefaultString("postgres.database", "beego"),
		beego.AppConfig.DefaultString("postgres.user", "beego_group"),
		beego.AppConfig.DefaultString("postgres.password", "123456"),
		beego.AppConfig.DefaultString("postgres.host", "127.0.0.1"),
		beego.AppConfig.DefaultInt("postgres.port", 5432))

	db, err := gorm.Open("postgres", dataSource)
	if err != nil {
		panic(err)
	}
	gormDB = db
	gormDB.SingularTable(true)
	if !gormDB.HasTable(&CasbinRole{}) {
		gormDB.CreateTable(&CasbinRole{})
	}
	if !gormDB.HasTable(&CasbinUser{}) {
		gormDB.CreateTable(&CasbinUser{})
	}
	if !gormDB.HasTable(&CasbinPermission{}) {
		gormDB.CreateTable(&CasbinPermission{})
	}
}