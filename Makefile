build:
	go build -gcflags="all=-N -l" -o ~/.terraform.d/plugins/localhost/disc/pritunl/0.0.1/darwin_amd64/terraform-provider-pritunl_v0.0.1 cmd/provider/main.go
	rm -rf terraform/.terraform/ terraform/.terraform.lock.hcl