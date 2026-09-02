package config

import (
	collector "github.com/cresenity/devcloud-collector-client-golang"
)

// Collector satu instance per proses, dipakai router dan handler untuk melaporkan
// galat ke devcloud. AppCode dipatok di kode karena itu identitas biner ini, bukan
// sesuatu yang perlu diatur ulang per deployment; semua yang lain (aktif/tidaknya,
// DocRoot, dst.) sengaja tidak diisi di sini supaya variabel lingkungan
// DEVCLOUD_COLLECTOR_* yang memutuskan - lihat README devcloud-collector-client-golang.
var Collector *collector.Collector

// LoadCollector membangun Collector dari env DEVCLOUD_COLLECTOR_*. Dipanggil sekali
// saat proses start, sesudah LoadAppConfig.
func LoadCollector() {
	Collector = collector.New(collector.Config{
		AppCode: "gate",
	})
}
