package service

import (
	"errors"
	"facedrop/models"
	"facedrop/repository"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}
func (s *UserService) CreateUserFace(userFace *models.UserFace) error {
	return s.userRepo.CreateUserFace(userFace)
}
func (s *UserService) RegisterUser(user *models.User) error {
	// Check if user already exists
	existingUser, err := s.userRepo.FindByEmail(user.Email)
	if err == nil && existingUser != nil {
		return errors.New("user already exists")
	}

	// TODO: Hash password before saving
	pwd,_ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	user.Password = string(pwd)
	return s.userRepo.Create(user)
}

func (s *UserService) Login(email, password string) (*models.User, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, err
	}	
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid password")
	}
	return user, nil
}	
func (s *UserService) GenerateToken(userID uint) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}



func (s *UserService) GetUserByID(id uint) (*models.User, error) {
	return s.userRepo.FindByID(id)
}

func (s *UserService) GetUserByEmail(email string) (*models.User, error) {
	return s.userRepo.FindByEmail(email)
}

func (s *UserService) UpdateUser(user *models.User) error {
	return s.userRepo.Update(user)
}

func (s *UserService) DeleteUser(id uint) error {
	return s.userRepo.Delete(id)
} 