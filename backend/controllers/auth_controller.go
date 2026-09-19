package controllers

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
	"golang.org/x/crypto/bcrypt"

	"live-polling-tool/backend/config"
	"live-polling-tool/backend/models"
)

type SignupRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func generateToken(user models.User) (string, error) {
	secret := os.Getenv("JWT_SECRET")

	if secret == "" {
		secret = "change-this-secret-in-production"
	}

	claims := jwt.MapClaims{
		"user_id": user.ID.Hex(),
		"email":   user.Email,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secret))
}

func Signup(c *gin.Context) {
	var req SignupRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request",
		})
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if req.Name == "" || req.Email == "" || len(req.Password) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Name, valid email and password of at least 6 characters are required",
		})
		return
	}

	if config.MongoDB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database is not connected",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := config.MongoDB.Collection("users")

	var existing models.User

	err := collection.FindOne(
		ctx,
		bson.M{"email": req.Email},
	).Decode(&existing)

	if err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": "Email already registered",
		})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(req.Password),
		bcrypt.DefaultCost,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to secure password",
		})
		return
	}

	user := models.User{
		ID:       bson.NewObjectID(),
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	_, err = collection.InsertOne(ctx, user)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create account",
		})
		return
	}

	token, err := generateToken(user)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create login token",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Account created successfully",
		"token":   token,
		"user": gin.H{
			"id":    user.ID.Hex(),
			"name":  user.Name,
			"email": user.Email,
		},
	})
}

func Login(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request",
		})
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if config.MongoDB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database is not connected",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := config.MongoDB.Collection("users")

	var user models.User

	err := collection.FindOne(
		ctx,
		bson.M{"email": req.Email},
	).Decode(&user)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid email or password",
		})
		return
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(req.Password),
	)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid email or password",
		})
		return
	}

	token, err := generateToken(user)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create login token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   token,
		"user": gin.H{
			"id":    user.ID.Hex(),
			"name":  user.Name,
			"email": user.Email,
		},
	})
}
