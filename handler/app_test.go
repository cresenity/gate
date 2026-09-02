package handler

import (
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/cresenity/gate/config"
	"github.com/gin-gonic/gin"
)

// pindah ke direktori sementara, karena handler membaca dan menulis "data/config.json"
// relatif terhadap direktori kerja proses
func inTempDir(t *testing.T) {
	t.Helper()

	asal, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(asal) })
}

func TestGetInfoReturnsNameAndVersion(t *testing.T) {
	config.AppConfig.Name = "gate"
	config.AppConfig.Version = "1.0.0"

	w := callHandler(GetInfo, http.MethodGet, "/api/info", nil)

	body := decodeBody(t, w)
	data, ok := body["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("data bukan objek: %v", body)
	}
	if data["name"] != "gate" || data["version"] != "1.0.0" {
		t.Errorf("data = %v", data)
	}
}

func TestCreateDataConfigurationCreatesFile(t *testing.T) {
	inTempDir(t)

	if !createDataConfiguration() {
		t.Fatal("createDataConfiguration mengembalikan false")
	}
	if _, err := os.Stat("data/config.json"); err != nil {
		t.Fatalf("data/config.json tidak terbentuk: %v", err)
	}
}

// Penjaga. createDataConfiguration memakai os.Mkdir, bukan os.MkdirAll. Aman selama
// "data" hanya satu tingkat di bawah direktori kerja, dan test ini yang memastikannya
// tetap begitu — kalau jalurnya kelak dibuat bertingkat, os.Mkdir gagal dan log.Fatal
// akan mematikan proses.
func TestCreateDataConfigurationHandlesNestedWorkdir(t *testing.T) {
	inTempDir(t)

	dalam := filepath.Join("a", "b")
	if err := os.MkdirAll(dalam, 0755); err != nil {
		t.Fatalf("mkdirall: %v", err)
	}
	if err := os.Chdir(dalam); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	if !createDataConfiguration() {
		t.Fatal("createDataConfiguration mengembalikan false")
	}
}

func TestSetDefaultIpAddressPersistsValue(t *testing.T) {
	inTempDir(t)

	w := callHandler(SetDefaultIpAddress, http.MethodPost, "/api/config/ip/10.0.0.9",
		gin.Params{{Key: "ipAddress", Value: "10.0.0.9"}})

	body := decodeBody(t, w)
	if body["status"] != true {
		t.Fatalf("status = %v", body["status"])
	}

	isi, err := os.ReadFile("data/config.json")
	if err != nil {
		t.Fatalf("baca config: %v", err)
	}
	var simpan map[string]interface{}
	if err := json.Unmarshal(isi, &simpan); err != nil {
		t.Fatalf("config bukan JSON: %v", err)
	}
	if simpan["DefaultIp"] != "10.0.0.9" {
		t.Errorf("DefaultIp tersimpan = %v", simpan["DefaultIp"])
	}
}

// BUG-13. Nilai ipAddress tidak pernah divalidasi sebagai alamat IP. Apa pun yang
// lolos pencocokan rute akan tersimpan dan kelak dipakai menyusun direktif proxy_pass.
func TestSetDefaultIpAddressMustRejectInvalidIp(t *testing.T) {
	inTempDir(t)

	w := callHandler(SetDefaultIpAddress, http.MethodPost, "/api/config/ip/bukan-ip",
		gin.Params{{Key: "ipAddress", Value: "bukan-ip"}})

	body := decodeBody(t, w)
	if body["status"] == true {
		t.Fatalf("BUG-13: nilai %q diterima sebagai alamat IP", "bukan-ip")
	}
}

// BUG-14. Kegagalan menulis config.json dibuang dengan "_ =", sehingga handler tetap
// menjawab status sukses meski tidak ada yang tersimpan.
func TestSetDefaultIpAddressMustReportWriteFailure(t *testing.T) {
	inTempDir(t)

	if err := os.Mkdir("data", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// bikin config.json menjadi direktori supaya penulisan berkas pasti gagal
	if err := os.Mkdir(filepath.Join("data", "config.json"), 0755); err != nil {
		t.Fatalf("mkdir config.json: %v", err)
	}

	w := callHandler(SetDefaultIpAddress, http.MethodPost, "/api/config/ip/10.0.0.9",
		gin.Params{{Key: "ipAddress", Value: "10.0.0.9"}})

	body := decodeBody(t, w)
	if body["status"] == true {
		t.Fatalf("BUG-14: penulisan gagal tetapi handler menjawab sukses")
	}
}

// BUG-15. GetConfiguration memanggil log.Fatal ketika data/config.json tidak terbaca
// atau bukan JSON yang sah. log.Fatal memanggil os.Exit(1), yang melewati Recovery
// milik gin, sehingga satu permintaan bisa mematikan seluruh daemon.
//
// Dijalankan sebagai subproses karena kalau tidak, ia akan mematikan proses test ini.
func TestGetConfigurationMustNotExitProcess(t *testing.T) {
	if os.Getenv("GATE_FATAL_HELPER") == "1" {
		inTempDir(t)
		_ = os.Mkdir("data", 0755)
		_ = os.WriteFile("data/config.json", []byte("{rusak"), 0644)
		callHandler(GetConfiguration, http.MethodGet, "/api/config", nil)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestGetConfigurationMustNotExitProcess", "-test.v")
	cmd.Env = append(os.Environ(), "GATE_FATAL_HELPER=1")
	keluaran, err := cmd.CombinedOutput()

	if ee, ok := err.(*exec.ExitError); ok && ee.ExitCode() == 1 {
		t.Fatalf("BUG-15: config.json rusak membuat proses keluar dengan kode 1 (log.Fatal), bukan menjawab galat.\nkeluaran subproses:\n%s", keluaran)
	}
}
