package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/spf13/viper"
)

var AppConfig Config

type Config struct {
	Name        string
	Port        int
	Environment string
	Debug       bool
	Version     string
	ApiKey      string
	IP          string
	AdminEmail  string
}

const defaultPort = 6000

func LoadAppConfig() {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("/")
	viper.AllowEmptyEnv(true)
	viper.AutomaticEnv()
	_ = viper.ReadInConfig()

	/*test*/

	AppConfig.Name = viper.GetString("APP_NAME")
	AppConfig.Version = viper.GetString("APP_VERSION")
	AppConfig.Port = viper.GetInt("PORT")
	AppConfig.Environment = viper.GetString("ENVIRONMENT")
	AppConfig.Debug = viper.GetBool("DEBUG")
	AppConfig.ApiKey = viper.GetString("API_KEY")
	AppConfig.IP = viper.GetString("IP")
	AppConfig.AdminEmail = viper.GetString("ADMIN_EMAIL")

	// BUG-23: DEBUG=true tidak boleh ikut berjalan di lingkungan production, apa pun
	// isi env var DEBUG-nya.
	if strings.EqualFold(AppConfig.Environment, "production") {
		AppConfig.Debug = false
	}

	// BUG-19/BUG-01: API_KEY kosong dulu diterima diam-diam dan membuka seluruh API
	// tanpa autentikasi. Gagal aman: pakai kunci acak sekali-jalan alih-alih menolak
	// start, supaya satu instance yang lupa di-set env tetap tertutup dari luar.
	if AppConfig.ApiKey == "" {
		AppConfig.ApiKey = generateRandomApiKey()
		log.Printf("[INIT] API_KEY kosong, memakai kunci acak untuk proses ini: %s", AppConfig.ApiKey)
	}

	// BUG-20: PORT kosong menghasilkan 0, membuat layanan mendengarkan di porta acak.
	if AppConfig.Port == 0 {
		AppConfig.Port = defaultPort
		log.Printf("[INIT] PORT kosong, memakai porta default %d", defaultPort)
	}

	log.Println("[INIT] configuration loaded")
}

func generateRandomApiKey() string {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("insecure-fallback-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(buf)
}
