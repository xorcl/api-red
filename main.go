package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron"
	"github.com/xorcl/api-red/balance"
	"github.com/xorcl/api-red/bus"
	"github.com/xorcl/api-red/busstop"
	"github.com/xorcl/api-red/common"
	"github.com/xorcl/api-red/metronetwork"
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
		if len(c.Errors) != 0 {
			// 2nd Try
			c.Next()
		}
	}
}

type Parser interface {
	GetRoute() string
	StartParser()
	Parse(c *gin.Context)
	StopParser()
	GetCronTasks() []*common.CronTask
}

func main() {
	parsers := []Parser{
		&balance.Parser{},
		&busstop.Parser{},
		&metronetwork.Parser{},
		&bus.Parser{},
	}
	r := gin.Default()
	r.RedirectTrailingSlash = false
	r.Use(CORSMiddleware())
	c := cron.New()
	for _, parser := range parsers {
		parser.StartParser()
		r.GET(fmt.Sprintf("/%s", parser.GetRoute()), parser.Parse)
		defer parser.StopParser()
		for _, task := range parser.GetCronTasks() {
			c.AddFunc(task.Time, func() {
				log.Printf("Executing %s task...", parser.GetRoute())
				err := task.Execute()
				if err != nil {
					log.Printf("Error executing task: %s", err)
				}
			})
			// Execute now too
			err := task.Execute()
			if err != nil {
				log.Printf("Error executing task: %s", err)
			}
		}
	}
	r.Run()
}
