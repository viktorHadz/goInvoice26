package clients

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ClientAPI struct {
	Svc ClientService
}

func (a ClientAPI) Register(r *gin.Engine) {
	r.GET("/api/clients", a.getAllClients)
	r.POST("/api/clients", a.createClient)
	r.DELETE("/api/clients/:id", a.deleteClient)
	r.PATCH("/api/clients/:id", a.updateClient)
}

func (a ClientAPI) getAllClients(c *gin.Context) {
	clients, err := a.Svc.GetAll(c.Request.Context())
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server error"})
		return
	}
	c.JSON(http.StatusOK, clients)
}

func (a ClientAPI) createClient(c *gin.Context) {
	var in ClientInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	id, err := a.Svc.Create(c.Request.Context(), in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server error"})
		return
	}

	out := Client{
		ID:          id,
		Name:        in.Name,
		CompanyName: in.CompanyName,
		Address:     in.Address,
		Email:       in.Email,
	}

	c.JSON(http.StatusCreated, out)
}

func (a ClientAPI) deleteClient(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	affected, err := a.Svc.Delete(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server error"})
		return
	}

	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (a ClientAPI) updateClient(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var in UpdateClientInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	affected, err := a.Svc.Update(c.Request.Context(), id, in)
	if err != nil {
		if err.Error() == "no fields to update" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server error"})
		return
	}

	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	updated, err := a.Svc.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server error"})
		return
	}

	c.JSON(http.StatusOK, updated)
}
