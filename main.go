package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/adityaw24/go-aws-garasi/configs"
	"github.com/adityaw24/go-aws-garasi/internal/handler"
	"github.com/adityaw24/go-aws-garasi/internal/repo"
	"github.com/adityaw24/go-aws-garasi/internal/usecase"
	"github.com/adityaw24/go-aws-garasi/middleware"
	"github.com/adityaw24/go-aws-garasi/utils"
	"github.com/gin-gonic/gin"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current working directory: %v", err)
	}

	cfg, err := configs.LoadConfig(cwd)
	if err != nil {
		log.Fatalf("Error loading config: %s", err)
	}

	router := gin.Default()
	router.Use(middleware.CORSMiddleware())

	// Initialize AWS clients
	client, presignClient, err := configs.ConnectAWS(cfg)
	if err != nil {
		log.Fatal(err)
	}

	timeout := time.Duration(cfg.TIMEOUT) * time.Second

	repoUpload := repo.NewRepoUpload(client, presignClient, cfg.BUCKET_NAME, timeout)
	usecasesUpload := usecase.NewUsecaseUpload(repoUpload)
	handlerUpload := handler.NewHandlerUpload(usecasesUpload)

	router.NoRoute(func(c *gin.Context) {
		utils.ErrorResp(c, http.StatusNotFound, "page not found")
	})

	v1 := router.Group(cfg.API_GROUP)

	v1.POST("/upload", handlerUpload.UploadFile)
	v1.GET("/preview/:key", handlerUpload.PreviewFile)
	v1.PUT("/update", handlerUpload.UpdateFile)
	v1.GET("/list", handlerUpload.ListObjects)
	v1.DELETE("/delete/:key", handlerUpload.DeleteFile)
	v1.PUT("/update-object", handlerUpload.UpdateObject)

	srv := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", cfg.PORT),
		Handler: router,
		// IdleTimeout:  time.Minute,
		// ReadTimeout:  time.Duration(cfg.TIMEOUT) * time.Second,
		// WriteTimeout: time.Duration(cfg.TIMEOUT) * time.Second,
	}

	err = srv.ListenAndServe()
	if err != nil {
		log.Fatalf("Error listening and serving: %v", err)
	}

	// fmt.Println("AWS configuration loaded successfully!")
	// fmt.Println("Region:", cfg.Region)
}
