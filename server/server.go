package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/NYTimes/gziphandler"
	"github.com/gorilla/mux"
	"github.com/sahib/wedlist/cache"
	"github.com/sahib/wedlist/db"
	"github.com/sahib/wedlist/server/endpoints"
	"github.com/sahib/config"
)

type Server struct {
	db    *db.Database
	srv   *http.Server
	cache *cache.SessionCache
}

func getTLSConfig(cfg *config.Config) (*tls.Config, error) {
	certPath := cfg.String("server.certfile")
	keyPath := cfg.String("server.keyfile")
	if certPath != "" && keyPath != "" {
		cert, err := tls.LoadX509KeyPair(certPath, keyPath)
		if err != nil {
			return nil, err
		}

		// PCI DSS 3.2.1. demands offering TLS >= 1.1:
		return &tls.Config{
			Certificates:             []tls.Certificate{cert},
			MinVersion:               tls.VersionTLS11,
			PreferServerCipherSuites: true,
		}, nil
	}

	return nil, nil
}

func NewServer(cfg *config.Config, db *db.Database, cache *cache.SessionCache) *Server {
	router := mux.NewRouter()
	router.Handle("/api/v0/list", endpoints.NewListHandler(db)).Methods("GET")
	router.Handle("/api/v0/add", endpoints.NewAddHandler(db)).Methods("POST")
	router.Handle("/api/v0/delete", endpoints.NewDelHandler(db)).Methods("POST")
	router.Handle("/api/v0/reserve", endpoints.NewReserveHandler(db)).Methods("POST")
	router.Handle("/api/v0/login", endpoints.NewLoginHandler(db, cache, cfg)).Methods("POST")
	router.Handle("/api/v0/logout", endpoints.NewLogoutHandler(cache)).Methods("GET")
	router.Handle("/api/v0/token/{token}", endpoints.NewTokenHandler(db, cache, cfg)).Methods("GET")

	// Redirects to either login or list view:
	router.Handle("/", endpoints.NoAuth(&indexHandler{db: db, cache: cache}))

	// Static content:
	router.PathPrefix("/").Handler(endpoints.NoAuth(http.FileServer(http.Dir("./static/"))))

	authMiddleware := endpoints.NewAuthMiddleware(db, cache)
	router.Use(authMiddleware.Middleware)

	tlsConfig, err := getTLSConfig(cfg)
	if err != nil {
		log.Printf("warning: failed to load tls config: %v", err)
	}

	return &Server{
		db: db,
		srv: &http.Server{
			Addr:              fmt.Sprintf(":%d", cfg.Int("server.port")),
			Handler:           gziphandler.GzipHandler(router),
			ReadHeaderTimeout: 10 * time.Second,
			WriteTimeout:      10 * time.Second,
			IdleTimeout:       360 * time.Second,
			TLSConfig:         tlsConfig,
		},
	}
}

type indexHandler struct {
	db    *db.Database
	cache *cache.SessionCache
}

func (ih *indexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user, err := endpoints.IsAuthenticated(r, ih.cache, ih.db)
	if err != nil || user == nil {
		log.Printf("login check failed: %v (%s)", err, r.Host)
		http.Redirect(w, r, "/login.html", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/list.html", http.StatusSeeOther)
}

func (srv *Server) Serve() error {
	log.Printf("running on %s", srv.srv.Addr)

	if srv.srv.TLSConfig != nil {
		return srv.srv.ListenAndServeTLS("", "")
	}

	return srv.srv.ListenAndServe()
}

func (srv *Server) Terminate() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return srv.srv.Shutdown(ctx)
}

func (srv *Server) Close() error {
	return srv.srv.Close()
}
