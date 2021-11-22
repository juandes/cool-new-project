package main

import (
	"github.com/juandes/cool-new-project/internal/stats"
)

func main() {
	service := stats.NewStatsService()
	service.Start()
}
