package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"roving_web/config"
	"roving_web/handlers"
	"roving_web/models"
	"syscall"
	"time"

	gorillaHandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {
	config.Load()

	models.InitializeModels()
	// TODO: Check FD limit in Gatling for performance testing and set to a higher number
	var rlim syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rlim); err != nil {
		fmt.Println("Error getting rlimit:", err)
		return
	}
	// TODO: Consider set to max for resource limit
	// rlim.Cur = rlim.Max
	// syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rlim)
	fmt.Printf("Current Limit: %d, Max Limit: %d\n", rlim.Cur, rlim.Max)

	port := flag.String("port", "80", "server port")
	flag.Parse()

	r := mux.NewRouter()

	corsObj := gorillaHandlers.CORS(
		gorillaHandlers.AllowedOrigins([]string{"*"}),
		gorillaHandlers.AllowedMethods([]string{"POST", "OPTIONS"}),
		gorillaHandlers.MaxAge(86400), // 24 hours
	)

	// TODO: rate limiting on domain level
	r.HandleFunc("/api/event", handlers.EventHandler).Methods("POST")

	r.HandleFunc("/api/bounce-rate", handlers.BounceRateHandler).Methods("GET")
	r.HandleFunc("/api/common-journey", handlers.CommonJourneyHandler).Methods("GET")
	r.HandleFunc("/api/conversion-rate", handlers.ConversionRateHandler).Methods("GET")
	r.HandleFunc("/api/country-ranking", handlers.CountryRankingHandler).Methods("GET")
	r.HandleFunc("/api/current-visitors", handlers.CurrentVisitorsHandler).Methods("GET")
	r.HandleFunc("/api/funnel", handlers.FunnelHandler).Methods("GET")
	r.HandleFunc("/api/top-devices", handlers.TopDevicesHandler).Methods("GET")
	r.HandleFunc("/api/top-referrers", handlers.TopReferrersHandler).Methods("GET")
	r.HandleFunc("/api/page-views", handlers.PageviewsHandler).Methods("GET")
	r.HandleFunc("/api/unique-visitors", handlers.UniqueVisitorsHandler).Methods("GET")

	srv := &http.Server{
		Addr:              ":" + *port,
		Handler:           corsObj(r),
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1 MB	
	}
	log.Default().Println("Listening on port " + *port)
	log.Fatal(srv.ListenAndServe())
}
