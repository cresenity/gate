package handler

import (
	"net"
	"net/http"
	"path/filepath"
	"sync"

	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/cresenity/gate/config"
	dtf "github.com/cresenity/gate/datatransfer"
	"github.com/gin-gonic/gin"
)

// BUG-28: berkas konfigurasi dulu selalu memakai jalur relatif literal "data/config.json",
// jadi lokasinya berpindah mengikuti direktori kerja proses. GATE_DATA_DIR memungkinkan
// dep deployment memilih jalur absolut yang tetap; kosong tetap memakai "data" relatif
// seperti sebelumnya.
func dataDir() string {
	if dir := os.Getenv("GATE_DATA_DIR"); dir != "" {
		return dir
	}
	return "data"
}

func configFilePath() string {
	return filepath.Join(dataDir(), "config.json")
}

// BUG-26: data/config.json dulu dibaca/ditulis tanpa penguncian ataupun penulisan
// atomik. configMu menutup celah antar-goroutine di proses yang sama (gate berjalan
// sebagai satu instance per container, jadi mutex in-process cukup).
var configMu sync.Mutex

func GetInfo(c *gin.Context) {
	c.JSON(
		http.StatusOK,
		dtf.Response{
			Status: true,
			Data: dtf.AppInfo{
				Name:    config.AppConfig.Name,
				Version: config.AppConfig.Version,
			},
		},
	)
}

func GetConfiguration(c *gin.Context) {
	data, err := readConfiguration()
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			dtf.Response{
				Status:  false,
				Message: "Error read configuration: " + err.Error(),
			},
		)
		return
	}

	c.JSON(
		http.StatusOK,
		dtf.Response{
			Status: true,
			Data:   data,
		},
	)
}

func SetDefaultIpAddress(c *gin.Context) {
	createDataConfiguration()
	ip := c.Param("ipAddress")

	// BUG-13: ip default tidak pernah divalidasi sebagai alamat IP sebelum disimpan.
	if net.ParseIP(ip) == nil {
		c.JSON(
			http.StatusBadRequest,
			dtf.Response{
				Status:  false,
				Message: "Invalid IP address",
			},
		)
		return
	}

	data := dtf.Configuration{
		DefaultIp: ip,
	}

	// BUG-14: kegagalan menulis config.json dulu dibuang dengan "_ =", sehingga
	// handler tetap menjawab sukses walau tidak ada yang benar-benar tersimpan.
	if err := writeConfiguration(data); err != nil {
		c.JSON(
			http.StatusInternalServerError,
			dtf.Response{
				Status:  false,
				Message: "Error write configuration: " + err.Error(),
			},
		)
		return
	}

	c.JSON(
		http.StatusOK,
		dtf.Response{
			Status: true,
			Data:   data,
		},
	)
}

// readConfiguration membaca data/config.json. BUG-15: sebelumnya galat baca/parse di sini
// (dipakai juga oleh GetConfiguration dan getIpDefault) langsung log.Fatal, mematikan
// seluruh daemon atas satu permintaan yang menyentuh berkas rusak.
func readConfiguration() (dtf.Configuration, error) {
	configMu.Lock()
	defer configMu.Unlock()

	createDataConfiguration()

	var data dtf.Configuration
	content, err := ioutil.ReadFile(configFilePath())
	if err != nil {
		return data, err
	}
	if err := json.Unmarshal(content, &data); err != nil {
		return data, err
	}
	return data, nil
}

// writeConfiguration menulis data/config.json secara atomik (tulis ke berkas sementara lalu
// rename) supaya proses lain yang membacanya di tengah penulisan tidak pernah melihat isi
// yang terpotong (bagian dari perbaikan BUG-26).
func writeConfiguration(data dtf.Configuration) error {
	configMu.Lock()
	defer configMu.Unlock()

	file, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		return err
	}

	tmp := configFilePath() + ".tmp"
	if err := ioutil.WriteFile(tmp, file, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, configFilePath())
}

func createDataConfiguration() bool {
	dir := dataDir()
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatal("Error create directory data: ", err)
		}
	}

	filename := configFilePath()
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		data := dtf.Configuration{}
		file, _ := json.MarshalIndent(data, "", " ")
		if err := ioutil.WriteFile(filename, file, 0644); err != nil {
			log.Fatalf("Failed to create file: %s", err)
		}
	}
	return true
}
