package routes

import (
	"fmt"
	"net/http"
	"os"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/parthpati1102/todo-gin-jwt/config"
	"github.com/parthpati1102/todo-gin-jwt/controllers"
	"github.com/parthpati1102/todo-gin-jwt/models"
)

// SetupRouter initializes all routes and JWT configuration
func SetupRouter() *gin.Engine {
	r := gin.Default()

	fmt.Println("JWT_SECRET:", os.Getenv("JWT_SECRET"))

	r.SetFuncMap(nil)
	r.LoadHTMLGlob("templates/*.html")

	// Serve static files (CSS/JS if any)
	r.Static("/static", "./static")

	// r.GET("/test-register", func(c *gin.Context) {
	// 	c.HTML(http.StatusOK, "register.html", nil)
	// })

	// r.GET("/test-login", func(c *gin.Context) {
	// 	c.HTML(http.StatusOK, "login.html", nil)
	// })

	// r.GET("/test-lists", func(c *gin.Context) {
	// 	c.HTML(http.StatusOK, "lists.html", nil)
	// })

	// r.GET("/test-create", func(c *gin.Context) {
	// 	c.HTML(http.StatusOK, "create.html", nil)
	// })

	// r.GET("/test-edit", func(c *gin.Context) {
	// 	c.HTML(http.StatusOK, "edit.html", nil)
	// })

	// // ✅ Add this test route here
	// r.GET("/test", func(c *gin.Context) {
	// 	fmt.Println("✅ Test route hit")
	// 	c.HTML(http.StatusOK, "test.html", nil)
	// })

	// TEMPORARY: Skip JWT cookie check for testing
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusSeeOther, "/lists")
	})

	// Public routes

	r.GET("/register", controllers.RegisterPage)
	r.POST("/register", controllers.RegisterHandler)
	r.GET("/login", controllers.LoginPage)

	// JWT Middleware setup
	identityKey := os.Getenv("JWT_IDENTITY_KEY")
	if identityKey == "" {
		identityKey = "email"
	}

	authMiddleware, err := jwt.New(&jwt.GinJWTMiddleware{
		Realm:       "todo zone",
		Key:         []byte(os.Getenv("JWT_SECRET")),
		Timeout:     time.Hour * 24,
		MaxRefresh:  time.Hour * 24,
		IdentityKey: identityKey,
		TokenLookup: "cookie:" + getCookieName() + ",header:Authorization",
		CookieName:  getCookieName(),

		// --- Authentication (login) ---
		Authenticator: func(c *gin.Context) (interface{}, error) {
			var loginVals struct {
				Email    string `form:"email" json:"email" binding:"required"`
				Password string `form:"password" json:"password" binding:"required"`
			}
			if err := c.ShouldBind(&loginVals); err != nil {
				return "", jwt.ErrMissingLoginValues
			}
			email := loginVals.Email
			password := loginVals.Password

			coll := config.MongoClient.Database(config.DBName).Collection("users")
			ctx := c.Request.Context()
			user, err := models.GetUserByEmail(ctx, coll, email)
			if err != nil || user == nil {
				return nil, jwt.ErrFailedAuthentication
			}
			if !user.CheckPassword(password) {
				return nil, jwt.ErrFailedAuthentication
			}
			return map[string]interface{}{identityKey: user.Email}, nil
		},

		// --- Payload inside JWT token ---
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(map[string]interface{}); ok {
				return jwt.MapClaims{
					identityKey: v[identityKey],
					"jti":       uuid.New().String(),
				}
			}
			return jwt.MapClaims{}
		},

		// --- Extract identity from claims ---
		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)
			if e, ok := claims[identityKey].(string); ok {
				return map[string]interface{}{identityKey: e}
			}
			return nil
		},

		// --- Authorization check ---
		Authorizator: func(data interface{}, c *gin.Context) bool {
			return true // All authenticated users allowed
		},

		// --- On successful login ---
		LoginResponse: func(c *gin.Context, code int, token string, expire time.Time) {
			cookieName := getCookieName()

			// Use secure cookies in production (HTTPS)
			secure := os.Getenv("GIN_MODE") == "release"

			// Set JWT cookie
			c.SetCookie(
				cookieName,
				token,
				int(time.Until(expire).Seconds()),
				"/",
				"",
				secure,
				true, // httpOnly
			)

			// ✅ No need to manually extract claims — they are not yet available here.
			// The identity will be automatically set by customIdentitySetter
			// on the next authenticated request.

			// Redirect to the lists page after successful login
			c.Redirect(http.StatusSeeOther, "/lists")
		},

		// --- On unauthorized access ---
		Unauthorized: func(c *gin.Context, code int, message string) {
			c.Redirect(http.StatusSeeOther, "/login")
		},

		TimeFunc: time.Now,
	})

	if err != nil {
		panic("JWT Error: " + err.Error())
	}

	// --- Login route ---
	r.POST("/login", authMiddleware.LoginHandler)

	// --- Logout route ---
	r.GET("/logout", controllers.LogoutHandler)

	// Protected group
	auth := r.Group("/")
	auth.Use(authMiddleware.MiddlewareFunc(), customIdentitySetter(identityKey))
	{
		auth.GET("/lists", controllers.ListTodosPage)
		auth.GET("/lists/create", controllers.CreateTodoPage)
		auth.POST("/lists/create", controllers.CreateTodoHandler)
		auth.GET("/lists/edit/:id", controllers.EditTodoPage)
		auth.POST("/lists/edit/:id", controllers.UpdateTodoHandler)
		auth.POST("/lists/delete/:id", controllers.DeleteTodoHandler)
	}

	// Fallback: redirect unknown routes to login
	r.NoRoute(func(c *gin.Context) {
		c.Redirect(http.StatusSeeOther, "/login")
	})

	return r
}

// getCookieName returns cookie name for JWT
func getCookieName() string {
	if v := os.Getenv("COOKIE_NAME"); v != "" {
		return v
	}
	return "jwt"
}

// customIdentitySetter adds identity to Gin context
func customIdentitySetter(identityKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := jwt.ExtractClaims(c)
		if val, ok := claims[identityKey].(string); ok {
			c.Set("identity", val)
		}
		c.Next()
	}
}
