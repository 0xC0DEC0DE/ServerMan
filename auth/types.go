package auth

type User struct {
	ID     int      `json:"id"`
	Email  string   `json:"email"`
	Groups []string `json:"groups"`
}
