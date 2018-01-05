package models

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	_ "github.com/lib/pq"
)

var (
	UserList map[string]*User
)

func init() {
	UserList = make(map[string]*User)
	u := User{"user_11111", "astaxie", "11111", Profile{"male", 20, "Singapore", "astaxie@gmail.com"}}
	UserList["user_11111"] = &u

	dataSource := fmt.Sprintf(
		"dbname=%s user=%s password=%s host=%s port=%d sslmode=disable",
		beego.AppConfig.String("postgres.database"),
		beego.AppConfig.String("postgres.user"),
		beego.AppConfig.String("postgres.password"),
		beego.AppConfig.String("postgres.host"),
		beego.AppConfig.DefaultInt("postgres.port", 5432))

	orm.RegisterDriver("postgres", orm.DRPostgres)
	orm.RegisterDataBase("default", "postgres", dataSource)
	orm.SetMaxOpenConns("default", 30)
}

type User struct {
	Id       string  `bson:"_id" orm:"auto"`
	Username string  `bson:"name" orm:"unique"`
	Password string  `bson:"password"`
	Profile  Profile `bson:"profile"`
}

type Profile struct {
	Gender  string `bson:"gender"`
	Age     int    `bson:"age"`
	Address string `bson:"address"`
	Email   string `bson:"email"`
}

func AddUser(u User) string {
	u.Id = "user_" + strconv.FormatInt(time.Now().UnixNano(), 10)
	UserList[u.Id] = &u
	return u.Id
}

func GetUser(uid string) (u *User, err error) {
	if u, ok := UserList[uid]; ok {
		return u, nil
	}
	return nil, errors.New("User not exists")
}

func GetAllUsers() map[string]*User {
	return UserList
}

func UpdateUser(uid string, uu *User) (a *User, err error) {
	if u, ok := UserList[uid]; ok {
		if uu.Username != "" {
			u.Username = uu.Username
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
		if u.Username == username && u.Password == password {
			return true
		}
	}
	return false
}

func DeleteUser(uid string) {
	delete(UserList, uid)
}
