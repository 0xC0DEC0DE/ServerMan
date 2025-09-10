package main

import (
	"context"
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"time"

	api "github.com/callowaysutton/servercon/api"
	auth "github.com/callowaysutton/servercon/auth"
	db "github.com/callowaysutton/servercon/db"
	"github.com/callowaysutton/servercon/sync"
	"github.com/joho/godotenv"
	"golang.org/x/net/websocket"

	vncProxy "github.com/evangwt/go-vncproxy"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load .env if present
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, continuing with system env")
	}

	auth_client, err := auth.New()
	if err != nil {
		log.Fatalf("Failed to initialize the authenticator: %v", err)
	}

	// Initialize Gin router
	r := gin.Default()
	ch := db.Connect()

	// Run migrations
	db.RunMigrations(ch, "./migrations")

	// Start sync service (syncs every 5 minutes)
	syncService := sync.NewSyncService(ch, 5*time.Minute)
	syncService.Start()

	// Start session cleanup (runs every hour)
	auth.StartSessionCleanup(ch, time.Hour)

	// Initialize sessions
	gob.Register(map[string]interface{}{})
	store := cookie.NewStore([]byte("secret"))

	// Auth routes
	r.Use(sessions.Sessions("auth-session", store))
	r.GET("/login", auth.Login(auth_client))
	r.GET("/callback", auth.Callback(auth_client, ch))
	r.GET("/logout", auth.Logout(ch))

	// Token handler: resolve token dynamically
	tokenHandler := func(token string) (string, error) {
		ctx := context.Background()

		var (
			name    string
			vncHost string
			vncPort int
		)

		// Run your query
		err := ch.QueryRowContext(ctx, `
            SELECT d.name, c.vnc_host, c.vnc_port
            FROM server_details_cache d
            JOIN server_credentials_cache c ON d.server_id = c.server_id
            WHERE d.name = ?
            LIMIT 1
        `, token).Scan(&name, &vncHost, &vncPort)

		if err != nil {
			return "", fmt.Errorf("invalid token")
		}

		// Construct the target host:port (assuming static IP)
		backend := fmt.Sprintf("%s:%d", vncHost, vncPort)

		// (Optional) you can also inject password auth here
		// if go-vncproxy supports passing password downstream.

		return backend, nil
	}

	// Create the proxy server
	vncProxy := vncProxy.New(&vncProxy.Config{
		LogLevel: vncProxy.DebugLevel,
		TokenHandler: func(r *http.Request) (addr string, err error) {
			// validate token and get forward vnc addr
			token := r.URL.Query().Get("token")
			if token == "" {
				return "", fmt.Errorf("missing token")
			}
			addr, err = tokenHandler(token)
			if err != nil {
				return "", err
			}
			return
		},
	})

	r.GET("/ws", func(ctx *gin.Context) {
		h := websocket.Handler(vncProxy.ServeWS)
		h.ServeHTTP(ctx.Writer, ctx.Request)
	})

	// API routes
	api_routes := r.Group("/api")
	api_routes.Use(auth.IsAuthenticated(ch))
	{
		api_routes.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "pong"})
		})

		api_routes.GET("/ping/:ip", api.PingIP)

		api_routes.GET("/users", func(c *gin.Context) {
			rows, _ := ch.Query("SELECT id, email FROM users LIMIT 10")
			defer rows.Close()

			users := []map[string]interface{}{}
			for rows.Next() {
				var id int
				var email string
				rows.Scan(&id, &email)
				users = append(users, gin.H{"id": id, "email": email})
			}
			c.JSON(200, users)
		})

		api_routes.GET("/servers", func(c *gin.Context) {
			handler := api.GetServers(c, ch)
			handler(c)
		})
		api_routes.GET("/server/:id", func(c *gin.Context) {
			handler := api.GetServer(c, ch)
			handler(c)
		})
		api_routes.GET("/server/:id/snapshots", func(c *gin.Context) {
			handler := api.GetServerSnapshots(c, ch)
			handler(c)
		})
		api_routes.GET("/server/:id/credentials", func(c *gin.Context) {
			handler := api.GetServerCredentials(c, ch)
			handler(c)
		})
		api_routes.GET("/user", api.GetUser(ch))
		api_routes.GET("/os_options", func(c *gin.Context) {
			api.GetOsOptions(c, ch)
		})
		api_routes.GET("/apps", func(c *gin.Context) {
			api.GetApps(c, ch)
		})

		// Management API routes
		api_routes.POST("/server/:id/action/:action", api.ServerAction)
		api_routes.POST("/server/:id/reinstall", api.ReinstallServer)
		api_routes.POST("/server/:id/restore-snapshot", api.RestoreSnapshot)
		api_routes.POST("/server/:id/reset-password", api.ResetRootPassword)
		api_routes.POST("/server/:id/console/:action", api.ChangeConsoleStatus)

		api_routes.POST("/admin/sync", api.TriggerSync(ch, syncService.ManualSync))
		api_routes.GET("/admin/api-keys/status", api.GetAPIKeyStatus(ch, syncService.GetAPIKeyStatus))
		api_routes.GET("/admin/users", api.GetAllUsers(ch))
		api_routes.POST("/admin/users", api.AddUser(ch))
		api_routes.PUT("/admin/users/:email", api.UpdateUserGroups(ch))
		api_routes.DELETE("/admin/users/:email", api.DeleteUser(ch))
	}

	// Serve frontend (catch-all *after* /api)
	r.NoRoute(func(c *gin.Context) {
		c.File("./frontend/dist/index.html")
	})
	r.Static("/static", "./frontend/dist/static")

	r.Run(":8080")
}
