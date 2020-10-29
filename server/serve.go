package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/fiatjaf/ilno/config"
	"github.com/fiatjaf/ilno/database"
	"github.com/fiatjaf/ilno/ilno"
	"github.com/fiatjaf/ilno/logger"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/sony/sonyflake"
)

// Serve starts a new HTTP server.
func Serve(cfg config.Config) {
	server := &http.Server{
		Addr:           cfg.Host + ":" + cfg.Port,
		Handler:        setupHandler(cfg),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    20 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	logger.Info(`Listening on %q without TLS`, server.Addr)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		logger.Fatal(`Server failed to start: %v`, err)
	}
}

func setupHandler(cfg config.Config) http.Handler {
	router := mux.NewRouter()

	storage, err := database.New(cfg.Database, 1*time.Second)
	if err != nil {
		logger.Fatal("init database failed: %s", err)
	}
	registerRoute(router, ilno.New(cfg, storage))

	c := cors.New(cors.Options{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowCredentials: true,
		AllowedHeaders:   []string{"Origin", "Referer", "Content-Type"},
		ExposedHeaders:   []string{"X-Set-Cookie", "Date"},
		AllowedMethods:   []string{"HEAD", "GET", "POST", "PUT", "DELETE"},
		Debug:            false,
	})

	return setRequestID(sonyflakeRequestID())(c.Handler(router))
}

func setRequestID(nextRequestID func() string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-Id")
			if requestID == "" {
				requestID = nextRequestID()
			}
			ctx := context.WithValue(r.Context(), ilno.ILNOContextKey, requestID)
			w.Header().Set("X-Request-Id", requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func sonyflakeRequestID() func() string {
	sf := sonyflake.NewSonyflake(sonyflake.Settings{})
	return func() string {
		v, err := sf.NextID()
		if err != nil {
			// NextID can continue to generate IDs for about 174 years from StartTime.
			// But after the Sonyflake time is over the limit, NextID returns an error.
			return "174 years later"
		}
		return fmt.Sprintf("%X", v)
	}
}
