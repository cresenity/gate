package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func runCORS(method, origin string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(CORS)
	r.GET("/api/info", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	r.DELETE("/api/domain/x", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	req := httptest.NewRequest(method, "/api/info", nil)
	if method == http.MethodDelete {
		req = httptest.NewRequest(method, "/api/domain/x", nil)
	}
	req.Header.Set("Origin", origin)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	return w
}

// BUG-03. cors.New() menghasilkan gin.HandlerFunc yang langsung dibuang, tidak pernah
// dipanggil, sehingga tidak satu pun header CORS ikut pada balasan.
func TestCORSMustSetAllowOriginHeader(t *testing.T) {
	w := runCORS(http.MethodGet, "https://devcloud.cresenity.com")

	if w.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Fatalf("BUG-03: Access-Control-Allow-Origin tidak diset; seluruh header balasan: %v", w.Header())
	}
}

// BUG-04. Router mendaftarkan DELETE /api/domain/:domain, tetapi AllowMethods pada
// konfigurasi CORS hanya memuat GET, POST, PUT, PATCH.
func TestCORSMustAllowDeleteMethod(t *testing.T) {
	w := runCORS(http.MethodDelete, "https://devcloud.cresenity.com")

	allow := w.Header().Get("Access-Control-Allow-Methods")
	if allow == "" {
		t.Skipf("dilewati: header CORS memang tidak pernah diset, lihat BUG-03")
	}
	if !containsMethod(allow, "DELETE") {
		t.Fatalf("BUG-04: DELETE terdaftar di router tetapi tidak diizinkan CORS: %q", allow)
	}
}

func containsMethod(header, method string) bool {
	for i := 0; i+len(method) <= len(header); i++ {
		if header[i:i+len(method)] == method {
			return true
		}
	}
	return false
}
