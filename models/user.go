package models

import (
	"errors"
	"fmt"
	"strconv"
	"time"
	"encoding/json"

	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"

	"golang.org/x/crypto/bcrypt"

	"github.com/astaxie/beego"
)

var (
	UserList map[int64]*User
	gormDB *gorm.DB
)

func init() {
	UserList = make(map[int64]*User)
	u := User{1111, "astaxie", "11111", time.Now(), time.Now(), Profile{"male", 20, "Singapore", "astaxie@gmail.com"}}
	UserList[1111] = &u

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
}

type User struct {
	Id         int64  `bson:"_id" gorm:"primary_key;AUTO_INCREMENT"`
	Name       string  `bson:"name" gorm:"not null;unique;column:name;"`
	Password   string  `bson:"password" gorm:"not null"`
	CreateTime time.Time
	UpdateTime time.Time
	Profile    Profile `bson:"profile" db:"profile"`
}

type User2 struct {
	Id         int64  `gorm:"primary_key;AUTO_INCREMENT"`
	Name       string  `gorm:"not null;unique;column:name;"`
	Password   string  `gorm:"not null"`
	CreateTime time.Time 
	UpdateTime time.Time
	Profile    string
	Profile2   Profile `gorm:"-"`
}

type Profile struct {
	Gender  string `bson:"gender"`
	Age     int    `bson:"age"`
	Address string `bson:"address"`
	Email   string `bson:"email"`
}

func (u *User2) BeforeCreate() error {
	now := time.Now()
	u.CreateTime = now
	u.UpdateTime = now
	data, err := json.Marshal(&u.Profile2)
	if err != nil {
		return err
	}
	u.Profile = string(data)
	return nil
}

func (u *User2) AfterCreate() {
	u.Profile = ""
}

func (u *User2) AfterUpdate() {
	u.Profile = ""
}

func (u *User2) BeforeUpdate() error {
	u.UpdateTime = time.Now()
	data, err := json.Marshal(&u.Profile2)
	if err != nil {
		return err
	}
	u.Profile = string(data)	
	return nil
}

func (u *User2) AfterFind() error {
	err := json.Unmarshal([]byte(u.Profile), &u.Profile2)
	u.Profile = ""
	return err
}

func AddUser(u User) int64 {
	u.Id = time.Now().UnixNano()
	UserList[u.Id] = &u
	return u.Id
}

func GetUser(uid string) (u *User, err error) {
	int64, err := strconv.ParseInt(uid, 10, 64)
	if u, ok := UserList[int64]; ok {
		return u, nil
	}
	return nil, errors.New("User not exists")
}

func GetAllUsers() map[int64]*User {
	return UserList
}

func UpdateUser(uid string, uu *User) (a *User, err error) {
	int64, err := strconv.ParseInt(uid, 10, 64)
	if u, ok := UserList[int64]; ok {
		if uu.Name != "" {
			u.Name = uu.Name
		}
		if uu.Password != "" {
			u.Password = uu.Password
		}
		if uu.Profile.Age != 0 {
			u.Profile.Age = uu.Profile.Age
		}
		if uu.Profile.Address != "" {
			u.Profile.Address = uu.Profile.Address
		}
		if uu.Profile.Gender != "" {
			u.Profile.Gender = uu.Profile.Gender
		}
		if uu.Profile.Email != "" {
			u.Profile.Email = uu.Profile.Email
		}
		return u, nil
	}
	return nil, errors.New("User Not Exist")
}

func Login(username, password string) bool {
	for _, u := range UserList {
		if u.Name == username && u.Password == password {
			return true
		}
	}
	return false
}

func DeleteUser(uid string) {
	//delete(UserList, uid)
}

const (
	// CaptchaWidth is width of image
	CaptchaWidth = 110

	// CaptchaHeight is height of image
	CaptchaHeight = 45

	// CaptchaCodeLen is the length of code
	CaptchaCodeLen = 4

	// BcryptCost is the strength of encryption
	BcryptCost = 12
)

func encryptPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GetAndVerifyUser(name, password string) (*User2, error) {	
	user := &User2{}
	gormDB.Where("name = ?", name).First(user)
	if checkPasswordHash(password, user.Password) {
		return user, nil
	}
	return nil, errors.New("user name or password is wrong")
}
