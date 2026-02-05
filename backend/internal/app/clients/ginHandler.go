package clients

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ClientAPI struct {
	Svc ClientService
}

func (a ClientAPI) Register(r *gin.Engine) {
	r.GET("/clients", a.getAllClients)
	r.POST("/clients", a.createClient)
}

func (a ClientAPI) getAllClients(c *gin.Context) {
	clients, err := a.Svc.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server error"})
		return
	}
	c.JSON(http.StatusOK, clients)
}

func (a ClientAPI) createClient(c *gin.Context) {
	var in CreateClientInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	id, err := a.Svc.Create(c.Request.Context(), in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}
