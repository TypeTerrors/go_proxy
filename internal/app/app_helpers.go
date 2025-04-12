package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"prx/internal/models"
	"prx/internal/services"
)

func (a *App) getRedirectionRecords(host string) (string, error) {
	var targetURL string
	var ok bool

	targetURL, ok = a.readRedirectRecord(host)
	if !ok {
		a.Log.Warn("No redirect record found in memory for host:", "host", host)
		var err error
		redirectRecords, err := a.Kube.GetProxyMappings(a.namespace, a.name)
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

func (a *App) getAllRedirectionRecords() (map[string]string, error) {

	res := a.RedirectRecords

	if len(res) < 1 {
		redirectRecords, err := a.Kube.GetProxyMappings(a.namespace, a.name)
		if err != nil {
			a.Log.Error("Error getting redirect records from cluster", "err", err)
			return res, fmt.Errorf("no redirect records found in cluster %s", err)
		}

		return redirectRecords, nil
	}
	return res, nil
}

func (a *App) Response(w http.ResponseWriter, data interface{}, statusCode int) {
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

func (a *App) readRedirectRecord(host string) (string, bool) {
	a.mu.Lock()
	targetURL, ok := a.RedirectRecords[host]
	a.mu.Unlock()

	return targetURL, ok
}

func (a *App) deleteRedirectRecords(host string) {
	a.deleteRedirectRecordsInMemory(host)
	a.deleteRedirectRecordsInCluster(host)
}

func (a *App) deleteRedirectRecordsInMemory(host string) {
	a.mu.Lock()
	delete(a.RedirectRecords, host)
	a.mu.Unlock()
}

func (a *App) deleteRedirectRecordsInCluster(host string) {
	a.Kube.DeleteProxy(a.namespace, host+"-ingress", host+"-tls")
}

func (a *App) setRedirectRecords(from, to string) {
	a.setRedirectRecordsInMemory(from, to)
	a.setRedirectRecordsInCluster(from, to)
}

func (a *App) setRedirectRecordsInMemory(from, to string) {
	a.mu.Lock()
	a.RedirectRecords[from] = to
	a.mu.Unlock()
}

func (a *App) setRedirectRecordsInCluster(from, to string) {
	a.Kube.AddProxyMapping(a.namespace, a.name, services.ProxyMapping{
		From: from,
		To:   to,
	})
}

// Use this simply to avoid typing out extra syntax for fmt.Errorf(). Because its shorter thats why...
func (a *App) Err(err string, messages ...any) error {
	return fmt.Errorf(err, messages...)
}
