package models

import (
	"time"
	"fmt"
	"log"
	"reflect"
	"testing"
	//"encoding/json"

	"github.com/lib/pq"
	"github.com/jinzhu/gorm"
)

type Example struct {
	Test pq.Int64Array `gorm:"type:bigint[]"`
}

type JsonbExample struct {
	ID    int64   `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Name  string `gorm:"not null;unique"`
	Roles string `gorm:"type:jsonb not null default '{}'::jsonb"`
}

func TestArrayExample(t *testing.T) {
	db, err := gorm.Open("postgres", "dbname=beego user=beego_group password=123456 host=127.0.0.1 port=5432 sslmode=disable")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	
	db.DropTableIfExists(&Example{})
	db.CreateTable(&Example{})

	TestExample := Example{[]int64{1, 2, 3, 4}}
	db.Debug().Create(&TestExample)

	var Result Example

	db.Debug().First(&Result)
	if !reflect.DeepEqual(TestExample, Result) {
		fmt.Printf("Failure!")
	}	
}

func TestPostgresJsonbORM(t *testing.T) {
	// Setup Postgres - set CONN_STRING to connect to an empty database
	db, err := gorm.Open("postgres", "dbname=beego user=beego_group password=123456 host=127.0.0.1 port=5432 sslmode=disable")
	if err != nil {
		panic(err)
	}
	defer db.Close()

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

func TestAdminUser(t *testing.T) {
	db, err := gorm.Open("postgres", "dbname=beego user=beego_group password=123456 host=127.0.0.1 port=5432 sslmode=disable")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	db.SingularTable(true)

	db.DropTableIfExists(&JsonbExample{})
	db.CreateTable(&JsonbExample{})

	// roles := make(CasbinUserRoleList, 0)
	// roles = append(roles, CasbinUserRole{ID: 1, Name: "test"})
	// roles = append(roles, CasbinUserRole{ID: 2, Name: "test2"})
	// data, _ := json.Marshal(roles)
	// db.Debug().Save(&JsonbExample{Name: "userA", Roles: string(data)})

	// user := &JsonbExample{}
	// err = db.First(user, 1).Error
	// if err == nil {
	// 	var roles CasbinUserRoleList
	// 	json.Unmarshal([]byte(user.Roles), &roles)
	// 	data := &CasbinUserData{ID: user.ID, Name: user.Name, Roles: roles}
	// 	log.Printf("%d", data)
	// }
}

func TestAdminPermission(t *testing.T) {
	db, err := gorm.Open("postgres", "dbname=beego user=beego_group password=123456 host=127.0.0.1 port=5432 sslmode=disable")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	db.SingularTable(true)

	db.Debug().Delete(&CasbinPermission{Model: Model{ID: 8}})
	db.Debug().Where("id IN (?)", []uint{11, 12}).Delete(CasbinPermission{})
	
	p := CasbinPermission{
		Name: "test",
		Resource: "/url/r",
		Action: "GET",
	}
	// AutoMigrate
	//db.DropTable(&CasbinPermission{})
	if !db.HasTable(&CasbinPermission{}) {
		db.Debug().AutoMigrate(&CasbinPermission{})	
	}
	
	//db.DropTable(&CasbinGroup{})
	if !db.HasTable(&CasbinGroup{}) {
		db.Debug().CreateTable(&CasbinGroup{})		
	}	

	err = db.Debug().Save(&p).Error
	if err != nil {
		t.Errorf("save permission failed:%v", err)
		return
	}	
	g := CasbinGroup{
		Name: "group",
		Permissions: make([]CasbinPermission, 0),
	}
	g.Permissions = append(g.Permissions, p)

	err = db.Debug().Save(&g).Error
	if err != nil {
		t.Errorf("save permission group failed:%v", err)
		return
	}
	g.Permissions = append(g.Permissions, CasbinPermission{Name: "test2", Resource: "/url/2", Action: "POST"})
	err = db.Debug().Save(&g).Error

	g2 := &CasbinGroup{}
	var p2 []CasbinPermission
	db.Debug().First(&g2, 1).Association("Permissions").Find(&p2)
	g2.Permissions = p2
	log.Printf("permissions:%d", len(g2.Permissions))

	var g3 []CasbinGroup
	err = db.Debug().Preload("Permissions").Find(&g3).Error // load all permissions associated with group in one shot
	if err != nil {
		t.Errorf("save permission group failed:%v", err)
		return
	}	

	p3 := &CasbinPermission{Name: "test3", Resource: "/url/3", Action: "PUT"}
	db.Debug().Save(p3)
	gormDB.Model(&CasbinGroup{Model: gorm.Model{ID: g2.ID}}).Association("Permissions").Replace([]CasbinPermission{CasbinPermission{Model: Model{ID: p3.ID}}})

	//db.Debug().Model(&g).Association("Permissions").Delete(&p)
	db.DropTable(&CasbinRole{})
	if !db.HasTable(&CasbinRole{}) {
		db.CreateTable(&CasbinRole{})	
	}
	role := &CasbinRole{
		Name: "aaa", 
		Permissions: []CasbinPermission{*p3},
	}
	db.Debug().Save(role)
}

func TestPermissionV2(t *testing.T) {
	db, err := gorm.Open("postgres", "dbname=beego user=beego_group password=123456 host=127.0.0.1 port=5432 sslmode=disable")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	db.SingularTable(true)
	
	db.DropTableIfExists(&CasbinPermissionV2{})
	db.CreateTable(&CasbinPermissionV2{})

	p1 := &CasbinPermissionV2{Name:"folder1", Parent: 0}
	db.Save(p1)

	c1 := &CasbinPermissionV2{Name:"folder1", Parent: p1.ID, Resource: "/url/c1", Action:"GET"}
	db.Save(c1)
	c2 := &CasbinPermissionV2{Name:"folder1", Parent: p1.ID, Resource: "/url/c2", Action:"GET"}
	db.Save(c2)

	var roots []CasbinPermissionV2
	err = db.Where("parent = ?", 0).Find(&roots).Error
	for i := range roots {
		var children []CasbinPermissionV2
		db.Where("parent = ?", roots[i].ID).Find(&children)
		roots[i].Children = children
	}

	if err != nil {
		t.Errorf("save permission group failed:%v", err)
		return
	}	
}