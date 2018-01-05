package dao

import (
	"time"

	"golang.org/x/net/context"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/slover2000/beego_demo/models"
	"github.com/slover2000/prisma"
	"github.com/slover2000/prisma/hystrix"
	p "github.com/slover2000/prisma/thirdparty"
)

// MongoConfig is the settings of mongo
type MongoConfig struct {
	Addrs       []string `required:"true"`
	DialTimeout int      `default:"10"`
	Database    string   `required:"true"`
	PoolLimit   int      `default:"10"`
}

var mongoInstance *mgo.Session

// InitMongoClient initialize mongo client
func InitMongoClient(cfg *MongoConfig) error {
	client, err := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:     cfg.Addrs,
		Timeout:   time.Duration(cfg.DialTimeout) * time.Second,
		Database:  cfg.Database,
		PoolLimit: cfg.PoolLimit,
	})

	if err == nil {
		// Optional. Switch the session to a monotonic behavior.
		client.SetMode(mgo.Monotonic, true)
		mongoInstance = client
	}

	return err
}

// CloseMongoClient should close all resources of mongo client
func CloseMongoClient() {
	if mongoInstance != nil {
		mongoInstance.Close()
	}
}

// StoreUserInfo store user info into db
func StoreUserInfo(user *models.User) bool {
	sessionCopy := mongoInstance.Copy()
	defer sessionCopy.Close()

	c := sessionCopy.DB("").C("user")
	if err := c.Insert(user); err != nil {
		return false
	}

	return true
}

// QueryAllUser query user info from db
func QueryAllUser(ctx context.Context) ([]models.User, error) {
	ctx = hystrix.WithGroup(ctx, "mongo")
	ctx = p.JoinDatabaseContextValue(ctx, prisma.MongoName, "test_db", "user", "find", "find all users")
	reqctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	values, err := prisma.Do(
		reqctx,
		func() (interface{}, error) {
			sessionCopy := mongoInstance.Copy()
			defer sessionCopy.Close()

			var users []models.User
			c := sessionCopy.DB("").C("user")
			err := c.Find(bson.M{}).All(&users)
			return users, err
		})
	cancel()

	if err != nil {
		return nil, err
	}
	users, _ := values.([]models.User)
	return users, nil
}

// QueryUser query user info from db
func QueryUser(name string) (*models.User, error) {
	sessionCopy := mongoInstance.Copy()
	defer sessionCopy.Close()

	user := &models.User{}
	c := sessionCopy.DB("").C("user")
	err := c.FindId(name).One(&user)

	return user, err
}

// UserNameExists check wether user name already exists
func UserNameExists(name string) bool {
	sessionCopy := mongoInstance.Copy()
	defer sessionCopy.Close()

	c := sessionCopy.DB("").C("user")
	count, _ := c.FindId(name).Count()

	return count > 0
}

// RemoveUser remove user from db by name
func RemoveUser(id string) bool {
	sessionCopy := mongoInstance.Copy()
	defer sessionCopy.Close()

	c := sessionCopy.DB("").C("user")
	err := c.Remove(bson.M{"_id": id})

	return err == nil
}
