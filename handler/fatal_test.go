package handler

import (
	"net/http"
	"os"
	"os/exec"
	"testing"

	"github.com/gin-gonic/gin"
)

// jalankan satu test ini sendiri di subproses, supaya log.Fatal di dalamnya
// mematikan subprosesnya dan bukan proses test induk
func runInSubprocess(t *testing.T, nama string) (int, string) {
	t.Helper()

	cmd := exec.Command(os.Args[0], "-test.run=^"+nama+"$", "-test.v")
	cmd.Env = append(os.Environ(), "GATE_FATAL_HELPER=1")
	keluaran, err := cmd.CombinedOutput()

	kode := 0
	if ee, ok := err.(*exec.ExitError); ok {
		kode = ee.ExitCode()
	}

	return kode, string(keluaran)
}

func siapkanBerkasDomain(t *testing.T, nama string) bool {
	t.Helper()

	if err := os.MkdirAll("/etc/nginx/conf.d", 0755); err != nil {
		return false
	}
	isi := "server {\n\tlocation / {\n\t\tproxy_pass http://10.0.0.1;\n\t}\n}"

	return os.WriteFile(getPathDomain(nama), []byte(isi), 0644) == nil
}

// BUG-21. UpdateDomain memanggil log.Fatalf ketika "nginx -s reload" gagal. log.Fatal
// memanggil os.Exit(1), yang melewati Recovery milik gin, sehingga satu permintaan
// yang gagal memuat ulang nginx mematikan seluruh daemon gate. Ini jalur yang paling
// mudah terpicu di produksi: cukup konfigurasi nginx yang tidak sah, atau nginx yang
// sedang tidak berjalan.
func TestUpdateDomainMustNotExitProcess(t *testing.T) {
	nama := "uji-update.invalid"

	if os.Getenv("GATE_FATAL_HELPER") == "1" {
		if !siapkanBerkasDomain(t, nama) {
			t.Skip("dilewati: /etc/nginx/conf.d tidak bisa disiapkan")
		}
		callHandler(UpdateDomain, http.MethodPut, "/api/domain/"+nama+"/10.0.0.2",
			gin.Params{{Key: "domain", Value: nama}, {Key: "ipAddress", Value: "10.0.0.2"}})
		return
	}
	defer os.Remove(getPathDomain(nama))

	kode, keluaran := runInSubprocess(t, "TestUpdateDomainMustNotExitProcess")

	if kode == 1 {
		t.Fatalf("BUG-21: kegagalan reload nginx mematikan proses (exit 1) alih-alih menjawab galat.\nkeluaran subproses:\n%s", keluaran)
	}
}

// BUG-22. DeleteDomain memakai log.Fatalf pada jalur yang sama. Selain mematikan
// daemon, ia melakukannya setelah berkas konfigurasi dan sertifikat sudah terhapus,
// jadi prosesnya berhenti di tengah dengan keadaan yang tidak utuh.
func TestDeleteDomainMustNotExitProcess(t *testing.T) {
	nama := "uji-delete.invalid"

	if os.Getenv("GATE_FATAL_HELPER") == "1" {
		if !siapkanBerkasDomain(t, nama) {
			t.Skip("dilewati: /etc/nginx/conf.d tidak bisa disiapkan")
		}
		if err := os.MkdirAll("/etc/letsencrypt/renewal", 0755); err != nil {
			t.Skip("dilewati: jalur renewal tidak bisa disiapkan")
		}
		if err := os.WriteFile(getPathRenewalDomain(nama), []byte("x"), 0644); err != nil {
			t.Skip("dilewati: berkas renewal tidak bisa ditulis")
		}
		callHandler(DeleteDomain, http.MethodDelete, "/api/domain/"+nama,
			gin.Params{{Key: "domain", Value: nama}})
		return
	}
	defer os.Remove(getPathDomain(nama))
	defer os.Remove(getPathRenewalDomain(nama))

	kode, keluaran := runInSubprocess(t, "TestDeleteDomainMustNotExitProcess")

	if kode == 1 {
		t.Fatalf("BUG-22: kegagalan reload nginx mematikan proses (exit 1) setelah berkas terlanjur dihapus.\nkeluaran subproses:\n%s", keluaran)
	}
}
