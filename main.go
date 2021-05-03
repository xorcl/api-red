package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/xorcl/api-red/balance"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

type Parser interface {
	GetRoute() string
	StartParser()
	Parse(c *gin.Context)
	StopParser()
}

const API_ROOT = "red"

func main() {
	parsers := []Parser{
		&balance.Parser{},
	}
	r := gin.Default()
	r.Use(CORSMiddleware())
	for _, parser := range parsers {
		parser.StartParser()
		r.GET(fmt.Sprintf("/%s/%s", API_ROOT, parser.GetRoute()), parser.Parse)
		defer parser.StopParser()
	}
	r.Run()
}
