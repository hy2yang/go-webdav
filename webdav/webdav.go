package webdav

import (
	"context"
	"log"
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
	*User
	Auth  bool
	Cors  CorsCfg
	Users map[string]*User
}

// ConfigBasedWebdavHandler is a wrapper around config to expose ServeHTTP only
type ConfigBasedWebdavHandler struct {
	Config        *Config
	allowAllHosts bool
	handlers      map[*User]*webdav.Handler
}

func HandlerFromConfig(c *Config) *ConfigBasedWebdavHandler {
	allowAllHosts := false
	for _, v := range c.Cors.AllowedHosts {
		if v == "*" {
			allowAllHosts = true
			break
		}
	}

	return &ConfigBasedWebdavHandler{Config: c, allowAllHosts: allowAllHosts, handlers: make(map[*User]*webdav.Handler)}
}

func (h *ConfigBasedWebdavHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	u := h.Config.User
	requestOrigin := r.Header.Get("Origin")

	if h.Config.Cors.Enabled && requestOrigin != "" {
		// Add CORS headers before any operation so even on a 401 unauthorized status, CORS will work.
		h.setCORSHeaders(h.Config.Cors, requestOrigin, w)

		if r.Method == "OPTIONS" {
			return
		}
	}

	if h.Config.Auth {
		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
	}

	// get reuest user and auth status
	user, ok := h.checkAuth(r)
	if !ok {
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}

	if user != nil {
		u = user
	}

	if r.Method == "HEAD" {
		w = newResponseWriterNoBody(w)
	}

	h.serveFiles(u, w, r)
}

func (h *ConfigBasedWebdavHandler) setCORSHeaders(cors CorsCfg, requestOrigin string, w http.ResponseWriter) {
	headers := w.Header()

	hostAllowed := isAllowedHost(h.Config.Cors.AllowedHosts, requestOrigin)

	if h.allowAllHosts || hostAllowed {
		headers.Set("Access-Control-Allow-Headers",
			strings.Join(h.Config.Cors.AllowedHeaders, ", "))
		headers.Set("Access-Control-Allow-Methods",
			strings.Join(h.Config.Cors.AllowedMethods, ", "))

		if h.Config.Cors.Credentials {
			headers.Set("Access-Control-Allow-Credentials", "true")
		}

		if len(h.Config.Cors.ExposedHeaders) > 0 {
			headers.Set("Access-Control-Expose-Headers",
				strings.Join(h.Config.Cors.ExposedHeaders, ", "))
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

	if !h.Config.Auth {
		if ok {
			user, _ = h.Config.Users[username]
		}
		return user, true

	} else {
		authorized := true

		if !ok {
			authorized = false
		}

		user, ok := h.Config.Users[username]
		if !ok { //user not found in config
			authorized = false
		}

		if user != nil && !checkPassword(user.Password, password) {
			log.Println("Wrong Password for user", username)
			authorized = false
		}

		return user, authorized
	}
}

func (h *ConfigBasedWebdavHandler) serveFiles(user *User, w http.ResponseWriter, r *http.Request) {
	if !userHasPermission(user, r) {
		http.Error(w, "No permission", http.StatusForbidden)
		return
	}

	handler := h.getHanlderOf(user)

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

// GetConfigBasedWebdavHandler generates a handler from config
// func GetConfigBasedWebdavHandler(cfg Config) ConfigBasedWebdavHandler {
// 	handler := ConfigBasedWebdavHandler{
// 		ServeHTTP: cfg.ServeHTTP,
// 	}
// 	return handler
// }

// ServeHTTP determines if the request is for this plugin, and if all prerequisites are met.
// func (c *Config) ServeHTTP(w http.ResponseWriter, r *http.Request) {
// 	u := c.User
// 	requestOrigin := r.Header.Get("Origin")

// 	// Add CORS headers before any operation so even on a 401 unauthorized status, CORS will work.
// 	if c.Cors.Enabled && requestOrigin != "" {
// 		headers := w.Header()

// 		allowedHeaders := strings.Join(c.Cors.AllowedHeaders, ", ")
// 		allowedMethods := strings.Join(c.Cors.AllowedMethods, ", ")
// 		exposedHeaders := strings.Join(c.Cors.ExposedHeaders, ", ")

// 		allowAllHosts := len(c.Cors.AllowedHosts) == 1 && c.Cors.AllowedHosts[0] == "*"
// 		allowedHost := isAllowedHost(c.Cors.AllowedHosts, requestOrigin)

// 		if allowAllHosts {
// 			headers.Set("Access-Control-Allow-Origin", "*")
// 		} else if allowedHost {
// 			headers.Set("Access-Control-Allow-Origin", requestOrigin)
// 		}

// 		if allowAllHosts || allowedHost {
// 			headers.Set("Access-Control-Allow-Headers", allowedHeaders)
// 			headers.Set("Access-Control-Allow-Methods", allowedMethods)

// 			if c.Cors.Credentials {
// 				headers.Set("Access-Control-Allow-Credentials", "true")
// 			}

// 			if len(c.Cors.ExposedHeaders) > 0 {
// 				headers.Set("Access-Control-Expose-Headers", exposedHeaders)
// 			}
// 		}
// 	}

// 	if r.Method == "OPTIONS" && c.Cors.Enabled && requestOrigin != "" {
// 		return
// 	}

// 	// Authentication
// 	if c.Auth {
// 		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

// 		// Gets the correct user for this request.
// 		username, password, ok := r.BasicAuth()
// 		if !ok {
// 			http.Error(w, "Not authorized", 401)
// 			return
// 		}

// 		user, ok := c.Users[username]
// 		if !ok {
// 			http.Error(w, "Not authorized", 401)
// 			return
// 		}

// 		if !checkPassword(user.Password, password) {
// 			log.Println("Wrong Password for user", username)
// 			http.Error(w, "Not authorized", 401)
// 			return
// 		}

// 		u = user
// 	} else {
// 		// Even if Auth is disabled, we might want to get
// 		// the user from the Basic Auth header. Useful for Caddy
// 		// plugin implementation.
// 		username, _, ok := r.BasicAuth()
// 		if ok {
// 			if user, ok := c.Users[username]; ok {
// 				u = user
// 			}
// 		}
// 	}

// 	// Checks for user permissions relatively to this PATH.
// 	if !u.Allowed(r.URL.Path) {
// 		w.WriteHeader(http.StatusForbidden)
// 		return
// 	}

// 	if r.Method == "HEAD" {
// 		w = newResponseWriterNoBody(w)
// 	}

// 	// If this request modified the files and the user doesn't have permission
// 	// to do so, return forbidden.
// 	/*
// 		if (r.Method == "PUT" || r.Method == "POST" || r.Method == "MKCOL" ||
// 			r.Method == "DELETE" || r.Method == "COPY" || r.Method == "MOVE") &&
// 			!u.Modify {
// 			w.WriteHeader(http.StatusForbidden)
// 			return
// 		}
// 	*/

// 	if isMethodAllowed(r) && !u.Modify {
// 		w.WriteHeader(http.StatusForbidden)
// 		return
// 	}

// 	// Excerpt from RFC4918, section 9.4:
// 	//
// 	// 		GET, when applied to a collection, may return the contents of an
// 	//		"index.html" resource, a human-readable view of the contents of
// 	//		the collection, or something else altogether.
// 	//
// 	// Get, when applied to collection, will return the same as PROPFIND method.
// 	if r.Method == "GET" {
// 		info, err := u.Handler.FileSystem.Stat(context.TODO(), r.URL.Path)
// 		if err == nil && info.IsDir() {
// 			r.Method = "PROPFIND"

// 			if r.Header.Get("Depth") == "" {
// 				r.Header.Add("Depth", "1")
// 			}
// 		}
// 	}

// 	// Runs the WebDAV.
// 	u.Handler.ServeHTTP(w, r)
// }
