package handler

import (
	"net/http"

	dtf "github.com/cresenity/gate/datatransfer"
	"github.com/gin-gonic/gin"
)

func InstallSsl(c *gin.Context) {
	c.JSON(
		http.StatusOK,
		dtf.Response{
			Status:  true,
			Message: "TODO : implement install ssl",
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

func GetDomainStatus(c *gin.Context) {
	c.JSON(
		http.StatusOK,
		dtf.Response{
			Status:  true,
			Message: "TODO : implement get domain status",
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
