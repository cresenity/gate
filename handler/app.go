package handler

import (
	"net/http"

	"encoding/json"
	"io/ioutil"
	"log"

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
	content, err := ioutil.ReadFile("data/config.json")
	if err != nil {
		log.Fatal("Error when opening file: ", err)
	}
	var data dtf.Configuration
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
	ip := c.Param("ipAddress")
	data := dtf.Configuration{
		DefaultIp: "127.0.0.1",
		Ip:        ip,
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
