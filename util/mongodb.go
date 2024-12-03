package util

import (
	"context"
	"go-graphql-user-svc/config"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetMongoDB(cfg *config.Config) *mongo.Database {
	// Connect to MongoDB
	// dbUser := cfg.DBConfig.User
	// dbPassword := cfg.DBConfig.Password
	// dbHost := cfg.DBConfig.Host
	// dbPort := cfg.DBConfig.Port
	// dsnFmt := "mongodb://%v:%v@%v:%v"
	// dsn := fmt.Sprintf(dsnFmt, dbUser, dbPassword, dbHost, dbPort)
	// client, err := mongo.Connect(
	// 	context.Background(),
	// 	options.Client().ApplyURI(dsn),
	// )
	client, err := mongo.Connect(context.Background(), options.Client().
		ApplyURI("mongodb://172.24.0.3:27017"))
	if err != nil {
		log.Fatal(err)
	}

	db := client.Database(cfg.DBConfig.DBName)
	return db
}
