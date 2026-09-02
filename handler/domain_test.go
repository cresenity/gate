package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// panggil satu handler dengan parameter rute yang sudah terisi
func callHandler(h gin.HandlerFunc, method, path string, params gin.Params) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, path, nil)
	c.Params = params
	h(c)

	return w
}

func decodeBody(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()

	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("balasan bukan JSON yang sah (status %d, badan %q): %v", w.Code, w.Body.String(), err)
	}

	return body
}

func TestGetPathDomain(t *testing.T) {
	if got := getPathDomain("example.com"); got != "/etc/nginx/conf.d/example.com.conf" {
		t.Errorf("getPathDomain = %q", got)
	}
}

func TestGetNameDomainWWW(t *testing.T) {
	if got := getNameDomainWWW("example.com"); got != "www.example.com" {
		t.Errorf("getNameDomainWWW = %q", got)
	}
}

func TestGetPathCertificateDomain(t *testing.T) {
	if got := getPathCertificateDomain("example.com"); got != "/etc/letsencrypt/live/example.com" {
		t.Errorf("getPathCertificateDomain = %q", got)
	}
}

func TestGetPathRenewalDomain(t *testing.T) {
	if got := getPathRenewalDomain("example.com"); got != "/etc/letsencrypt/renewal/example.com" {
		t.Errorf("getPathRenewalDomain = %q", got)
	}
}

func TestTemplateFileContainsDomainAndUpstream(t *testing.T) {
	out := getTemplateFile("example.com", "10.0.0.1")

	if !strings.Contains(out, "server_name  example.com") {
		t.Errorf("server_name tidak ada pada template:\n%s", out)
	}
	if !strings.Contains(out, "proxy_pass http://10.0.0.1") {
		t.Errorf("proxy_pass tidak ada pada template:\n%s", out)
	}
}

// BUG-08. Parameter ipAddress masuk mentah ke berkas konfigurasi nginx. Nilai yang
// memuat ";" dan baris baru (bisa lewat %3B dan %0A pada jalur URL) menyisipkan
// direktif nginx tambahan ke dalam berkas yang dihasilkan.
func TestTemplateFileMustNotAllowNginxInjection(t *testing.T) {
	jahat := "10.0.0.1;\n\t}\n\tserver {\n\t\tlisten 8080;\n\t\troot /etc;\n\t}\n\tserver {"

	out := getTemplateFile("example.com", jahat)

	if strings.Contains(out, "listen 8080") {
		t.Fatalf("BUG-08: ipAddress menyisipkan direktif nginx tambahan:\n%s", out)
	}
}

// BUG-09. Nama domain tidak pernah dibatasi pada karakter yang sah. Jalur berkas
// dirangkai lewat Sprintf sehingga "../" ikut apa adanya ke dalam jalur sasaran.
func TestGetPathDomainMustRejectTraversal(t *testing.T) {
	got := getPathDomain("../../etc/passwd")

	if strings.Contains(got, "..") {
		t.Fatalf("BUG-09: jalur berkas memuat traversal: %q", got)
	}
}

func TestGetDomainStatusRespondsForUnknownDomain(t *testing.T) {
	w := callHandler(GetDomainStatus, http.MethodGet, "/api/domain/status/tidak-ada.invalid",
		gin.Params{{Key: "domain", Value: "tidak-ada.invalid"}})

	body := decodeBody(t, w)
	if body["status"] != false {
		t.Errorf("domain tidak dikenal seharusnya status=false, dapat %v", body["status"])
	}
	if body["message"] != "Domain not found" {
		t.Errorf("pesan = %v", body["message"])
	}
}

// BUG-10. Bila berkas domain ada tetapi net.LookupHost gagal, handler melakukan
// "return" tanpa menulis balasan apa pun sehingga klien menerima 200 berbadan kosong,
// berbeda dari seluruh jalur galat lain yang menjawab JSON.
func TestGetDomainStatusAlwaysWritesResponse(t *testing.T) {
	nama := "domain-tidak-mungkin-resolve.invalid"
	berkas := getPathDomain(nama)

	if err := os.MkdirAll("/etc/nginx/conf.d", 0755); err != nil {
		t.Skipf("dilewati: /etc/nginx/conf.d tidak bisa dibuat (%v)", err)
	}
	if err := os.WriteFile(berkas, []byte("server {}"), 0644); err != nil {
		t.Skipf("dilewati: tidak bisa menulis %s (%v)", berkas, err)
	}
	defer os.Remove(berkas)

	w := callHandler(GetDomainStatus, http.MethodGet, "/api/domain/status/"+nama,
		gin.Params{{Key: "domain", Value: nama}})

	if w.Body.Len() == 0 {
		t.Fatalf("BUG-10: DNS gagal membuat handler kembali tanpa menulis balasan (status %d, badan kosong)", w.Code)
	}
}

// BUG-11. total dihitung dari len(strings.Split(output_ls, "\n")), dan keluaran ls
// selalu berakhir dengan baris baru sehingga menghasilkan satu elemen kosong.
// Akibatnya total selalu satu lebih banyak daripada jumlah domain yang sebenarnya.
func TestGetAllDomainStatusTotalMatchesItems(t *testing.T) {
	w := callHandler(GetAllDomainStatus, http.MethodGet, "/api/domain/status/all", nil)

	body := decodeBody(t, w)
	data, ok := body["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("data bukan objek: %v", body)
	}

	total, _ := data["total"].(float64)
	items, _ := data["items"].([]interface{})

	if int(total) != len(items) {
		t.Fatalf("BUG-11: total=%d tetapi items berisi %d entri", int(total), len(items))
	}
}
