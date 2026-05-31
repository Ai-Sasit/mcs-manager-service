package models

type User struct {
	ID           string   `bson:"id" json:"id"`
	Username     string   `bson:"username" json:"username"`
	PasswordHash string   `bson:"password_hash" json:"-"`
	Role         string   `bson:"role" json:"role"` // "admin" or "viewer"
	Tokens       []string `bson:"tokens" json:"-"`
	CreatedAt    string   `bson:"created_at" json:"created_at"`
	UpdatedAt    string   `bson:"updated_at" json:"updated_at"`
}

type UserPublic struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
}
