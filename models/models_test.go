package models

import (
	"time"
	"fmt"
	"testing"

	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
)

func TestPostgresJsonbORM(t *testing.T) {
	// Setup Postgres - set CONN_STRING to connect to an empty database
	db, err := gorm.Open("postgres", "dbname=beego user=beego_group password=123456 host=127.0.0.1 port=5432 sslmode=disable")
	if err != nil {
		panic(err)
	}

	// db.LogMode(true)
	passwd, _ := encryptPassword("123456")
	user := &User2{
		Name: "test",
		Password: passwd,
		Profile2: Profile{
			Gender: "男",
			Address: "测试",
			Age: 18,
			Email: "aaa@163.com",
		},
	}
	now := time.Now()
	db.Model(user).Updates(User2{CreateTime: now, UpdateTime: now})
	db.Save(user)
	user2 := User2{}
	db.Where("name = ?", "test").First(&user2)

	// Create Table
	type ClassRoom struct {
		gorm.Model
		State string `sql:"type:JSONB NOT NULL DEFAULT '{}'::JSONB"`
	}

	// AutoMigrate
	db.DropTable(&ClassRoom{})
	db.Debug().AutoMigrate(&ClassRoom{})

	// JSON to insert
	STATE := `{"uses-kica": false, "hide-assessments-intro": true, "most-recent-grade-skew": 1.5}`

	classRoom := ClassRoom{State: STATE}
	db.Save(&classRoom)
	

	// Select the row
	var result ClassRoom
	db.First(&result)

	if result.State == STATE {
		fmt.Println("SUCCESS: Selected JSON == inserted JSON")
	} else {
		fmt.Println("FAILED: Selected JSON != inserted JSON")
		fmt.Println("Inserted: " + STATE)
		fmt.Println("Selected: " + result.State)
	}	
	db.Delete(&classRoom)
	db.Unscoped().Delete(&classRoom)
}
