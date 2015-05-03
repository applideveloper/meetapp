package db

import (
	"fmt"

	"os"

	"net/url"

	"strings"

	"golang.org/x/net/context"
	"gopkg.in/mgo.v2"
)

const mongoDBName = "meetapp"

type mongodb string

var databaseName string

func MongoDB(ctx context.Context) *mgo.Session {
	key := mongodb(mongoDBName)
	db, _ := ctx.Value(key).(*mgo.Session)
	return db
}

func DBName() string {
	return databaseName
}

func OpenMongoDB(ctx context.Context) context.Context {
	uri, dbName := getHerokuURI()
	databaseName = dbName
	fmt.Println("mongoDB", uri, databaseName)

	sesh, err := mgo.Dial(uri)
	if err != nil {
		panic(err)
	}
	ctx = context.WithValue(ctx, mongodb(mongoDBName), sesh)
	return ctx
}

func getHerokuURI() (uri string, dbName string) {
	// default
	uri = fmt.Sprintf("%s:%d", "localhost", 27017)
	dbName = mongoDBName

	mongoURI := os.Getenv("MONGOLAB_URI")
	if mongoURI == "" {
		return
	}
	mongoInfo, err := url.Parse(mongoURI)
	if err != nil {
		return
	}

	uri = mongoURI
	dbName = strings.Replace(mongoInfo.Path, "/", "", 1)
	return
}

func CloseMongoDB(ctx context.Context) context.Context {
	sesh := MongoDB(ctx)
	if sesh == nil {
		fmt.Println("not found mongoDB")
	}
	sesh.Close()
	ctx = context.WithValue(ctx, mongodb(mongoDBName), nil)
	return ctx
}
