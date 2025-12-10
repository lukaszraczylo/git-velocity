package models

// Author represents a Git/GitHub author
type Author struct {
	ID        int64  `json:"id,omitempty"`
	Login     string `json:"login"`
	Name      string `json:"name,omitempty"`
	Email     string `json:"email,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

// DisplayName returns the best available name for display
func (a *Author) DisplayName() string {
	if a.Name != "" {
		return a.Name
	}
	if a.Login != "" {
		return a.Login
	}
	if a.Email != "" {
		return a.Email
	}
	return "Unknown"
}
