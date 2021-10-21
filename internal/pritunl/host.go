package pritunl

type Host struct {
	ID       string `json:"id,omitempty"`
	Hostname string `json:"hostname"`
}
