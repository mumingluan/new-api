package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetXuancatMarker(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"457": true})
}
