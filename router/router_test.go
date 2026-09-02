package router

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// build router sambil menangkap panic, supaya kegagalan pendaftaran rute
// terbaca sebagai kegagalan test dan bukan menghentikan seluruh berkas.
func buildRouter() (r *gin.Engine, panicked interface{}) {
	gin.SetMode(gin.TestMode)
	defer func() { panicked = recover() }()
	r = InitializeRouter()
	return
}

// Penjaga. GET "status/:domain" dan GET "status/all" berada pada posisi yang sama di
// pohon rute. Diuji karena versi gin tertentu menolak segmen statis berdampingan dengan
// wildcard; gin 1.8.2 menerimanya. Test ini menangkapnya bila dependensi dinaikkan.
func TestInitializeRouterMustNotPanic(t *testing.T) {
	r, panicked := buildRouter()

	if panicked != nil {
		t.Fatalf("InitializeRouter() panic saat mendaftarkan rute: %v", panicked)
	}
	if r == nil {
		t.Fatal("InitializeRouter() mengembalikan nil")
	}
}

func TestAllExpectedRoutesRegistered(t *testing.T) {
	r, panicked := buildRouter()
	if panicked != nil {
		t.Skipf("dilewati: router gagal dibangun (%v)", panicked)
	}

	want := []string{
		"GET /api/info",
		"GET /api/config",
		"POST /api/config/ip/:ipAddress",
		"POST /api/domain/:domain/:ipAddress",
		"PUT /api/domain/:domain/:ipAddress",
		"DELETE /api/domain/:domain",
		"GET /api/domain/status/:domain",
		"GET /api/domain/status/all",
	}

	got := map[string]bool{}
	for _, ri := range r.Routes() {
		got[fmt.Sprintf("%s %s", ri.Method, ri.Path)] = true
	}

	for _, w := range want {
		if !got[w] {
			t.Errorf("rute tidak terdaftar: %s", w)
		}
	}
}

// BUG-06. router.GET("/") didaftarkan tanpa handler sama sekali. Rantai handler
// terisi Logger+Recovery dari gin.Default(), jadi "/" menjawab 200 berbadan kosong
// alih-alih 404 seperti jalur yang memang tidak ada.
func TestRootRouteShouldNotBeRegisteredEmpty(t *testing.T) {
	r, panicked := buildRouter()
	if panicked != nil {
		t.Skipf("dilewati: router gagal dibangun (%v)", panicked)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code == http.StatusOK && w.Body.Len() == 0 {
		t.Fatalf("BUG-06: GET / menjawab 200 dengan badan kosong (rute tanpa handler)")
	}
}

// Preflight CORS tidak akan pernah berhasil: OPTIONS tidak didaftarkan, dan Auth
// berjalan sebelum CORS sehingga preflight tanpa Authorization ditolak lebih dulu.
func TestPreflightRequestIsHandled(t *testing.T) {
	r, panicked := buildRouter()
	if panicked != nil {
		t.Skipf("dilewati: router gagal dibangun (%v)", panicked)
	}

	req := httptest.NewRequest(http.MethodOptions, "/api/info", nil)
	req.Header.Set("Origin", "https://devcloud.cresenity.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code == http.StatusUnauthorized || w.Code == http.StatusNotFound {
		t.Fatalf("BUG-07: preflight OPTIONS dijawab %d; Auth berjalan sebelum CORS dan OPTIONS tidak didaftarkan", w.Code)
	}
}
