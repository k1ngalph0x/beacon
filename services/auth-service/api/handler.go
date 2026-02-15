package api

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/k1ngalph0x/beacon/services/auth-service/config"
	"github.com/k1ngalph0x/beacon/services/auth-service/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Handler struct {
	DB           *gorm.DB
	Config       *config.Config
}

type SignUpRequest struct{
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type SignInRequest struct{
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type CreateProjectRequest struct {
	Name string `json:"name" binding:"required,min=2,max=100"`
}


type Claims struct{
	UserID string `json:"user_id"`
	Email string `json:"email"`
	Role string `json:"role"`
	jwt.RegisteredClaims
}

func NewHandler(db *gorm.DB, cfg *config.Config) *Handler {
	return &Handler{DB: db, Config: cfg}
}

func generateKey(prefix string) (string, error){
	b := make([]byte, 24)
	_, err := rand.Read(b)
	if err != nil{
		return "", err
	}

	return prefix + base64.RawURLEncoding.EncodeToString(b), nil
}


func (h *Handler) GenerateJWT(userId, email string)(string, error){
	expiration := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: userId,
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiration),
			IssuedAt: jwt.NewNumericDate(time.Now()),
			Issuer: "Beacon-auth",
			Subject: userId,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(h.Config.TOKEN.JwtKey))

	if err != nil{
		fmt.Println(err)
		return "", err
	}

	return tokenString, nil
}

func(h *Handler) GenerateRefreshToken(userId string) (string, error){
	tokenByte := make([]byte, 32)
	_, err := rand.Read(tokenByte)
	if err != nil{
		return "", err
	}

	token := base64.RawURLEncoding.EncodeToString(tokenByte)

	refreshToken := models.RefreshToken{
		UserId:    userId,
		Token:     token,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour), 
	}
	
	if err := h.DB.Create(&refreshToken).Error; err != nil {
		return "", err
	}

	return token, nil
}

func(h *Handler) Refresh(c *gin.Context){
	var tokenRecord models.RefreshToken
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing refresh token"})
		return
	}
	err = h.DB.Preload("User").Where("token = ?", refreshToken).First(&tokenRecord).Error 
	if err != nil{
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	if time.Now().After(tokenRecord.ExpiresAt) {
		h.DB.Delete(&tokenRecord)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token expired"})
		return
	}

	token, err := h.GenerateJWT(
		tokenRecord.User.UserId,
		tokenRecord.User.Email,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})

}

func(h *Handler) SignUp(c *gin.Context){

	var req SignUpRequest
	var existingUser models.User

	err := c.ShouldBindJSON(&req) 
	if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
        return
    }

	if req.Email == "" || req.Password == ""{
		c.JSON(http.StatusBadRequest, gin.H{"error":"Email and password are required"})
		return 
	} 

	email := strings.ToLower(strings.TrimSpace(req.Email))
	password := strings.TrimSpace(req.Password)

	result := h.DB.Where("email = ?", email).First(&existingUser)
 	if result.Error == nil {
        c.JSON(http.StatusConflict, gin.H{"error":"Email already exists"})
        return 
    } else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
        c.JSON(http.StatusInternalServerError, gin.H{"error":"Database error"})
        return 
    }
	
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	
	if err!=nil{
		c.JSON(http.StatusInternalServerError, gin.H{"error":"Failed to create a new user"})
		return 
	}

	user := models.User{
		Email: email,
		Password: string(hashedPassword),
	}

	result = h.DB.Create(&user)

	if result.Error != nil{
		c.JSON(http.StatusInternalServerError, gin.H{"error":"Failed to create a new user"})
		return 
	}

	token, err := h.GenerateJWT(user.UserId, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	refreshToken, err := h.GenerateRefreshToken(user.UserId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	c.SetCookie(
		"refresh_token",
		refreshToken,
		30*24*60*60,
		"/",
		"",
		false, 
		true,  
	)

	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"token":   token,
		"user": gin.H{
			"id":    user.UserId,
			"email": user.Email,
		},
	})
	
}

func(h *Handler) SignIn(c *gin.Context) {
	var req SignInRequest
	var existingUser models.User

	err := c.ShouldBindJSON(&req) 
	if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
        return
    }

	if req.Email == "" || req.Password == ""{
		c.JSON(http.StatusBadRequest, gin.H{"error":"Email and password are required"})
		return 
	} 

	email := strings.ToLower(strings.TrimSpace(req.Email))
	password := strings.TrimSpace(req.Password)
		
	result := h.DB.Where("email = ?", email).First(&existingUser)

	if result.Error != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(password))			
	if err!=nil{
		c.JSON(http.StatusUnauthorized, gin.H{"error":"Invalid credentials"})
		return 
	}

	token, err := h.GenerateJWT(existingUser.UserId, email)

	if err!=nil{
		c.JSON(http.StatusInternalServerError, gin.H{"error":"Failed to create a new user"})
		return 
	}

	refreshToken, err := h.GenerateRefreshToken(existingUser.UserId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong"})
		return
	}
	//expiresAt := time.Now().Add(30 * 24 * time.Hour)

	c.SetCookie("refresh_token",
		refreshToken,
		30*24*60*60,
		"/",
		"",
		false,  
		true,  
	)

	c.JSON(http.StatusOK, gin.H{"message":"Login successful", "token":token, "email": existingUser.Email})
	
}

func(h *Handler) CreateProject(c *gin.Context){
	var req CreateProjectRequest

	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	userId, exists := c.Get("user_id")
	if !exists{
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	publicKey, err := generateKey("pk_live_")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate key"})
		return
	}

	secretKey, err := generateKey("sk_live_")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate key"})
		return
	}

	project := models.Project{
		UserId: userId.(string),
		Name: req.Name,
		PublicKey: publicKey,
		SecretKey: secretKey,
		IsActive: true,
	}

	result := h.DB.Create(&project)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Project created",
		"project": gin.H{
			"id":         project.ID,
			"name":       project.Name,
			"public_key": project.PublicKey,
			"secret_key": project.SecretKey,
		},
	})
}