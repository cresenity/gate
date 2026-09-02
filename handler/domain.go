package handler

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"log"
	"net/http"
	"net/url"
	"os"

	gincollector "github.com/cresenity/devcloud-collector-client-golang/gin"
	"github.com/cresenity/gate/config"
	dtf "github.com/cresenity/gate/datatransfer"
	"github.com/gin-gonic/gin"
)

const (
	filePath            string = "/etc/nginx/conf.d/"
	filePathCertificate        = "/etc/letsencrypt/live/"
	filePathRenewal            = "/etc/letsencrypt/renewal/"
	// filePath            string = "/usr/local/etc/nginx/servers/"
	fileTemplate string = `server {
		server_name www.%s;
		return 301 $scheme://%s$request_uri;

	}
	server {
		listen       80;
		listen  [::]:80;
		server_name  %s;

		location / {
			proxy_pass http://%s;
			proxy_set_header Host $http_host;
			proxy_set_header X-Forwarded-For $remote_addr;
			proxy_set_header X-Forwarded-Host $host;
			proxy_set_header X-Forwarded-Proto $scheme;
			proxy_set_header X-Forwarded-Server $host;
			proxy_set_header X-Forwarded-Port $server_port;

		}

		error_page   500 502 503 504  /50x.html;
		location = /50x.html {
			root   /usr/share/nginx/html;
		}
	}`
)

func InstallSsl(c *gin.Context) {
	errCode := 0
	errMessage := ""
	// var ipvalid bool

	nameDomain := c.Param("domain")
	ip := c.Param("ipAddress")
	if len(ip) == 0 {
		defaultIp, err := readDefaultIp()
		if err != nil {
			gincollector.Report(config.Collector, c, err)
			c.JSON(http.StatusInternalServerError, dtf.Response{
				Status:  false,
				Message: "Error read default ip: " + err.Error(),
			})
			return
		}
		ip = defaultIp
	}

	// BUG-08/BUG-13: ip dipakai mentah menyusun direktif nginx (proxy_pass) di
	// CreateDomain di bawah; tanpa validasi ini nilai apa pun yang lolos pencocokan
	// rute bisa menyuntik direktif nginx tambahan ke berkas vhost.
	if net.ParseIP(ip) == nil {
		c.JSON(http.StatusBadRequest, dtf.Response{
			Status:  false,
			Message: "Invalid IP address",
		})
		return
	}

	var isSSL bool
	var isConnectIp bool
	var isConnectIpWWW bool
	var isWWWStatusSSL bool

	//chcek domain ada atau tidak
	if errCode == 0 {
		validateDomain := validateDomain(nameDomain)
		if !validateDomain {
			errCode++
			errMessage = "Domain not valid"
		}
	}

	if errCode == 0 {
		targetDomain := nameDomain
		desiredIP := config.AppConfig.IP
		log.Println(" IP Config :", desiredIP)

		if len(desiredIP) > 0 {
			ips, err := net.LookupIP(targetDomain)
			if err != nil {
				// BUG-25: log.Panicln di sini membatalkan balasan JSON yang sudah
				// dibangun tepat di atasnya; gin Recovery menangkap panic-nya dan
				// menjawab 500 generik, bukan errMessage yang sebenarnya.
				errCode++
				errMessage = "Error lookup ip"
				log.Println("Error looking up IP for domain:", err)
				gincollector.Report(config.Collector, c, err)
			}

			for _, ip := range ips {
				log.Println("Error Cerbot IP:", ip.String())
				if ip.String() == desiredIP {
					isConnectIp = true
				}
			}

			if !isConnectIp {
				errCode++
				errMessage = "ip can't connect in default ip :" + desiredIP
			}
		}
	}

	if errCode == 0 {
		_, err := CreateDomain(nameDomain, ip)
		if err != nil {
			log.Println("Error creating file:", err)
			errMessage = "error create file"
			errCode++
		}
	}

	if errCode == 0 {
		targetDomainWWW := nameDomain
		desiredIP := config.AppConfig.IP
		log.Println(" IP Config :", desiredIP)

		if len(desiredIP) > 0 {
			ips, err := net.LookupIP(getNameDomainWWW(targetDomainWWW))

			if err == nil {
				for _, ip := range ips {
					log.Println("Error Cerbot IP:", ip.String())
					if ip.String() == desiredIP {
						isConnectIpWWW = true
					}
				}

				if isConnectIpWWW {
					cmdWWW := exec.Command("certbot", "--nginx", "-d", getNameDomainWWW(nameDomain), "--non-interactive", "--agree-tos", "-m", config.AppConfig.AdminEmail)
					_, err := cmdWWW.CombinedOutput()
					if err != nil {
						log.Println("Error Cerbot file:", err)
						errCode++
						errMessage = fmt.Sprintf("ERROR CERBOT: %s\n", err)
					}
					isWWWStatusSSL = checkCertificate(getNameDomainWWW(nameDomain))
				}
			}

		}
	}

	if errCode == 0 {
		cmd := exec.Command("certbot", "--nginx", "-d", nameDomain, "--non-interactive", "--agree-tos", "-m", config.AppConfig.AdminEmail)
		_, err := cmd.CombinedOutput()
		if err != nil {
			log.Println("Error Cerbot file:", err)
			errCode++
			errMessage = fmt.Sprintf("ERROR CERBOT: %s\n", err)
		}
	}

	if errCode == 0 {
		isSSL = checkCertificate(nameDomain)
	}

	c.JSON(
		http.StatusOK,
		dtf.Response{
			Status:  errCode == 0,
			Code:    errCode,
			Message: errMessage,
			Data: map[string]interface{}{
				"isConnectIp":  isConnectIp,
				"statusSSL":    isSSL,
				"wwwStatusSSL": isWWWStatusSSL,
			},
		},
	)
}

func UpdateDomain(c *gin.Context) {

	domain := c.Param("domain")
	ip := c.Param("ipAddress")

	var isSSL bool
	var errMessage string
	var errCode int

	// BUG-08/BUG-13: sama seperti InstallSsl, ip dipakai mentah menyusun proxy_pass
	// lewat regex di bawah.
	if net.ParseIP(ip) == nil {
		c.JSON(http.StatusBadRequest, dtf.Response{
			Status:  false,
			Message: "Invalid IP address",
		})
		return
	}

	filePath := getPathDomain(domain)
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		errCode++
		errMessage = "Domain not founds"
	}

	if errCode == 0 {
		content, err := ioutil.ReadFile(getPathDomain(domain))
		if err != nil {
			// BUG-15 (jalur serupa): log.Fatal di sini akan mematikan daemon,
			// sedangkan errCode/errMessage di atasnya sudah cukup untuk menjawab galat.
			errCode++
			errMessage = "Error Read File"
			log.Println("Error when opening file: ", err)
			gincollector.Report(config.Collector, c, err)
		} else {
			strContent := string(content)
			re := regexp.MustCompile(`(?m)proxy_pass\s+http://[^;]+`)
			newContent := re.ReplaceAllString(strContent, "proxy_pass http://"+ip)
			// Tulis kembali isi file
			if err := ioutil.WriteFile(getPathDomain(domain), []byte(newContent), 0644); err != nil {
				errCode++
				errMessage = "Error when writing file"
			}
		}
	}

	if errCode == 0 {
		cmd := exec.Command("nginx", "-s", "reload")
		err := cmd.Run()
		if err != nil {
			// BUG-21: log.Fatalf di sini memanggil os.Exit(1), melewati Recovery
			// gin dan mematikan seluruh daemon atas satu kegagalan reload nginx.
			// errMessage yang sudah dibangun tepat di atas ini sudah cukup untuk
			// menjawab galatnya lewat balasan JSON normal di bawah.
			errCode++
			errMessage = fmt.Sprintf("Failed run nginx command: %s", err)
			log.Println("Failed to run nginx command: ", err)
			gincollector.Report(config.Collector, c, err)
		}
	}

	if errCode == 0 {
		isSSL = checkCertificate(domain)
	}

	c.JSON(
		http.StatusOK,
		dtf.Response{
			Status:  errCode == 0,
			Code:    errCode,
			Message: errMessage,
			Data: map[string]interface{}{
				"ip":        ip,
				"statusSSL": isSSL,
			},
		},
	)
}

func DeleteDomain(c *gin.Context) {
	domain := c.Param("domain")

	errCode := 0
	errMessage := fmt.Sprintf("Success Delete Domain %s", domain)

	// Menentukan file yang akan dihapus
	filePath := getPathDomain(domain)

	// Check domain tersebut ada atau tidak
	_, errPath := os.Stat(filePath)
	if os.IsNotExist(errPath) {
		errCode++
		errMessage = "Domain not founds"
	}

	// Menjalankan perintah untuk menghapus file
	if errCode == 0 {
		err := os.Remove(filePath)
		if err != nil {
			// Menangkap error jika ada
			errCode++
			errMessage = fmt.Sprintf("Error: %s", err)

		}
	}

	// Menjalankan Perintah untuk menghapus file ssl
	if errCode == 0 {
		folderPath := getPathCertificateDomain(domain)
		err := os.RemoveAll(folderPath)
		if err != nil {
			errCode++
			errMessage = "Error Delete Path Certificate"
		}
		wwwFolderPath := getPathCertificateDomain("www." + domain)
		if _, err := os.Stat(wwwFolderPath); err == nil {
			err := os.RemoveAll(wwwFolderPath)
			if err != nil {
				errCode++
				errMessage = "Error Delete Path WWW Certificate"
			}
		}
	}

	// Menjalankan Perintah untuk menghapus file ssl
	if errCode == 0 {
		folderPath := getPathRenewalDomain(domain)
		err := os.Remove(folderPath)
		if err != nil {
			errCode++
			errMessage = "Error Delete Path Renewal"
		}
		wwwFolderPath := getPathRenewalDomain("www." + domain)
		if _, err := os.Stat(wwwFolderPath); err == nil {
			err := os.Remove(wwwFolderPath)
			if err != nil {
				errCode++
				errMessage = "Error Delete Path WWW Renewal"
			}
		}
	}

	if errCode == 0 {
		cmd := exec.Command("nginx", "-s", "reload")
		err := cmd.Run()
		if err != nil {
			// BUG-22: log.Fatalf di sini mematikan daemon setelah berkas konfigurasi
			// dan sertifikat sudah terhapus, jadi pemanggil tidak pernah tahu apa
			// yang sempat terjadi. errMessage sudah dibangun di atas, cukup dijawab.
			errCode++
			errMessage = fmt.Sprintf("Failed run nginx command: %s", err)
			log.Println("Failed to run nginx command: ", err)
			gincollector.Report(config.Collector, c, err)
		}
	}

	c.JSON(
		http.StatusOK,
		dtf.Response{
			Status:  errCode == 0,
			Code:    errCode,
			Message: errMessage,
		},
	)
}

func GetDomainStatus(c *gin.Context) {
	domain := c.Param("domain")

	errCode := 0
	errMessage := ""

	var ipDomain []string
	var isSSL bool
	var isSSLWWW bool
	var isConnectIpWWW bool

	filePath := getPathDomain(domain)
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		errCode++
		errMessage = "Domain not found"
	}

	if errCode == 0 {
		// BUG-10: sebelumnya "return" langsung di sini membuat handler tidak
		// menjawab apa pun (200 berbadan kosong) ketika DNS gagal, berbeda dari
		// seluruh jalur galat lain yang menjawab JSON.
		ips, err := net.LookupHost(domain)
		if err != nil {
			log.Println("Error when looking up host:", err)
			errCode++
			errMessage = "Error lookup host"
		} else {
			ipDomain = append(ipDomain, ips...)
		}
	}

	if errCode == 0 {
		targetDomainWWW := domain
		desiredIP := config.AppConfig.IP
		log.Println(" IP Config :", desiredIP)

		if len(desiredIP) > 0 {
			ips, err := net.LookupIP(getNameDomainWWW(targetDomainWWW))
			if err == nil {
				for _, ip := range ips {
					log.Println("Error Cerbot IP:", ip.String())
					if ip.String() == desiredIP {
						isConnectIpWWW = true
					}
				}

				if isConnectIpWWW {
					isSSLWWW = checkCertificate(getNameDomainWWW(targetDomainWWW))
				}
			}
		}
	}

	if errCode == 0 {
		isSSL = checkCertificate(domain)
	}

	c.JSON(
		http.StatusOK,
		dtf.Response{
			Status:  errCode == 0,
			Code:    errCode,
			Message: errMessage,
			Data: map[string]interface{}{
				"ip":           ipDomain,
				"statusSSL":    isSSL,
				"statusSSLWWW": isSSLWWW,
			},
		},
	)
}

// BUG-11: sebelumnya total dihitung dari exec.Command("ls", ...) lalu
// len(strings.Split(output, "\n")) — keluaran ls selalu berakhir dengan baris baru
// sehingga menghasilkan satu elemen kosong ekstra, total pun selalu satu lebih
// banyak daripada items. os.ReadDir tidak memerlukan proses baru, parsing teks, atau
// koreksi off-by-one.
func GetAllDomainStatus(c *gin.Context) {
	var errCode int
	var errMessage string
	var dataDomain []map[string]interface{}

	entries, err := os.ReadDir(filePath)
	if err != nil {
		errCode++
		errMessage = fmt.Sprintf("Error : %s", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".conf")
		if len(name) == 0 {
			continue
		}
		ips, _ := net.LookupHost(name)
		var ipDomain []string
		ipDomain = append(ipDomain, ips...)
		dataDomain = append(dataDomain, map[string]interface{}{
			"domain": name,
			"ip":     ipDomain,
		})
	}

	data := map[string]interface{}{
		"total": len(dataDomain),
		"items": dataDomain,
	}

	c.JSON(
		http.StatusOK,
		dtf.Response{
			Status:  errCode == 0,
			Code:    errCode,
			Message: errMessage,
			Data:    data,
		},
	)
}

// BUG-09: nama domain dulu tidak dibatasi sama sekali sebelum dirangkai jadi jalur
// berkas lewat Sprintf, sehingga "../" ikut apa adanya ke dalam jalur sasaran yang
// dipakai os.Create/os.Remove/ioutil.WriteFile (dan lewat getPathCertificateDomain
// juga os.RemoveAll). filepath.Base membuang komponen direktori apa pun di dalamnya.
func getPathDomain(name string) string {
	name = filepath.Base(name)
	if name == "." || name == ".." || name == string(os.PathSeparator) {
		name = ""
	}
	return fmt.Sprintf(filePath+"%s.conf", name)
}

func getNameDomainWWW(name string) string {
	return fmt.Sprintf("www.%s", name)
}

func getPathCertificateDomain(name string) string {
	return fmt.Sprintf(filePathCertificate+"%s", name)
}

func getPathRenewalDomain(name string) string {
	return fmt.Sprintf(filePathRenewal+"%s", name)
}

// BUG-08: ip dulu masuk mentah ke template. net.ParseIP menolak nilai yang memuat
// ";" atau baris baru (dipakai untuk menyuntik direktif nginx tambahan) sebelum ikut
// dirangkai; jalur pemanggil (InstallSsl/UpdateDomain) sudah menolaknya lebih awal,
// ini lapis pertahanan kedua langsung pada pembuatan template.
func getTemplateFile(name, ip string) string {
	if net.ParseIP(ip) == nil {
		ip = "127.0.0.1"
	}
	return fmt.Sprintf(fileTemplate, name, name, name, ip)
}

func CreateDomain(name string, ip string) (*os.File, error) {
	file, err := os.Create(getPathDomain(name))
	if err != nil {
		panic(err)
	}
	defer file.Close()

	if _, err := file.WriteString(getTemplateFile(name, ip)); err != nil {
		panic(err)
	}
	return file, err
}

func validateDomain(domain string) bool {
	var validateDomain bool
	validateDomain = true
	_, err := net.LookupHost(domain)
	if err != nil {
		validateDomain = false
	}

	_, err = url.Parse("http://" + domain)
	if err != nil {
		validateDomain = false
	}

	return validateDomain
}

// BUG-24: tls.Dial tanpa batas waktu bisa menggantung selamanya kalau lawan bicara
// tidak pernah membalas jabat tangan TLS, menghabiskan kapasitas server pada domain
// yang macet. 5 detik cukup untuk jabat tangan TLS normal tanpa membuat pengecekan
// status jadi lambat pada domain yang memang tidak merespons.
func checkCertificate(domain string) bool {
	ssl := true
	domainWithPort := domain + ":443"
	dialer := &net.Dialer{Timeout: 5 * time.Second}
	conn, err := tls.DialWithDialer(dialer, "tcp", domainWithPort, nil)
	if err != nil {
		log.Println("Error: ", err)
		return false
	}
	defer conn.Close()

	state := conn.ConnectionState()
	certs := state.PeerCertificates
	if len(certs) == 0 {
		ssl = false
	}
	return ssl
}

// readDefaultIp membaca DefaultIp dari data/config.json lewat readConfiguration
// (handler/app.go), yang sudah tidak lagi log.Fatal pada berkas rusak (BUG-15).
func readDefaultIp() (string, error) {
	data, err := readConfiguration()
	if err != nil {
		return "", err
	}
	return data.DefaultIp, nil
}
