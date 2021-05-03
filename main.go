package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/xorcl/api-red/balance"
)

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
	for _, parser := range parsers {
		parser.StartParser()
		r.GET(fmt.Sprintf("/%s/%s", API_ROOT, parser.GetRoute()), parser.Parse)
		defer parser.StopParser()
	}
	r.Run()
}
