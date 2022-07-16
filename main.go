package main

import (
	"context"
	"fmt"
	"log"

	"example.com/sarang-apis/configbusiness"
	"example.com/sarang-apis/configstorage"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	server      *gin.Engine
	cs          configstorage.ConfigStorageService
	cc          configbusiness.ConfigController
	ctx         context.Context
	cfgc        *mongo.Collection
	mongoclient *mongo.Client
	err         error
)

func init() {
	ctx = context.TODO()

	mongoconn := options.Client().ApplyURI("mongodb+srv://tuankiet:miniproject@bankaccount.lfuju.mongodb.net/?retryWrites=true&w=majority")
	mongoclient, err = mongo.Connect(ctx, mongoconn)
	if err != nil {
		log.Fatal("error while connecting with mongo", err)
	}
	err = mongoclient.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal("error while trying to ping mongo", err)
	}

	//gcloudpubsub.Init(ctx, "GCP_PROJECT_ID", "...")

	fmt.Println("mongo connection established")

	cfgc = mongoclient.Database("configuration").Collection("config")
	index := mongo.IndexModel{
		Keys: bson.M{
			"name": 1,
			"env":  1,
		},
	}

	cfgc.Indexes().CreateOne(ctx, index)
	cs = configstorage.NewUserService(cfgc, ctx)
	cc = configbusiness.New(cs)
	server = gin.Default()
	//redislock.ConnectRedisV1(&redislock.RedisConnectionConfig{
	//	Addr:     "address",
	//	UserName: "conf_name",
	//	Password: "password",
	//})
}

func main() {
	defer mongoclient.Disconnect(ctx)

	basepath := server.Group("/v1")
	cc.RegisterUserRoutes(basepath)

	log.Fatal(server.Run(":5678"))

}
