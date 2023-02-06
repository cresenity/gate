package handler

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os/exec"
	"regexp"
	"strings"

	// "net"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/cresenity/gate/config"
	dtf "github.com/cresenity/gate/datatransfer"
	"github.com/gin-gonic/gin"
)

const (
	filePath            string = "/etc/nginx/conf.d/"
	filePathCertificate        = "/etc/letsencrypt/live/"
	// filePath            string = "/usr/local/etc/nginx/servers/"
	fileTemplate string = `server {
		listen       80;
		listen  [::]:80;
		server_name  %s;
	
		location / {
			proxy_pass http://%s;
			proxy_set_header Host $http_host;
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
		ip = getIpDefault()
	}

	var isSSL bool
	var isConnectIp bool

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
				errCode++
				errMessage = "Error lookup ip"
				log.Panicln("Error looking up IP for domain:", err)
			}

			for _, ip := range ips {
				log.Println("Error Cerbot IP:", ip.String())
				if ip.String() == desiredIP {
					isConnectIp = true
				}
			}

			if !isConnectIp {
				errCode++
				errMessage = "ip can't connect in defualt ip :" + desiredIP
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
		cmd := exec.Command("certbot", "--nginx", "-d", nameDomain)
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
			Status:  true,
			Code:    errCode,
			Message: errMessage,
			Data: map[string]interface{}{
				"isConnectIp": isConnectIp,
				"statusSSL":   isSSL,
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

	filePath := getPathDomain(domain)
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		errCode++
		errMessage = "Domain not founds"
	}

	if errCode == 0 {
		content, err := ioutil.ReadFile(getPathDomain(domain))
		if err != nil {
			errCode++
			errMessage = "Error Read File"
			log.Fatal("Error when opening file: ", err)
		}

		strContent := string(content)
		re := regexp.MustCompile(`(?m)proxy_pass\s+http://[^;]+`)
		newContent := re.ReplaceAllString(strContent, "proxy_pass http://"+ip)
		// Tulis kembali isi file
		err = ioutil.WriteFile(getPathDomain(domain), []byte(newContent), 0644)
		if err != nil {
			errCode++
			errMessage = "Error when writing file"
		}
	}

	if errCode == 0 {
		cmd := exec.Command("nginx", "-s", "reload")
		err := cmd.Run()
		if err != nil {
			errCode++
			errMessage = fmt.Sprintf("Failed run nginx command: %s", err)
			log.Fatalf("Failed to run nginx command: %s", err)
		}
	}

	if errCode == 0 {
		isSSL = checkCertificate(domain)
	}

	c.JSON(
		http.StatusOK,
		dtf.Response{
			Status:  true,
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

	// Menjalankan perintah untuk menghapus file
	err := os.Remove(filePath)
	if err != nil {
		// Menangkap error jika ada
		errCode++
		errMessage = fmt.Sprintf("Error: %s", err)

	}

	c.JSON(
		http.StatusOK,
		dtf.Response{
			Status:  true,
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

	filePath := getPathDomain(domain)
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		errCode++
		errMessage = "Domain not founds"
	}

	if errCode == 0 {
		ips, err := net.LookupHost(domain)
		if err != nil {
			fmt.Println("Error when looking up host:", err)
			return
		}
		for _, ip := range ips {
			ipDomain = append(ipDomain, ip)
		}

	}

	if errCode == 0 {
		isSSL = checkCertificate(domain)
	}

	c.JSON(
		http.StatusOK,
		dtf.Response{
			Status:  true,
			Code:    errCode,
			Message: errMessage,
			Data: map[string]interface{}{
				"ip":        ipDomain,
				"statusSSL": isSSL,
			},
		},
	)
}

func GetAllDomainStatus(c *gin.Context) {
	var errCode int
	var errMessage string
	var dataDomain []map[string]interface{}
	var domains []string

	cmd := exec.Command("ls", filePath)

	// Mengambil output perintah sebagai bytes.Buffer
	var out bytes.Buffer
	cmd.Stdout = &out

	// Menjalankan perintah
	err := cmd.Run()
	if err != nil {
		errCode++
		errMessage = fmt.Sprintf("Error : %s", err)
	}

	// Mengubah output perintah menjadi string
	str := out.String()

	// str := out.String()
	domains = strings.Split(str, "\n")
	countDomain := len(domains)

	for _, v := range domains {
		var ipDomain []string
		name := strings.Replace(v, ".conf", "", -1)
		if len(name) > 0 {
			// isSSL := checkCertificate(name)
			ips, _ := net.LookupHost(name)
			for _, ip := range ips {
				ipDomain = append(ipDomain, ip)
			}
			domain := map[string]interface{}{
				"domain": name,
				"ip":     ipDomain,
				// "statusSSL": false,
			}
			dataDomain = append(dataDomain, domain)
		}

	}

	data := map[string]interface{}{
		"total": countDomain,
		"items": dataDomain,
	}

	c.JSON(
		http.StatusOK,
		dtf.Response{
			Status:  true,
			Code:    errCode,
			Message: errMessage,
			Data:    data,
		},
	)
}

func getPathDomain(name string) string {
	return fmt.Sprintf(filePath+"%s.conf", name)
}

func getPathCertificateDomain(name string) string {
	return fmt.Sprintf(filePathCertificate+"%s", name)
}

func getTemplateFile(name, ip string) string {
	return fmt.Sprintf(fileTemplate, name, ip)
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

func checkCertificate(domain string) bool {
	ssl := true
	domainWithPort := domain + ":443"
	conn, err := tls.Dial("tcp", domainWithPort, nil)
	if err != nil {
		log.Println("Error: ", err)
	}
	defer conn.Close()

	state := conn.ConnectionState()
	certs := state.PeerCertificates
	if len(certs) == 0 {
		ssl = false
	}
	return ssl
}

// get IP pada file data/config.json
func getIpDefault() string {
	filename := "data/config.json"
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		createDataConfiguration()
	}
	var data dtf.Configuration
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal("Error when opening file: ", err)
	}
	err = json.Unmarshal(content, &data)
	if err != nil {
		log.Fatal("Error during Unmarshal(): ", err)
	}
	return data.DefaultIp
}
