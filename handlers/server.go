package handlers

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/statping-ng/statping-ng/utils"
)

func startServer(host string) error {
	httpServer = &http.Server{
		Addr:         host,
		WriteTimeout: timeout,
		ReadTimeout:  timeout,
		IdleTimeout:  timeout,
		Handler:      router,
	}
	httpServer.SetKeepAlivesEnabled(false)
	return httpServer.ListenAndServe()
}

func startSSLServer(ip string, port int) error {
	cfg := &tls.Config{
		MinVersion:               tls.VersionTLS12,
		CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		},
	}
	srv := &http.Server{
		Addr:         fmt.Sprintf("%v:%v", ip, port),
		Handler:      router,
		TLSConfig:    cfg,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
		WriteTimeout: timeout,
		ReadTimeout:  timeout,
		IdleTimeout:  timeout,
	}

	certFile := utils.DirPath("server.crt")
	keyFile := utils.DirPath("server.key")

	return srv.ListenAndServeTLS(certFile, keyFile)
}
