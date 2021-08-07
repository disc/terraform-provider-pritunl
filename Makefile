build:
	go build -gcflags="all=-N -l" -o ~/.terraform.d/plugins/registry.terraform.io/disc/pritunl/0.0.1/darwin_amd64/terraform-provider-pritunl_v0.0.1 main.go
	rm -rf terraform/.terraform/ terraform/.terraform.lock.hcl

test:
	@docker rm tf_pritunl_acc_test -f || true
	@docker run --name tf_pritunl_acc_test --rm -d --privileged -p 1194:1194/udp -p 1194:1194/tcp -p 80:80/tcp -p 443:443/tcp -p 27017:27017/tcp jippi/pritunl

	sleep 10

	@chmod +x ./tools/wait-for-it.sh
	./tools/wait-for-it.sh localhost:27017 -- echo "mongodb is up"

	@docker exec -i tf_pritunl_acc_test mongo < ./tools/mongo.js

	TF_ACC=1 \
	PRITUNL_URL="https://localhost/" \
	PRITUNL_INSECURE="true" \
	PRITUNL_TOKEN=tfacctest_token \
	PRITUNL_SECRET=tfacctest_secret \
	go test -v -cover -count 1 ./internal/provider

	@docker rm tf_pritunl_acc_test -f