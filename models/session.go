package models

// Session represents a single session conncected to the web site.
type Session struct {
	Username   string
	RemoteAddr string
	Host       string
	CreatedAt  string
	ExpiresAt  string
}
