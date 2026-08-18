package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"suporter-backend/internal/middleware"
)

func TestCORSMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Open Mode with *", func(t *testing.T) {
		r := gin.New()
		r.Use(middleware.CORSMiddleware("*"))
		r.GET("/api/v1/test", func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})

		req, _ := http.NewRequest(http.MethodGet, "/api/v1/test", nil)
		req.Header.Set("Origin", "https://untrusted.com")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
		if origin := w.Header().Get("Access-Control-Allow-Origin"); origin != "*" {
			t.Errorf("expected Access-Control-Allow-Origin '*', got '%s'", origin)
		}
	})

	t.Run("Whitelist Mode - Allowed Origin", func(t *testing.T) {
		r := gin.New()
		r.Use(middleware.CORSMiddleware("https://suporter.id,http://localhost:3000"))
		r.GET("/api/v1/test", func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})

		req, _ := http.NewRequest(http.MethodGet, "/api/v1/test", nil)
		req.Header.Set("Origin", "https://suporter.id")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if origin := w.Header().Get("Access-Control-Allow-Origin"); origin != "https://suporter.id" {
			t.Errorf("expected Access-Control-Allow-Origin 'https://suporter.id', got '%s'", origin)
		}
	})

	t.Run("Whitelist Mode - Public OBS Overlay Bypass", func(t *testing.T) {
		r := gin.New()
		r.Use(middleware.CORSMiddleware("https://suporter.id"))
		r.GET("/overlay/:uuid", func(c *gin.Context) {
			c.String(http.StatusOK, "overlay")
		})

		req, _ := http.NewRequest(http.MethodGet, "/overlay/123-abc", nil)
		req.Header.Set("Origin", "null") // OBS browser source origin
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if origin := w.Header().Get("Access-Control-Allow-Origin"); origin != "*" {
			t.Errorf("expected Access-Control-Allow-Origin '*' for OBS overlay, got '%s'", origin)
		}
	})

	t.Run("Preflight OPTIONS Request", func(t *testing.T) {
		r := gin.New()
		r.Use(middleware.CORSMiddleware("https://suporter.id"))
		r.POST("/api/v1/donations", func(c *gin.Context) {
			c.String(http.StatusOK, "created")
		})

		req, _ := http.NewRequest(http.MethodOptions, "/api/v1/donations", nil)
		req.Header.Set("Origin", "https://suporter.id")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected OPTIONS preflight status 200, got %d", w.Code)
		}
	})
}
