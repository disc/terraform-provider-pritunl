package pritunl

type Organization struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name"`
}

func ConvertMapToOrganization(data map[string]interface{}) Organization {
	var organization Organization

	if v, ok := data["id"]; ok {
		organization.ID = v.(string)
	}
	if v, ok := data["name"]; ok {
		organization.Name = v.(string)
	}

	return organization
}
