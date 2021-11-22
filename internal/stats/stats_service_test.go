package stats

import "testing"

func Test_CalculateAverageClicks(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		service := NewStatsService()
		averages := service.calculateAverageClicks(map[string]int{
			"PR":  5,
			"USA": 10,
			"JP":  15,
		})

		if averages["PR"] != 5.0/float64(service.daysToConsider) {
			t.Errorf("Expected %f, got %f", 5.0/float64(service.daysToConsider), averages["PR"])
		}
	})
}
