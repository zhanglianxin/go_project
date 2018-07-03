package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"fmt"
	"log"
	"time"
	"os"
	"io"
	"github.com/gin-gonic/gin/binding"
)

var DB = make(map[string]string)

func setupRouter() *gin.Engine {
	// Disable Console Color
	// gin.DisableConsoleColor()
	r := gin.Default()

	// Ping test
	r.GET("/ping", func(c *gin.Context) {
		// c.String(200, "pong")
        c.JSON(200, gin.H{"message": "pong"})
	})

	// Get user value
	r.GET("/user/:name", func(c *gin.Context) {
		user := c.Params.ByName("name")
		value, ok := DB[user]
		if ok {
			c.JSON(200, gin.H{"user": user, "value": value})
		} else {
			c.JSON(200, gin.H{"user": user, "status": "no value"})
		}
	})

	// Authorized group (uses gin.BasicAuth() middleware)
	// Same than:
	// authorized := r.Group("/")
	// authorized.Use(gin.BasicAuth(gin.Credentials{
	//	  "foo":  "bar",
	//	  "manu": "123",
	//}))
	authorized := r.Group("/", gin.BasicAuth(gin.Accounts{
		"foo":  "bar", // user:foo password:bar
		"manu": "123", // user:manu password:123
	}))

	authorized.POST("admin", func(c *gin.Context) {
		user := c.MustGet(gin.AuthUserKey).(string)

		// Parse JSON
		var json struct {
			Value string `json:"value" binding:"required"`
		}

		if c.Bind(&json) == nil {
			DB[user] = json.Value
			c.JSON(200, gin.H{"status": "ok"})
		}
	})

	return r
}

func main() {
	// r := setupRouter()
	// // Listen and Server in 0.0.0.0:8080
	// r.Run(":8080")

	// Default with the Logger and Recovery middleware already attached
	router := gin.Default()

	multiplartOrUrlencodedForm(router)
	queryAndPostForm(router)
	uploadSingleFile(router)
	uploadMultipleFiles(router)

	v1 := router.Group("/v1")
	{
		// 301
		v1.GET("/", func(context *gin.Context) {
			context.String(http.StatusOK, "this is v1")
		})
	}

	v2 := router.Group("/v2")
	{
		v2.GET("", func(context *gin.Context) {
			context.String(http.StatusOK, "this is v2")
		})
	}

	go func() {
		// Write log file
		f, _ := os.Create(fmt.Sprintf("gin-%s.log", time.Now().Format("2006-01-02")))
		gin.DefaultWriter = io.MultiWriter(f, os.Stdout)
		// Without middleware by default
		router := gin.Default()
		router.Use(gin.Logger())

		router.GET("/benchmark", MyBenchLogger(), func(context *gin.Context) {
			time.Sleep(100 * time.Millisecond)
			context.String(http.StatusOK, "this is the benchmark")
		})

		router.Run(":8090")
	}()

	router.POST("/login_json", func(context *gin.Context) {
		var json Login

		if err := context.ShouldBindWith(&json, binding.JSON); err == nil {
			if json.User == "manu" && json.Password == "123" {
				context.JSON(http.StatusOK, gin.H{"status": "you are logged in"})
			} else {
				context.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
			}
		} else {
			context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
	})

	router.POST("/login_form", func(context *gin.Context) {
		var form Login

		if err := context.ShouldBindWith(&form, binding.Form); err == nil {
			if form.User == "manu" && form.Password == "123" {
				context.JSON(http.StatusOK, gin.H{"status": "you are logged in"})
			} else {
				context.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
			}
		} else {
			context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
	})

	router.POST("/login_query", func(context *gin.Context) {
		var query Login

		if err := context.ShouldBindWith(&query, binding.Default("GET", "")); err == nil {
			if query.User == "manu" && query.Password == "123" {
				context.JSON(http.StatusOK, gin.H{"status": "you are logged in"})
			} else {
				context.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
			}
		} else {
			context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
	})

	router.Run(":8080")
}

func multiplartOrUrlencodedForm(router *gin.Engine) {
	router.POST("form_post", func(context *gin.Context) {
		message := context.PostForm("message")
		nick := context.DefaultPostForm("nick", "anonymous")
		context.JSON(http.StatusOK, gin.H{
			"status": "posted",
			"message": message,
			"nick": nick,
		})
	})
}

func queryAndPostForm(router *gin.Engine) {
	router.POST("/post", func(context *gin.Context) {
		id := context.Query("id")
		page := context.DefaultQuery("page", "0")
		name := context.PostForm("name")
		message := context.PostForm("message")
		fmt.Printf("id: %s, page: %s, name: %s, message: %s", id, page, name, message)
	})
}

func uploadSingleFile(router *gin.Engine) {
	router.POST("/upload", func(context *gin.Context) {
		file, err := context.FormFile("file")
		if err != nil {
			context.String(http.StatusBadRequest, err.Error())
		}
		log.Println(file.Filename)

		context.String(http.StatusOK, fmt.Sprintf("'%s' uploaded!", file.Filename))
	})
}

func uploadMultipleFiles(router *gin.Engine) {
	router.POST("/uploads", func(context *gin.Context) {
		form, err := context.MultipartForm()
		if err != nil {
			context.String(http.StatusBadRequest, err.Error())
		}
		files := form.File["uploads[]"]

		for _, file := range files {
			log.Println(file.Filename)
		}
		context.String(http.StatusOK, fmt.Sprintf("%d files uploaded!", len(files)))
	})
}

// Custom middleware
func MyBenchLogger() gin.HandlerFunc {
	return func(context *gin.Context) {
		t := time.Now()
		context.Next()
		latency := time.Since(t)
		log.Println(latency)
	}
}

type Login struct {
	User string `form:"user" json:"user" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}