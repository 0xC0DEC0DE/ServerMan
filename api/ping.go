package api

import (
	"net/http"
	"os/exec"
	"strings"

	"github.com/gin-gonic/gin"
)

// PingIP handles ping requests to check if an IP is reachable
func PingIP(c *gin.Context) {
	ip := c.Param("ip")
	if ip == "" {
		c.JSON(http.StatusBadRequest, StatusMessage{Status: "error", Message: "IP address is required"})
		return
	}

	// Basic IP validation - just check if it looks like an IP
	if !isValidIP(ip) && !isValidDomain(ip) {
		c.JSON(http.StatusBadRequest, StatusMessage{Status: "error", Message: "Invalid IP address"})
		return
	}

	// Ping the IP with a timeout
	cmd := exec.Command("ping", "-c", "2", "-W", "2", ip)
	err := cmd.Run()

	status := "down"
	if err == nil {
		status = "up"
	}

	c.JSON(http.StatusOK, gin.H{"status": status})
}

// Simple IP validation
func isValidIP(ip string) bool {
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return false
	}
	for _, part := range parts {
		if len(part) == 0 || len(part) > 3 {
			return false
		}
		for _, char := range part {
			if char < '0' || char > '9' {
				return false
			}
		}
	}
	return true
}

func isValidDomain(domain string) bool {
	if len(domain) < 1 || len(domain) > 253 {
		return false
	}
	parts := strings.Split(domain, ".")
	for _, part := range parts {
		if len(part) < 1 || len(part) > 63 {
			return false
		}
		for _, char := range part {
			if !(char >= 'a' && char <= 'z') && !(char >= 'A' && char <= 'Z') && !(char >= '0' && char <= '9') && char != '-' {
				return false
			}
		}
	}
	return true
}
