package webdav

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/net/webdav"
)

// CorsCfg is the CORS config.
type CorsCfg struct {
	Enabled        bool
	Credentials    bool
	AllowedHeaders []string
	AllowedHosts   []string
	AllowedMethods []string
	ExposedHeaders []string
}

// Config is the configuration of a WebDAV instance.
type Config struct {
	Auth  bool
	Cors  CorsCfg
	Users map[string]*User
}

// ConfigBasedWebdavHandler is a wrapper around config to expose ServeHTTP only
type ConfigBasedWebdavHandler struct {
	config        *Config
	allowAllHosts bool
	handlers      map[*User]*webdav.Handler
}

//HandlerFromConfig generates config based handler
func HandlerFromConfig(c *Config) *ConfigBasedWebdavHandler {
	allowAllHosts := false
	for _, v := range c.Cors.AllowedHosts {
		if v == "*" {
			allowAllHosts = true
			break
		}
	}

	return &ConfigBasedWebdavHandler{
		config:        c,
		allowAllHosts: allowAllHosts,
		handlers:      make(map[*User]*webdav.Handler),
	}
}

func (h *ConfigBasedWebdavHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if h.config.Cors.Enabled {
		if requestOrigin := r.Header.Get("Origin"); requestOrigin != "" {
			// Add CORS headers before any operation so even on a 401 unauthorized status, CORS will work.
			h.setCORSHeaders(h.config.Cors, requestOrigin, w)

			if r.Method == "OPTIONS" {
				return
			}
		}
	}

	if h.config.Auth {
		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
	}

	// get reuest user and auth status
	user, ok := h.checkAuth(r)
	if !ok {
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}

	h.serveFiles(user, w, r)

}

func (h *ConfigBasedWebdavHandler) setCORSHeaders(cors CorsCfg, requestOrigin string, w http.ResponseWriter) {
	headers := w.Header()

	hostAllowed := isAllowedHost(h.config.Cors.AllowedHosts, requestOrigin)

	if h.allowAllHosts || hostAllowed {
		headers.Set("Access-Control-Allow-Headers",
			strings.Join(h.config.Cors.AllowedHeaders, ", "))
		headers.Set("Access-Control-Allow-Methods",
			strings.Join(h.config.Cors.AllowedMethods, ", "))

		if h.config.Cors.Credentials {
			headers.Set("Access-Control-Allow-Credentials", "true")
		}

		if len(h.config.Cors.ExposedHeaders) > 0 {
			headers.Set("Access-Control-Expose-Headers",
				strings.Join(h.config.Cors.ExposedHeaders, ", "))
		}
	}

	if h.allowAllHosts {
		headers.Set("Access-Control-Allow-Origin", "*")
	} else if hostAllowed {
		headers.Set("Access-Control-Allow-Origin", requestOrigin)
	}
}

func (h *ConfigBasedWebdavHandler) checkAuth(r *http.Request) (user *User, authorized bool) {
	username, password, ok := r.BasicAuth()

	if !h.config.Auth {
		if ok {
			user, _ = h.config.Users[username]
		}
		return user, true
	}

	authorized = false
	user, ok = h.config.Users[username]

	if ok && user != nil {

		if !checkPassword(user.Password, password) {
			fmt.Println("Wrong Password for user", username)
		} else {
			authorized = true
		}

	}

	return user, authorized

}

func (h *ConfigBasedWebdavHandler) serveFiles(user *User, w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.Method, r.URL)

	if !userHasPermission(user, r) {
		http.Error(w, "No permission", http.StatusForbidden)
		return
	}

	handler := h.getHanlderOf(user)

	if r.Method == "HEAD" {
		w = newResponseWriterNoBody(w)
	}

	if r.Method == "GET" {
		info, err := handler.FileSystem.Stat(context.TODO(), r.URL.Path)
		if err == nil && info.IsDir() {
			r.Method = "PROPFIND"

			if r.Header.Get("Depth") == "" {
				r.Header.Add("Depth", "1")
			}
		}
	}

	handler.ServeHTTP(w, r)
}

func (h *ConfigBasedWebdavHandler) getHanlderOf(user *User) *webdav.Handler {
	handler, ok := h.handlers[user]

	if !ok {
		handler = &webdav.Handler{
			FileSystem: webdav.Dir(user.Scope),
			LockSystem: webdav.NewMemLS(),
		}
		fmt.Println(webdav.Dir(user.Scope))
		h.handlers[user] = handler
	}

	return handler
}

// responseWriterNoBody is a wrapper used to suprress the body of the response
// to a request. Mainly used for HEAD requests.
type responseWriterNoBody struct {
	http.ResponseWriter
}

// newResponseWriterNoBody creates a new responseWriterNoBody.
func newResponseWriterNoBody(w http.ResponseWriter) *responseWriterNoBody {
	return &responseWriterNoBody{w}
}

// Header executes the Header method from the http.ResponseWriter.
func (w responseWriterNoBody) Header() http.Header {
	return w.ResponseWriter.Header()
}

// Write suprresses the body.
func (w responseWriterNoBody) Write(data []byte) (int, error) {
	return 0, nil
}

// WriteHeader writes the header to the http.ResponseWriter.
func (w responseWriterNoBody) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
}
