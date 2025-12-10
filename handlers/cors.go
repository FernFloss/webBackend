package handlers

import "github.com/gin-gonic/gin"

// CORSMiddleware allows cross-origin requests for testing/frontends hosted elsewhere.
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("Access-Control-Allow-Origin", "*")
		h.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		h.Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		h.Set("Access-Control-Expose-Headers", "Content-Length")

		// Handle preflight
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

