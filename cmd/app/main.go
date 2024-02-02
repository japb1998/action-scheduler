package main

import (
	"flag"
	"fmt"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/japb1998/action-scheduler/internal/controller/schedule"
)

func main() {
	port := flag.Int("p", 8080, "port")
	r := gin.Default()

	corsConfig := cors.DefaultConfig()

	corsConfig.AllowOrigins = []string{"*"}
	// To be able to send tokens to the server.
	corsConfig.AllowCredentials = true
	corsConfig.AllowHeaders = []string{"*"}
	corsConfig.AddAllowMethods("OPTIONS", "GET", "PUT", "PATCH")

	r.Use(cors.New(corsConfig))

	schedules := r.Group("/schedule")

	schedules.GET("", schedule.GetSchedules)
	schedules.POST("", schedule.CreateSchedule)
	schedules.GET("/:id", schedule.GetScheduleByID)
	schedules.DELETE(":id", schedule.DeleteSchedule)

	if err := r.Run(fmt.Sprintf(":%d", *port)); err != nil {
		panic(err)
	}
}
