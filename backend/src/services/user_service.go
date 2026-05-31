package services

import (
	"context"
	"fmt"
	"mc-manage-backend/src/models"
	"mc-manage-backend/src/utils"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type UserService struct {
	mu    sync.RWMutex
	users map[string]*models.User
	col   *mongo.Collection
}

func NewUserService(dataDir string) *UserService {
	svc := &UserService{
		users: make(map[string]*models.User),
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
		hash, _ := utils.HashPassword(adminPass)
		id := uuid.New().String()
		user := &models.User{
			ID:           id,
			Username:     adminUser,
			PasswordHash: hash,
			Role:         "admin",
			CreatedAt:    utils.NowISO(),
		}
		svc.users[id] = user
		svc.saveUser(user)
		logger.Info("[UserService] Seeded default admin user: "+adminUser, nil)
	}

	return svc
}

func (s *UserService) load() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := s.col.Find(ctx, bson.M{})
	if err != nil {
		logger.Error("[UserService] Failed to load users from MongoDB: "+err.Error(), nil)
		return
	}
	defer cursor.Close(ctx)

	var users []*models.User
	if err := cursor.All(ctx, &users); err != nil {
		logger.Error("[UserService] Failed to decode users from MongoDB: "+err.Error(), nil)
		return
	}
	for _, user := range users {
		s.users[user.ID] = user
	}
	logger.Info(fmt.Sprintf("[UserService] Loaded %d user(s)", len(users)), nil)
}

func (s *UserService) saveUser(user *models.User) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if _, err := s.col.ReplaceOne(ctx, bson.M{"id": user.ID}, user, options.Replace().SetUpsert(true)); err != nil {
		logger.Error("[UserService] Failed to save user "+user.ID+": "+err.Error(), nil)
	}
}

func (s *UserService) deleteUser(id string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if _, err := s.col.DeleteOne(ctx, bson.M{"id": id}); err != nil {
		logger.Error("[UserService] Failed to delete user "+id+": "+err.Error(), nil)
	}
}

func (s *UserService) Authenticate(username, password string) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, u := range s.users {
		if u.Username == username {
			if err := utils.CheckPasswordHash(password, u.PasswordHash); err != nil {
				return nil, fmt.Errorf("invalid credentials")
			}
			return u, nil
		}
	}
	return nil, fmt.Errorf("invalid credentials")
}

func (s *UserService) ListUsers() []models.UserPublic {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]models.UserPublic, 0, len(s.users))
	for _, u := range s.users {
		list = append(list, models.UserPublic{
			ID:        u.ID,
			Username:  u.Username,
			Role:      u.Role,
			CreatedAt: u.CreatedAt,
		})
	}
	return list
}

func (s *UserService) CreateUser(username, password, role string) (*models.UserPublic, error) {
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

	hash, err := utils.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password")
	}

	id := uuid.New().String()
	user := &models.User{
		ID:           id,
		Username:     username,
		PasswordHash: hash,
		Role:         role,
		CreatedAt:    utils.NowISO(),
	}
	s.users[id] = user
	s.saveUser(user)

	return &models.UserPublic{
		ID:        user.ID,
		Username:  user.Username,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	}, nil
}

func (s *UserService) UpdateUser(id, username, password, role string) (*models.UserPublic, error) {
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
		hash, err := utils.HashPassword(password)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password")
		}
		user.PasswordHash = hash
	}

	if role != "" && (role == "admin" || role == "viewer") {
		user.Role = role
	}

	s.saveUser(user)
	return &models.UserPublic{
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

func (s *UserService) GetUser(id string) (*models.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[id]
	return u, ok
}

func (s *UserService) GetUserByUsername(username string) (*models.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, u := range s.users {
		if u.Username == username {
			return u, true
		}
	}
	return nil, false
}

// AddToken appends a JWT string to the user's token list and persists it.
func (s *UserService) AddToken(userID, token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	user, ok := s.users[userID]
	if !ok {
		return fmt.Errorf("user not found")
	}
	user.Tokens = append(user.Tokens, token)
	s.saveUser(user)
	return nil
}

// RemoveToken removes a JWT string from the user's token list and persists it.
func (s *UserService) RemoveToken(userID, token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	user, ok := s.users[userID]
	if !ok {
		return fmt.Errorf("user not found")
	}
	filtered := user.Tokens[:0]
	for _, t := range user.Tokens {
		if t != token {
			filtered = append(filtered, t)
		}
	}
	user.Tokens = filtered
	s.saveUser(user)
	return nil
}

// IsTokenValid returns true if the token is in the user's active token list.
func (s *UserService) IsTokenValid(username, token string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, u := range s.users {
		if u.Username == username {
			for _, t := range u.Tokens {
				if t == token {
					return true
				}
			}
			return false
		}
	}
	return false
}

func (s *UserService) ChangeOwnPassword(id, currentPassword, newPassword string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, ok := s.users[id]
	if !ok {
		return fmt.Errorf("user not found")
	}

	if err := utils.CheckPasswordHash(currentPassword, user.PasswordHash); err != nil {
		return fmt.Errorf("current password is incorrect")
	}

	if len(newPassword) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}

	hash, err := utils.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password")
	}
	user.PasswordHash = hash
	s.saveUser(user)
	return nil
}
