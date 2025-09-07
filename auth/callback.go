package auth

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// Handler for our callback.
func Callback(auth *Authenticator, ch *sql.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		fmt.Println(session.Get("state"), ctx.Query("state"))
		if ctx.Query("state") != session.Get("state") {
			ctx.String(http.StatusBadRequest, "Invalid state parameter.")
			return
		}

		// Exchange an authorization code for a token.
		token, err := auth.Exchange(ctx.Request.Context(), ctx.Query("code"))
		if err != nil {
			ctx.String(http.StatusUnauthorized, "Failed to exchange an authorization code for a token.")
			return
		}

		idToken, err := auth.VerifyIDToken(ctx.Request.Context(), token)
		if err != nil {
			ctx.String(http.StatusInternalServerError, "Failed to verify ID Token.")
			return
		}

		var profile map[string]interface{}
		if err := idToken.Claims(&profile); err != nil {
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}

		email, ok := profile["email"].(string)
		if !ok {
			ctx.Redirect(http.StatusTemporaryRedirect, "/?err=no_email")
			return
		}

		if err := EnsureUser(ch, email); err != nil {
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}

		// Create a database session
		sessionToken, err := CreateSession(ch, email)
		if err != nil {
			ctx.String(http.StatusInternalServerError, "Failed to create session: "+err.Error())
			return
		}

		// Store only the session token in the cookie session
		session.Set("session_token", sessionToken)
		if err := session.Save(); err != nil {
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}

		// Redirect to logged in page.
		ctx.Redirect(http.StatusTemporaryRedirect, "/")
	}
}

func EnsureUser(ch *sql.DB, email string) error {
	var exists string
	err := ch.QueryRow("SELECT email FROM users WHERE email = ? LIMIT 1", email).Scan(&exists)
	if err != nil {
		return err
	}
	if exists == "" {
		_, err = ch.Exec("INSERT INTO users (email, groups) VALUES (?, ?)", email, "[]")
		if err != nil {
			return err
		}
	}
	return nil
}
