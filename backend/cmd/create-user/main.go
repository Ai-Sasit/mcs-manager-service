package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"mc-manage-backend/src/utils"
)

func main() {
	_ = godotenv.Load()

	// Resolve credentials: env vars take priority, then positional CLI args.
	username := os.Getenv("CREATE_USERNAME")
	password := os.Getenv("CREATE_PASSWORD")
	role := os.Getenv("CREATE_ROLE")

	args := os.Args[1:]
	if username == "" && len(args) > 0 {
		username = args[0]
	}
	if password == "" && len(args) > 1 {
		password = args[1]
	}
	if role == "" && len(args) > 2 {
		role = args[2]
	}

	// Defaults and validation
	if username == "" || password == "" {
		fmt.Fprintln(os.Stderr, "Usage: CREATE_USERNAME=<u> CREATE_PASSWORD=<p> [CREATE_ROLE=admin|viewer] go run ./cmd/create-user")
		fmt.Fprintln(os.Stderr, "   or: go run ./cmd/create-user <username> <password> [admin|viewer]")
		os.Exit(1)
	}
	if len(password) < 8 {
		fmt.Fprintln(os.Stderr, "Error: password must be at least 8 characters")
		os.Exit(1)
	}
	role = strings.ToLower(role)
	if role != "admin" && role != "viewer" {
		role = "viewer"
	}

	db := utils.ConnectDB()
	defer utils.DisconnectDB(db)

	col := utils.GetCollection(db, "users")

	// Check for duplicate username
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var existing bson.M
	err := col.FindOne(ctx, bson.M{"username": username}).Decode(&existing)
	if err == nil {
		fmt.Fprintf(os.Stderr, "Error: username %q already exists\n", username)
		os.Exit(1)
	}
	if err != mongo.ErrNoDocuments {
		fmt.Fprintf(os.Stderr, "Error querying database: %v\n", err)
		os.Exit(1)
	}

	hash, err := utils.HashPassword(password)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error hashing password: %v\n", err)
		os.Exit(1)
	}

	user := bson.M{
		"id":            uuid.New().String(),
		"username":      username,
		"password_hash": hash,
		"role":          role,
		"created_at":    utils.NowISO(),
	}

	ctx2, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel2()

	if _, err := col.InsertOne(ctx2, user); err != nil {
		fmt.Fprintf(os.Stderr, "Error inserting user: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("[create-user] Created user %q with role %q\n", username, role)
}
