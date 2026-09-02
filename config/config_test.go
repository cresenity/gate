package config

import (
	"os"
	"testing"
)

func TestLoadAppConfigReadsEnvironment(t *testing.T) {
	t.Setenv("APP_NAME", "gate-test")
	t.Setenv("APP_VERSION", "9.9.9")
	t.Setenv("PORT", "8081")
	t.Setenv("DEBUG", "true")
	t.Setenv("API_KEY", "kunci-uji")
	t.Setenv("IP", "10.0.0.1")
	t.Setenv("ADMIN_EMAIL", "ops@cresenity.com")

	LoadAppConfig()

	if AppConfig.Name != "gate-test" {
		t.Errorf("Name = %q", AppConfig.Name)
	}
	if AppConfig.Port != 8081 {
		t.Errorf("Port = %d", AppConfig.Port)
	}
	if !AppConfig.Debug {
		t.Errorf("Debug = %v", AppConfig.Debug)
	}
	if AppConfig.ApiKey != "kunci-uji" {
		t.Errorf("ApiKey = %q", AppConfig.ApiKey)
	}
}

// BUG-19. Konfigurasi yang menentukan keamanan tidak pernah divalidasi. API_KEY
// yang kosong diterima diam-diam, dan akibatnya terlihat pada BUG-01: seluruh API
// terbuka tanpa kredensial. LoadAppConfig seharusnya menolak berjalan.
func TestLoadAppConfigMustRejectEmptyApiKey(t *testing.T) {
	t.Setenv("API_KEY", "")

	LoadAppConfig()

	if AppConfig.ApiKey == "" {
		t.Fatalf("BUG-19: API_KEY kosong diterima tanpa peringatan; lihat BUG-01")
	}
}

// BUG-20. PORT yang tidak diisi menghasilkan 0, dan main.go akan mendengarkan pada
// ":0" — kernel memilih porta acak, sehingga layanan hidup di porta yang tidak
// diketahui siapa pun alih-alih gagal dengan jelas.
func TestLoadAppConfigMustRejectZeroPort(t *testing.T) {
	os.Unsetenv("PORT")
	t.Setenv("PORT", "")

	LoadAppConfig()

	if AppConfig.Port == 0 {
		t.Fatalf("BUG-20: PORT kosong menghasilkan 0, main.go akan mendengarkan di porta acak")
	}
}
