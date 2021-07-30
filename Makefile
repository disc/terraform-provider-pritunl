build:
	go build -o terraform-provider-pritunl_v0.0.1 cmd/provider/main.go
	mkdir -p ~/.terraform.d/plugins/localhost/disc/pritunl/0.0.1/darwin_amd64/
	chmod +x ./terraform-provider-pritunl* && cp ./terraform-provider-pritunl* ~/.terraform.d/plugins/localhost/disc/pritunl/0.0.1/darwin_amd64/