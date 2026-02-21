// Copyright 2026 José Antonio Garibay Marcelo
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"log"
	"net/http"
	"os"

	"kyberlink-gateway/gateway"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	gateway.InitSessionStore()
	gateway.InitSecurity()

	port := os.Getenv("PORT")
	if port == "" {
		port = "45782"
	}

	backendURL := os.Getenv("BACKEND_URL")
	if backendURL != "" {
		gateway.BackendURL = backendURL
	}

	allowedOrigin := os.Getenv("CORS_ORIGIN")
	if allowedOrigin == "" {
		allowedOrigin = "*"
	}

	r := gin.Default()

	// 1MB body limit
	r.Use(func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)
		c.Next()
	})

	// CORS
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Kyber-Token")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	r.GET("/kempublic", gateway.RateLimitMiddleware(), gateway.HandleInitSession)
	r.POST("/gateway", gateway.HandleGateway)

	log.Printf("KyberLink Gateway running on :%s -> %s (CORS: %s)", port, gateway.BackendURL, allowedOrigin)

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
