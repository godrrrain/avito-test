package main

import (
	"context"
	"fmt"

	"avitotest/src/handler"
	"avitotest/src/storage"
	"avitotest/src/tool"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	postgresURL := fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s",
		"postgres", 5432, "program", "banners", "test")
	psqlDB, err := storage.NewPgStorage(context.Background(), postgresURL)
	if err != nil {
		fmt.Printf("Postgresql init: %s", err)
	} else {
		fmt.Println("Connected to PostreSQL")
	}
	defer psqlDB.Close()

	handler := handler.NewHandler(psqlDB)

	router := gin.Default()

	router.Use(cors.Default())

	router.GET("/user_banner", tool.AuthUserMiddleware(), handler.GetBannerForUser)
	router.GET("/banner", tool.AuthAdminMiddleware(), handler.GetBanners)
	router.POST("/banner", tool.AuthAdminMiddleware(), handler.CreateBanner)
	router.PATCH("/banner/:id/", tool.AuthAdminMiddleware(), handler.UpdateBanner)
	router.DELETE("/banner/:id/", tool.AuthAdminMiddleware(), handler.DeleteBanner)

	router.GET("/manage/health", handler.GetHealth)

	router.Run(":8080")
}
