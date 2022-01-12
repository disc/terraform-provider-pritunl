package pritunl

type Key struct {
	ID        string `json:"id,omitempty"`
	KeyUrl    string `json:"key_url"`
	KeyZipUrl string `json:"key_zip_url"`
	KeyOncURL string `json:"key_onc_url"`
	ViewUrl   string `json:"view_url"`
	UriUrl    string `json:"uri_url"`
}
