package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cresenity/gate/config"
	"github.com/gin-gonic/gin"
)

// jalankan Auth pada satu permintaan, kembalikan recorder dan apakah handler tercapai
func runAuth(header, query string) (*httptest.ResponseRecorder, bool) {
	gin.SetMode(gin.TestMode)

	reached := false
	r := gin.New()
	r.GET("/api/info", Auth, func(c *gin.Context) {
		reached = true
		c.String(http.StatusOK, "ok")
	})

	url := "/api/info"
	if query != "" {
		url += "?apiKey=" + query
	}
	req := httptest.NewRequest(http.MethodGet, url, nil)
	if header != "" {
		req.Header.Set("Authorization", header)
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	return w, reached
}

func TestAuthAcceptsValidBearerToken(t *testing.T) {
	config.AppConfig.ApiKey = "kunci-rahasia"

	w, reached := runAuth("Bearer kunci-rahasia", "")

	if w.Code != http.StatusOK || !reached {
		t.Fatalf("token benar seharusnya diterima, dapat status %d reached=%v", w.Code, reached)
	}
}

func TestAuthAcceptsValidQueryParam(t *testing.T) {
	config.AppConfig.ApiKey = "kunci-rahasia"

	w, reached := runAuth("", "kunci-rahasia")

	if w.Code != http.StatusOK || !reached {
		t.Fatalf("apiKey di query seharusnya diterima, dapat status %d reached=%v", w.Code, reached)
	}
}

func TestAuthRejectsWrongToken(t *testing.T) {
	config.AppConfig.ApiKey = "kunci-rahasia"

	w, reached := runAuth("Bearer salah", "")

	if w.Code != http.StatusUnauthorized || reached {
		t.Fatalf("token salah seharusnya 401, dapat status %d reached=%v", w.Code, reached)
	}
}

func TestAuthRejectsMissingCredential(t *testing.T) {
	config.AppConfig.ApiKey = "kunci-rahasia"

	w, reached := runAuth("", "")

	if w.Code != http.StatusUnauthorized || reached {
		t.Fatalf("tanpa kredensial seharusnya 401, dapat status %d reached=%v", w.Code, reached)
	}
}

// BUG-01. API_KEY yang tidak diisi membuat seluruh API terbuka tanpa kredensial:
// token kosong dibandingkan dengan ApiKey kosong menghasilkan sama, sehingga Auth meloloskannya.
func TestAuthMustRejectAnonymousWhenApiKeyUnset(t *testing.T) {
	config.AppConfig.ApiKey = ""

	w, reached := runAuth("", "")

	if w.Code != http.StatusUnauthorized || reached {
		t.Fatalf("BUG-01: API_KEY kosong meloloskan permintaan tanpa kredensial (status %d, handler tercapai=%v)", w.Code, reached)
	}
}

// BUG-02. Header tanpa awalan "Bearer " ikut diterima apa adanya, sehingga
// "Authorization: kunci-rahasia" sama sahnya dengan "Bearer kunci-rahasia".
func TestAuthShouldRequireBearerScheme(t *testing.T) {
	config.AppConfig.ApiKey = "kunci-rahasia"

	w, reached := runAuth("kunci-rahasia", "")

	if w.Code != http.StatusUnauthorized || reached {
		t.Fatalf("BUG-02: header tanpa skema Bearer ikut diterima (status %d, handler tercapai=%v)", w.Code, reached)
	}
}
