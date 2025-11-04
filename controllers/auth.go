package controllers

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/parthpati1102/todo-gin-jwt/config"
	"github.com/parthpati1102/todo-gin-jwt/models"
)

func RegisterPage(c *gin.Context) {
	c.HTML(http.StatusOK, "register.html", gin.H{"error": ""})
}

func LoginPage(c *gin.Context) {
	// ✅ FIXED: previously "base.html"
	c.HTML(http.StatusOK, "login.html", gin.H{"error": ""})
}

func RegisterHandler(c *gin.Context) {
	type req struct {
		Email    string `form:"email" binding:"required,email"`
		Password string `form:"password" binding:"required,min=6"`
	}
	var r req
	if err := c.ShouldBind(&r); err != nil {
		c.HTML(http.StatusBadRequest, "register.html", gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	coll := config.MongoClient.Database(config.DBName).Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	existing, err := models.GetUserByEmail(ctx, coll, r.Email)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "register.html", gin.H{"error": "Server error"})
		return
	}
	if existing != nil {
		c.HTML(http.StatusBadRequest, "register.html", gin.H{"error": "Email already registered"})
		return
	}

	user := models.User{
		Email:     r.Email,
		Password:  r.Password,
		CreatedAt: time.Now(),
	}
	if err := user.HashPassword(); err != nil {
		c.HTML(http.StatusInternalServerError, "register.html", gin.H{"error": "Server error hashing password"})
		return
	}

	_, err = coll.InsertOne(ctx, user)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "register.html", gin.H{"error": "Server error inserting user"})
		return
	}

	c.Redirect(http.StatusSeeOther, "/login")
}

func LogoutHandler(c *gin.Context) {
	cookieName := os.Getenv("COOKIE_NAME")
	if cookieName == "" {
		cookieName = "jwt"
	}
	c.SetCookie(cookieName, "", -1, "/", "", false, true)
	c.Redirect(http.StatusSeeOther, "/login")
}
