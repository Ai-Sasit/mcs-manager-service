package services

import (
	"fmt"
	"mc-manage-backend/src/utils"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           string `bson:"id" json:"id"`
	Username     string `bson:"username" json:"username"`
	PasswordHash string `bson:"password_hash" json:"password_hash"`
	Role         string `bson:"role" json:"role"` // "admin" or "viewer"
	CreatedAt    string `bson:"created_at" json:"created_at"`
}

type UserPublic struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
}

type UserService struct {
	mu    sync.RWMutex
	users map[string]*User
	col   *mongo.Collection
}

func NewUserService(dataDir string) *UserService {
	svc := &UserService{
		users: make(map[string]*User),
		col:   utils.GetCollection(utils.DB, "users"),
	}
	svc.load()

	if len(svc.users) == 0 {
		adminUser := os.Getenv("ADMIN_USERNAME")
		adminPass := os.Getenv("ADMIN_PASSWORD")
		if adminUser == "" {
			adminUser = "admin"
		}
		if adminPass == "" {
			adminPass = "admin"
		}
		hash, _ := bcrypt.GenerateFromPassword([]byte(adminPass), bcrypt.DefaultCost)
		id := uuid.New().String()
		user := &User{
			ID:           id,
			Username:     adminUser,
			PasswordHash: string(hash),
			Role:         "admin",
			CreatedAt:    time.Now().UTC().Format(time.RFC3339),
		}
		svc.users[id] = user
		svc.saveUser(user)
		logger.Info("[UserService] Seeded default admin user: "+adminUser, nil)
	}

	return svc
}

func (s *UserService) load() {
	ctx, cancel := utils.MongoContext(10 * time.Second)
	defer cancel()

	cursor, err := s.col.Find(ctx, bson.M{})
	if err != nil {
		logger.Error("[UserService] Failed to load users from MongoDB: "+err.Error(), nil)
		return
	}
	defer cursor.Close(ctx)

	var users []*User
	if err := cursor.All(ctx, &users); err != nil {
		logger.Error("[UserService] Failed to decode users from MongoDB: "+err.Error(), nil)
		return
	}
	for _, user := range users {
		s.users[user.ID] = user
	}
	logger.Info(fmt.Sprintf("[UserService] Loaded %d user(s)", len(users)), nil)
}

func (s *UserService) saveUser(user *User) {
	ctx, cancel := utils.MongoContext(10 * time.Second)
	defer cancel()
	if _, err := s.col.ReplaceOne(ctx, bson.M{"id": user.ID}, user, options.Replace().SetUpsert(true)); err != nil {
		logger.Error("[UserService] Failed to save user "+user.ID+": "+err.Error(), nil)
	}
}

func (s *UserService) deleteUser(id string) {
	ctx, cancel := utils.MongoContext(10 * time.Second)
	defer cancel()
	if _, err := s.col.DeleteOne(ctx, bson.M{"id": id}); err != nil {
		logger.Error("[UserService] Failed to delete user "+id+": "+err.Error(), nil)
	}
}

func (s *UserService) Authenticate(username, password string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, u := range s.users {
		if u.Username == username {
			if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
				return nil, fmt.Errorf("invalid credentials")
			}
			return u, nil
		}
	}
	return nil, fmt.Errorf("invalid credentials")
}

func (s *UserService) ListUsers() []UserPublic {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]UserPublic, 0, len(s.users))
	for _, u := range s.users {
		list = append(list, UserPublic{
			ID:        u.ID,
			Username:  u.Username,
			Role:      u.Role,
			CreatedAt: u.CreatedAt,
		})
	}
	return list
}

func (s *UserService) CreateUser(username, password, role string) (*UserPublic, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, u := range s.users {
		if u.Username == username {
			return nil, fmt.Errorf("username already exists")
		}
	}

	if role != "admin" && role != "viewer" {
		role = "viewer"
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password")
	}

	id := uuid.New().String()
	user := &User{
		ID:           id,
		Username:     username,
		PasswordHash: string(hash),
		Role:         role,
		CreatedAt:    time.Now().UTC().Format(time.RFC3339),
	}
	s.users[id] = user
	s.saveUser(user)

	return &UserPublic{
		ID:        user.ID,
		Username:  user.Username,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	}, nil
}

func (s *UserService) UpdateUser(id, username, password, role string) (*UserPublic, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, ok := s.users[id]
	if !ok {
		return nil, fmt.Errorf("user not found")
	}

	if username != "" && username != user.Username {
		for _, u := range s.users {
			if u.Username == username && u.ID != id {
				return nil, fmt.Errorf("username already exists")
			}
		}
		user.Username = username
	}

	if password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password")
		}
		user.PasswordHash = string(hash)
	}

	if role != "" && (role == "admin" || role == "viewer") {
		user.Role = role
	}

	s.saveUser(user)
	return &UserPublic{
		ID:        user.ID,
		Username:  user.Username,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	}, nil
}

func (s *UserService) DeleteUser(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.users[id]; !ok {
		return fmt.Errorf("user not found")
	}

	adminCount := 0
	for _, u := range s.users {
		if u.Role == "admin" {
			adminCount++
		}
	}
	if s.users[id].Role == "admin" && adminCount <= 1 {
		return fmt.Errorf("cannot delete the last admin user")
	}

	delete(s.users, id)
	s.deleteUser(id)
	return nil
}

func (s *UserService) GetUser(id string) (*User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[id]
	return u, ok
}
