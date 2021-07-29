package pritunl

type Server struct {
	ID       string `json:"id,omitempty"`
	Name     string `json:"name"`
	Protocol string `json:"protocol"`
	Cipher   string `json:"cipher"`
	Hash     string `json:"hash"`
}
