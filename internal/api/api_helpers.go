package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"prx/internal/models"
	"prx/internal/services"
)

func (a *Api) getRedirectionRecords(host string) (string, error) {
	var targetURL string
	var ok bool

	targetURL, ok = a.readRedirectRecord(host)
	if !ok {
		a.Log.Warn("No redirect record found in memory for host:", "host", host)
		var err error
		redirectRecords, err := a.Kube.GetProxyMappings(a.Namespace, a.Name)
		if err != nil {
			a.Log.Error("Error getting redirect records from cluster", "err", err)
			return "", fmt.Errorf("no redirect records found in cluster for host %s", host)
		}

		targetURL, ok = redirectRecords[host]
		if !ok {
			a.Log.Error("No redirect records found in cluster for host:", "host", host)
			return "", fmt.Errorf("no redirect records found in cluster for host %s", host)
		}
	}

	return targetURL, nil
}

func (a *Api) GetAllRedirectionRecords() (map[string]string, error) {

	res := a.RedirectRecords

	if len(res) < 1 {
		redirectRecords, err := a.Kube.GetProxyMappings(a.Namespace, a.Name)
		if err != nil {
			a.Log.Error("Error getting redirect records from cluster", "err", err)
			return res, fmt.Errorf("no redirect records found in cluster %s", err)
		}

		return redirectRecords, nil
	}
	return res, nil
}

func (a *Api) Response(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	var response any

	switch v := data.(type) {
	case error:
		response = models.Response{
			Success: false,
			Error:   v.Error(),
		}
		a.Log.Error(v.Error())
	default:
		if data != nil {
			response = data
		} else {
			response = models.Response{
				Success: true,
			}
			a.Log.Debug(response)
		}

	}

	if encodeErr := json.NewEncoder(w).Encode(response); encodeErr != nil {
		a.Log.Error("Failed to encode response")
		http.Error(w, "Failed to encode response", statusCode)
	}
}

func (a *Api) printSettings(token, secret string) {
	a.Log.Info("=====================================")
	a.Log.Info("Project    ", "name   ", a.Name)
	a.Log.Info("Version    ", "version", a.Version)
	a.Log.Info("JWT Secret ", "secret ", secret)
	a.Log.Info("JWT Token  ", "token  ", token)
	a.Log.Info("=====================================")
}

func (a *Api) readRedirectRecord(host string) (string, bool) {
	a.mu.Lock()
	targetURL, ok := a.RedirectRecords[host]
	a.mu.Unlock()

	return targetURL, ok
}

func (a *Api) DeleteRedirectRecords(host string) {
	a.deleteRedirectRecordsInMemory(host)
	a.deleteRedirectRecordsInCluster(host)
}

func (a *Api) deleteRedirectRecordsInMemory(host string) {
	a.mu.Lock()
	delete(a.RedirectRecords, host)
	a.mu.Unlock()
}

func (a *Api) deleteRedirectRecordsInCluster(host string) {
	a.Kube.DeleteProxy(a.Namespace, host)
}

func (a *Api) SetRedirectRecords(from, to string) {
	a.setRedirectRecordsInMemory(from, to)
	a.setRedirectRecordsInCluster(from, to)
}

func (a *Api) setRedirectRecordsInMemory(from, to string) {
	a.mu.Lock()
	a.RedirectRecords[from] = to
	a.mu.Unlock()
}

func (a *Api) setRedirectRecordsInCluster(from, to string) error {
	err := a.Kube.AddProxyMapping(a.Namespace, a.Name, services.ProxyMapping{
		From: from,
		To:   to,
	})
	return err
}

// Use this simply to avoid typing out extra syntax for fmt.Errorf(). Because its shorter thats why...
func (a *Api) Err(err string, messages ...any) error {
	return fmt.Errorf(err, messages...)
}
