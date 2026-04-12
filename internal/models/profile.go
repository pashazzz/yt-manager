package models

// Profile представляет пользователя, идентифицированного через Tailscale.
// ID = Tailscale-User-Login (e-mail адрес).
type Profile struct {
	ID   string `json:"id"`   // Tailscale-User-Login
	Name string `json:"name"` // Tailscale-User-Name
}
