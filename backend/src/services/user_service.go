package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           string `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"password_hash"`
	Role         string `json:"role"` // "admin" or "viewer"
	CreatedAt    string `json:"created_at"`
}

type UserPublic struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
}

type UserService struct {
	mu      sync.RWMutex
	users   map[string]*User
	dataDir string
}

func NewUserService(dataDir string) *UserService {
	svc := &UserService{
		users:   make(map[string]*User),
		dataDir: dataDir,
	}
	svc.load()

	// Seed default admin if no users exist
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
		svc.users[uuid.New().String()] = &User{
			ID:           uuid.New().String(),
			Username:     adminUser,
			PasswordHash: string(hash),
			Role:         "admin",
			CreatedAt:    time.Now().UTC().Format(time.RFC3339),
		}
		// Fix: use the ID as the map key
		for _, u := range svc.users {
			delete(svc.users, u.ID)
			svc.users[u.ID] = u
			break
		}
		svc.save()
		logger.Info("[UserService] Seeded default admin user: "+adminUser, nil)
	}

	return svc
}

func (s *UserService) load() {
	path := filepath.Join(s.dataDir, "users.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var users map[string]*User
	if err := json.Unmarshal(data, &users); err != nil {
		logger.Error("[UserService] Failed to parse users.json: "+err.Error(), nil)
		return
	}
	s.users = users
	logger.Info(fmt.Sprintf("[UserService] Loaded %d user(s)", len(users)), nil)
}

func (s *UserService) save() {
	path := filepath.Join(s.dataDir, "users.json")
	data, _ := json.MarshalIndent(s.users, "", "  ")
	os.WriteFile(path, data, 0644)
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

	// Check duplicate username
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
	s.save()

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

	// Check duplicate username (exclude self)
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

	s.save()
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

	// Prevent deleting the last admin
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
	s.save()
	return nil
}

func (s *UserService) GetUser(id string) (*User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[id]
	return u, ok
}
