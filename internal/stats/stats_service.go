package stats

import (
	"context"
	"net/http"
	"sync"
)

// StatsService is a service that provides the average clicks by country for a user's default group.
type StatsService struct {
	wg sync.WaitGroup

	httpSrv *http.Server
	httpMux *http.ServeMux
}

// NewStatsService creates a new StatsService.
func NewStatsService() *StatsService {
	mux := http.NewServeMux()
	mux.HandleFunc("/averages", computeAveragesByCountry)

	s := &StatsService{
		httpSrv: &http.Server{
			Addr:    ":8080",
			Handler: mux,
		},
		httpMux: mux,
	}

	s.wg.Add(1)
	go s.serve()
	return s
}

func (s *StatsService) serve() {
	defer s.wg.Done()
	s.httpSrv.ListenAndServe()
}

func (s *StatsService) Shutdown(ctx context.Context) error {
	err := s.httpSrv.Shutdown(ctx)
	s.wg.Wait()
	return err
}
