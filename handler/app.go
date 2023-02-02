package handler

import (
	"net/http"

	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/cresenity/gate/config"
	dtf "github.com/cresenity/gate/datatransfer"
	"github.com/gin-gonic/gin"
)

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
	data := dtf.Configuration{
		DefaultIp: ip,
	}
	file, _ := json.MarshalIndent(data, "", " ")
	_ = ioutil.WriteFile("data/config.json", file, 0644)
	c.JSON(
		http.StatusOK,
		dtf.Response{
			Status: true,
			Data:   data,
		},
	)
}
func createDataConfiguration() bool {
	dir := "data"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.Mkdir(dir, 0755)
		if err != nil {
			log.Fatal("Error create directory data: ", err)
		}
	}
	//create file config.json
	filename := "data/config.json"

	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		file, err := os.Create(filename)
		if err != nil {
			log.Fatalf("Failed to create file: %s", err)
		}
		data := dtf.Configuration{}
		file2, _ := json.MarshalIndent(data, "", " ")
		_ = ioutil.WriteFile("data/config.json", file2, 0644)
		defer file.Close()
	}
	return true

}
