package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/ohmayank/hako/internal/services"
)

func main() {
	mux := http.NewServeMux()

	srv := &http.Server{
		Addr: ":6000",
	}

	go func() {
		log.Printf("listening on port http://localhost%s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()
}
