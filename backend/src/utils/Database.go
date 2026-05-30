package utils

import (
	"context"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var DB *mongo.Client
var logger = NewLogger("mc-manage")

func ConnectDB() *mongo.Client {
	_ = godotenv.Load()

	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		logger.Error("[Database] Error connecting to MongoDB: "+err.Error(), nil)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := client.Ping(ctx, nil); err != nil {
		logger.Error("[Database] Error pinging MongoDB: "+err.Error(), nil)
		os.Exit(1)
	}

	DB = client
	EnsureMongoIndexes(DB)
	logger.Info("[Database] Connected to MongoDB", nil)
	return client
}

func DisconnectDB(client *mongo.Client) {
	if client == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Disconnect(ctx); err != nil {
		logger.Warn("[Database] Error disconnecting MongoDB: "+err.Error(), nil)
	}
}

func DatabaseName() string {
	name := os.Getenv("MONGO_DB_NAME")
	if name == "" {
		name = "mc-manage"
	}
	return name
}

func GetCollection(db *mongo.Client, collection string) *mongo.Collection {
	if db == nil {
		db = DB
	}
	if db == nil {
		db = ConnectDB()
	}
	return db.Database(DatabaseName()).Collection(collection)
}

func MongoContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return context.WithTimeout(context.Background(), timeout)
}

func EnsureMongoIndexes(db *mongo.Client) {
	if db == nil {
		return
	}
	ctx, cancel := MongoContext(15 * time.Second)
	defer cancel()

	indexes := map[string][]mongo.IndexModel{
		"servers": {
			{Keys: bson.D{{Key: "id", Value: 1}}, Options: options.Index().SetUnique(true)},
		},
		"users": {
			{Keys: bson.D{{Key: "id", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "username", Value: 1}}, Options: options.Index().SetUnique(true)},
		},
		"schedules": {
			{Keys: bson.D{{Key: "id", Value: 1}}, Options: options.Index().SetUnique(true)},
		},
		"settings": {
			{Keys: bson.D{{Key: "key", Value: 1}}, Options: options.Index().SetUnique(true)},
		},
		"audit_logs": {
			{Keys: bson.D{{Key: "id", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "timestamp", Value: -1}}},
		},
	}

	for collection, models := range indexes {
		if _, err := GetCollection(db, collection).Indexes().CreateMany(ctx, models); err != nil {
			logger.Warn("[Database] Failed to create indexes for "+collection+": "+err.Error(), nil)
		}
	}
}
