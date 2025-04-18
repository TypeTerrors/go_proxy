package app

import (
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"net/url"
	"prx/internal/models"
	"prx/internal/utils"
	"time"
)

func (a *App) HandleRequests(w http.ResponseWriter, req *http.Request) {

	targetURL, err := a.getRedirectionRecords(req.Host)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
	}

	a.Log.Debug("Proxying request", "host", req.Host, "target", targetURL)

	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		a.Response(w, a.Err("invalid url %s", err), http.StatusInternalServerError)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(parsedURL)
	proxy.ServeHTTP(w, req)
}

func (a *App) HandleAddNewProxy(w http.ResponseWriter, req *http.Request) {

	var body models.AddNewProxy
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		a.Response(w, a.Err("request body decode error %s", err), http.StatusInternalServerError)
		return
	}

	if err := utils.ValidateFields(body); err != "" {
		a.Response(w, a.Err("validation error %s", err), http.StatusInternalServerError)
		return
	}

	err := a.Kube.AddNewProxy(body, a.namespace, a.name)
	if err != nil {
		a.Response(w, a.Err("configuration error: %s", err), http.StatusInternalServerError)
		return
	}

	a.setRedirectRecords(body.From, body.To)

	a.Response(w, nil, http.StatusCreated)
}

func (a *App) HandleDeleteProxy(w http.ResponseWriter, req *http.Request) {

	var body models.DelOldProxy
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		a.Response(w, err, http.StatusInternalServerError)
		return
	}

	if errMsg := utils.ValidateFields(body); errMsg != "" {
		a.Response(w, a.Err("validation error: %s", errMsg), http.StatusBadRequest)
		return
	}

	err := a.Kube.DeleteProxy(a.namespace, body.From)
	if err != nil {
		a.Response(w, a.Err("configuration error %s", err), http.StatusInternalServerError)
		return
	}

	a.deleteRedirectRecords(body.From)

	a.Response(w, nil, http.StatusCreated)
}

func (a *App) HandlePatchProxy(w http.ResponseWriter, req *http.Request) {
	var body models.PatchOldProxy
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		a.Response(w, err, http.StatusInternalServerError)
		return
	}

	if errMsg := utils.ValidateFields(body); errMsg != "" {
		a.Response(w, a.Err("validation error: %s", errMsg), http.StatusBadRequest)
		return
	}

	err := a.Kube.DeleteProxy(a.namespace, body.From)
	if err != nil {
		a.Response(w, a.Err("configuration error %s", err), http.StatusInternalServerError)
		return
	}

	a.deleteRedirectRecords(body.From)

	err = a.Kube.AddNewProxy(body, a.namespace, a.name)
	if err != nil {
		a.Response(w, a.Err("configuration error: %s", err), http.StatusInternalServerError)
		return
	}

	a.setRedirectRecords(body.From, body.To)

	a.Response(w, nil, http.StatusCreated)
}

func (a *App) HandleGetRedirectionRecords(w http.ResponseWriter, req *http.Request) {

	records, err := a.getAllRedirectionRecords()
	if err != nil {
		a.Response(w, err, http.StatusInternalServerError)
	}

	if len(records) < 1 {
		a.Response(w, a.Err("no redirection records available"), http.StatusNoContent)
	}

	var res []models.RedirectionRecords
	for i, v := range records {
		record := models.RedirectionRecords{
			From: i,
			To:   v,
		}
		res = append(res, record)
	}

	a.Response(w, res, http.StatusOK)
}

func (a *App) StatusHandler(w http.ResponseWriter, req *http.Request) {
	response := models.Health{
		Status:  "OK",
		Time:    time.Now().Format(time.RFC3339),
		Version: a.version,
	}

	a.Response(w, response, http.StatusOK)
}
