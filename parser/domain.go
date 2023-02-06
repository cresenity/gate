package domain

import (
	"fmt"
	"os"
)

const (
	filePath     string = "/etc/nginx/conf.d/"
	fileTemplate string = `server {
		listen       80;
		listen  [::]:80;
		server_name  %s;
	
		location / {
			proxy_pass %s;
		}
	
		error_page   500 502 503 504  /50x.html;
		location = /50x.html {
			root   /usr/share/nginx/html;
		}
	}`
)

func getPathDomain(name string) string {
	return fmt.Sprintf(filePath+" %s", name)
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
