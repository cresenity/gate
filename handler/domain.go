package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os/exec"

	// "net"
	"log"
	"net/http"
	"os"

	dtf "github.com/cresenity/gate/datatransfer"
	"github.com/gin-gonic/gin"
)

const (
	filePath string = "/etc/nginx/conf.d/"
	// filePath     string = "/usr/local/etc/nginx/servers/"
	fileTemplate string = `server {
		listen       80;
		listen  [::]:80;
		server_name  %s;
	
		location / {
			proxy_pass http://%s;
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
		ip = "127.0.0.1"
	}

	// //TODO : check domain target ipAddress
	if errCode == 0 {
		createDataConfiguration()
		//create dir data
		var data dtf.Configuration
		content, err := ioutil.ReadFile("data/config.json")
		if err != nil {
			log.Fatal("Error when opening file: ", err)
		}
		err = json.Unmarshal(content, &data)
		if err != nil {
			log.Fatal("Error during Unmarshal(): ", err)
		}

		var isConnectIp bool

		targetDomain := nameDomain
		desiredIP := data.DefaultIp
		log.Println(" IP DEFAULT :", desiredIP)

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
		//TODO : add domain to nginx configuration
		_, err := CreateDomain(nameDomain, ip)
		if err != nil {
			log.Println("Error creating file:", err)
			errMessage = "error create file"
			errCode++
		}
	}

	if errCode == 0 {
		//TODO : run `certbot --nginx -d DOMAIN`
		cmd := exec.Command("certbot", "--nginx", "-d", nameDomain)
		_, err := cmd.CombinedOutput()
		if err != nil {
			log.Println("Error Cerbot file:", err)
			errCode++
			errMessage = fmt.Sprintf("ERROR CERBOT: %s\n", err)
		}
	}

	//TODO : do checking

	c.JSON(
		http.StatusOK,
		dtf.Response{
			Status:  true,
			Code:    errCode,
			Message: errMessage,
		},
	)
}

func UpdateDomain(c *gin.Context) {
	c.JSON(
		http.StatusOK,
		dtf.Response{
			Status:  true,
			Message: "TODO : implement update domain",
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
	// Menentukan file yang akan dihapus
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

	c.JSON(
		http.StatusOK,
		dtf.Response{
			Status:  true,
			Message: errMessage,
			Data:    ipDomain,
		},
	)
}

func GetAllDomainStatus(c *gin.Context) {
	c.JSON(
		http.StatusOK,
		dtf.Response{
			Status:  true,
			Message: "TODO : implement get all domain status",
		},
	)
}

func getPathDomain(name string) string {
	return fmt.Sprintf(filePath+"%s.conf", name)
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
