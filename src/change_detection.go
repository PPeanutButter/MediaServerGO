package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"os/exec"
)

func getTasks(c *gin.Context) {
	result, err := readStringFromCmd(exec.Command("change_detection", "--show_tasks"))
	if err != nil {
		c.JSON(http.StatusOK, result)
	}
}

func getConfigurations(c *gin.Context) {
	result, err := readStringFromCmd(exec.Command("change_detection", "--show_cfg"))
	if err != nil {
		c.JSON(http.StatusOK, result)
	}
}

func getRecords(c *gin.Context) {
	result, err := readStringFromCmd(exec.Command("change_detection", "--show_records"))
	if err != nil {
		c.JSON(http.StatusOK, result)
	}
}

func runOnce(c *gin.Context) {
	go func() {
		exec.Command("change_detection")
	}()
	c.Status(http.StatusOK)
}
