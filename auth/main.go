package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"popug_auth/handlers"
	kafkaapp "popug_auth/kafka"
	"popug_auth/model"
)

var identityKey = "id"
var roleKey = "role"

const postgresDNS = "host=localhost user=base password=secret dbname=auth port=5442 sslmode=disable"

func main() {
	port := os.Getenv("AUTH_PORT")
	r := gin.Default()

	if port == "" {
		port = "9096"
	}

	db, err := gorm.Open(postgres.Open(postgresDNS), &gorm.Config{})
	if err != nil {
		log.Fatalf("err to open DB %v \n", err)
	}
	err = db.AutoMigrate(&model.User{})
	if err != nil {
		log.Fatalf("err DB.AutoMigrate %v\n", err)
	}

	cudUser := kafkaapp.ProducerUserUPD()

	authMiddleware, err := jwt.New(&jwt.GinJWTMiddleware{
		Realm:       "test zone",
		Key:         []byte("MY_LONG_SECRET_KEY_1&*#@!#sd23lejrhvbo8347rtwehfcbsaj,dnc_sa@#@!#sd23"),
		Timeout:     time.Hour,
		MaxRefresh:  time.Hour,
		IdentityKey: identityKey,
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			v, ok := data.(*model.User)
			if ok {
				return jwt.MapClaims{
					identityKey: v.PublicID,
					roleKey:     v.Role,
				}
			}

			return jwt.MapClaims{}
		},
		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)
			userName := claims[identityKey].(string)
			role := claims[roleKey].(string)
			return &model.User{
				UserName: userName,
				Role:     role,
			}
		},
		Authenticator: func(c *gin.Context) (interface{}, error) {
			var loginVals model.LoginForm
			if err := c.ShouldBind(&loginVals); err != nil {
				return "", jwt.ErrMissingLoginValues
			}

			pwdSalt, pwdHash, err := model.CalculatePasswordHash(loginVals.Password)
			if err != nil {
				log.Fatal("err calculatePassHash")
				return nil, err
			}

			var user model.User
			err = db.Where("user_name = ?", loginVals.Username).First(&user).Error
			if err != nil {
				c.JSON(404, gin.H{"code": "USER_NOT_FOUND", "message": "USER NOT FOUND"})
				return nil, gorm.ErrRecordNotFound
			}

			if model.CheckPassword(loginVals.Password, pwdSalt, pwdHash) {
				return &user, nil
			}

			return nil, fmt.Errorf("wrong username or password")
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			c.Redirect(http.StatusFound, "/signup") // Перенаправление на страницу авторизации
			c.Abort()
		},

		TokenLookup:   "header: Authorization, query: token, cookie: token",
		TokenHeadName: "Bearer",

		TimeFunc: time.Now,
		LoginResponse: func(c *gin.Context, code int, token string, expire time.Time) {
			c.SetCookie("token", token, 3600, "/", "localhost", false, true)
			c.Redirect(http.StatusSeeOther, "/")
		},
	})
	if err != nil {
		log.Fatal("JWT Error:" + err.Error())
	}

	errInit := authMiddleware.MiddlewareInit()
	if errInit != nil {
		log.Fatal("authMiddleware.MiddlewareInit() Error:" + errInit.Error())
	}

	r.GET("/signup", func(c *gin.Context) {
		handlers.SiginupPage(c)
	})

	r.GET("/", authMiddleware.MiddlewareFunc(), func(c *gin.Context) {
		handlers.UserListPage(c, db)
	})
	r.GET("/registerPage", func(c *gin.Context) {
		handlers.RegisterPage(c)
	})
	r.POST("/register", func(c *gin.Context) {
		handlers.CreateUserHandler(c, db, cudUser)
	})
	r.POST("/change-role", authMiddleware.MiddlewareFunc(), func(c *gin.Context) {
		handlers.ChangeUserRoleHandler(c, db, cudUser)
	})
	r.POST("/login", authMiddleware.LoginHandler)
	r.POST("/logout", logoutHandler)

	r.NoRoute(authMiddleware.MiddlewareFunc(), func(c *gin.Context) {
		claims := jwt.ExtractClaims(c)
		log.Printf("NoRoute claims: %#v\n", claims)
		c.JSON(404, gin.H{"code": "PAGE_NOT_FOUND", "message": "Page not found"})
	})

	//go kafka.ConsumeKafkaEvent()
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal(err)
	}
}

func logoutHandler(c *gin.Context) {
	c.SetCookie("token", "", -1, "/", c.Request.Host, false, true)
	c.Redirect(http.StatusSeeOther, "/login")
}
