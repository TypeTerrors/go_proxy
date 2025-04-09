package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"prx/internal/logger"
	"prx/internal/models"
	"prx/internal/utils"
	"time"
)

func (a *App) HandleRequests(w http.ResponseWriter, req *http.Request) {
	targetURL, ok := utils.RedirectRecords[req.Host]
	if !ok {
		a.Log.Error("No redirect record found for host:", "host", req.Host)
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	logger.Log.Info("Proxying request", "host", req.Host, "target", targetURL)

	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		a.Log.Error("Failed to parse target URL", "target", targetURL, "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(parsedURL)
	proxy.ServeHTTP(w, req)
}

func (a *App) HandleAddNewProxy(w http.ResponseWriter, req *http.Request) {

	_, err := a.Jwt.ValidateJWT(req.Header["Authorization"][0])
	if err != nil {
		http.Error(w, "Authentication failed", http.StatusUnauthorized)
	}

	var body models.AddNewProxy
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		return
	}

	if errMsg := utils.ValidateFields(body); errMsg != "" {
		http.Error(w, fmt.Sprintf("Validation error: %s", errMsg), http.StatusBadRequest)
		return
	}

	err = a.Kube.AddNewProxy(body, a.namespace)
	if err != nil {
		http.Error(w, "Configuration error", http.StatusInternalServerError)
		return
	}

	a.mu.Lock()
	a.RedirectRecords[body.From] = body.To
	a.mu.Unlock()

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Ingress and TLS secret created successfully"))
}

func (a *App) HandleDeleteProxy(w http.ResponseWriter, req *http.Request) {

	_, err := a.Jwt.ValidateJWT(req.Header["Authorization"][0])
	if err != nil {
		http.Error(w, "Authentication failed", http.StatusUnauthorized)
	}

	var body models.AddNewProxy
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		return
	}

	if errMsg := utils.ValidateFields(body); errMsg != "" {
		http.Error(w, fmt.Sprintf("Validation error: %s", errMsg), http.StatusBadRequest)
		return
	}

	err = a.Kube.DeleteProxy(a.namespace, body.From+"-ingress", body.From+"-tls")
	if err != nil {
		http.Error(w, "Configuration error", http.StatusInternalServerError)
		return
	}

	a.mu.Lock()
	delete(a.RedirectRecords, body.From)
	a.mu.Unlock()

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Ingress and TLS secret delete successfully"))
}

func (a *App) StatusHandler(w http.ResponseWriter, req *http.Request) {
	response := struct {
		Status string `json:"status"`
		Time   string `json:"time"`
	}{
		Status: "OK",
		Time:   time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Log.Error("Failed to encode response", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
