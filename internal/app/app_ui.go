package app

import (
	"net/http"
	"os"
	"strconv"
	"time"
	"strings"

	"prx/internal/models"
)

const uiCookieName = "prx_auth"

func (a *App) adminRoutes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", a.uiFront)
	mux.HandleFunc("GET /login", a.uiLoginGet)
	mux.HandleFunc("POST /login", a.uiLoginPost)
	mux.HandleFunc("POST /logout", a.uiLogoutPost)
	mux.HandleFunc("GET /dashboard", a.uiDashboardGet)
	mux.HandleFunc("POST /dashboard/add", a.uiDashboardAddPost)
	mux.HandleFunc("POST /dashboard/patch", a.uiDashboardPatchPost)
	mux.HandleFunc("POST /dashboard/delete", a.uiDashboardDeletePost)
	return mux
}

func (a *App) UILoginMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/login" {
			next.ServeHTTP(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/dashboard") || r.URL.Path == "/logout" {
			c, err := r.Cookie(uiCookieName)
			if err != nil || c == nil || c.Value == "" {
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}
			if _, err := a.Jwt.ValidateJWT(c.Value); err != nil {
				clearUICookie(w, r)
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (a *App) uiFront(w http.ResponseWriter, r *http.Request) {
	a.renderTemplate(w, r, "front.html", nil, http.StatusOK)
}

func (a *App) uiLoginGet(w http.ResponseWriter, r *http.Request) {
	a.renderTemplate(w, r, "login.html", map[string]any{
		"Error": r.URL.Query().Get("error"),
	}, http.StatusOK)
}

func (a *App) uiLoginPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/login?error=invalid+form", http.StatusFound)
		return
	}
	token := r.Form.Get("token")
	if token == "" {
		http.Redirect(w, r, "/login?error=missing+token", http.StatusFound)
		return
	}
	if _, err := a.Jwt.ValidateJWT(token); err != nil {
		http.Redirect(w, r, "/login?error=invalid+token", http.StatusFound)
		return
	}
	setUICookie(w, r, token)
	http.Redirect(w, r, "/dashboard", http.StatusFound)
}

func (a *App) uiLogoutPost(w http.ResponseWriter, r *http.Request) {
	clearUICookie(w, r)
	http.Redirect(w, r, "/", http.StatusFound)
}

func (a *App) uiDashboardGet(w http.ResponseWriter, r *http.Request) {
	records, _ := a.getAllRedirectionRecords()
	type row struct{ From, To string }
	var list []row
	for from, to := range records {
		list = append(list, row{From: from, To: to})
	}
	a.renderTemplate(w, r, "dashboard.html", map[string]any{
		"Records": list,
		"Error":   r.URL.Query().Get("error"),
	}, http.StatusOK)
}

func (a *App) uiDashboardAddPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/dashboard?error=invalid+form", http.StatusFound)
		return
	}
	from := r.Form.Get("from")
	to := r.Form.Get("to")
	if from == "" || to == "" {
		http.Redirect(w, r, "/dashboard?error=missing+fields", http.StatusFound)
		return
	}
	body := models.AddNewProxy{
		From: from,
		To:   to,
		Cert: r.Form.Get("cert"),
		Key:  r.Form.Get("key"),
	}
	if err := a.Kube.AddNewProxy(body, a.namespace, a.name); err != nil {
		http.Redirect(w, r, "/dashboard?error=config+error", http.StatusFound)
		return
	}
	a.setRedirectRecords(from, to)
	http.Redirect(w, r, "/dashboard", http.StatusFound)
}

func (a *App) uiDashboardPatchPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/dashboard?error=invalid+form", http.StatusFound)
		return
	}
	from := r.Form.Get("from")
	to := r.Form.Get("to")
	if from == "" || to == "" {
		http.Redirect(w, r, "/dashboard?error=missing+fields", http.StatusFound)
		return
	}
	if err := a.Kube.DeleteProxy(a.namespace, from); err != nil {
		http.Redirect(w, r, "/dashboard?error=delete+failed", http.StatusFound)
		return
	}
	a.deleteRedirectRecords(from)
	body := models.PatchOldProxy{
		From: from,
		To:   to,
		Cert: r.Form.Get("cert"),
		Key:  r.Form.Get("key"),
	}
	if err := a.Kube.AddNewProxy(body, a.namespace, a.name); err != nil {
		http.Redirect(w, r, "/dashboard?error=add+failed", http.StatusFound)
		return
	}
	a.setRedirectRecords(from, to)
	http.Redirect(w, r, "/dashboard", http.StatusFound)
}

func (a *App) uiDashboardDeletePost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/dashboard?error=invalid+form", http.StatusFound)
		return
	}
	from := r.Form.Get("from")
	if from == "" {
		http.Redirect(w, r, "/dashboard?error=missing+from", http.StatusFound)
		return
	}
	if err := a.Kube.DeleteProxy(a.namespace, from); err != nil {
		http.Redirect(w, r, "/dashboard?error=delete+failed", http.StatusFound)
		return
	}
	a.deleteRedirectRecords(from)
	http.Redirect(w, r, "/dashboard", http.StatusFound)
}

func setUICookie(w http.ResponseWriter, r *http.Request, token string) {
	age := 12 * time.Hour
	if v := os.Getenv("PRX_UI_COOKIE_MAX_AGE"); v != "" {
		if secs, err := strconv.Atoi(v); err == nil && secs > 0 {
			age = time.Duration(secs) * time.Second
		}
	}
	secure := false
	if r.Header.Get("X-Forwarded-Proto") == "https" || os.Getenv("PRX_UI_COOKIE_SECURE") == "true" {
		secure = true
	}
	http.SetCookie(w, &http.Cookie{
		Name:     uiCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(age.Seconds()),
		Secure:   secure,
	})
}

func clearUICookie(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     uiCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}
