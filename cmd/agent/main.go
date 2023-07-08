package main

import "github.com/CvitoyBamp/metricsexporter/internal/agent"

func main() {
	c := agent.CreateAgent("http://localhost:8080")
	c.RunAgent(2, 10)
}
