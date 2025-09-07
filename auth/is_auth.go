package auth

import (
	"database/sql"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func IsAuthenticated(db *sql.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		sessionToken := session.Get("session_token")

		if sessionToken == nil {
			ctx.Redirect(http.StatusSeeOther, "/")
			ctx.Abort()
			return
		}

		// Validate session token against database
		email, valid := ValidateSession(db, sessionToken.(string))
		if !valid {
			// Clear invalid session
			session.Clear()
			session.Save()
			ctx.Redirect(http.StatusSeeOther, "/")
			ctx.Abort()
			return
		}

		// Store email in context for use in handlers
		ctx.Set("user_email", email)
		ctx.Next()
	}
}
