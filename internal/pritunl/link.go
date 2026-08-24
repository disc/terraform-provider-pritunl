package pritunl

type Link struct {
	ID              string `json:"id,omitempty"`
	Name            string `json:"name,omitempty"`
	Server          string `json:"server,omitempty"`
	Status          string `json:"status,omitempty"`
	UseLocalAddress bool   `json:"use_local_address,omitempty"`
	Address         string `json:"address,omitempty"`
}
