.PHONY: build
build:
	go build -gcflags="all=-N -l" -o ~/.terraform.d/plugins/registry.terraform.io/disc/pritunl/0.0.1/darwin_amd64/terraform-provider-pritunl_v0.0.1 main.go

.PHONY: test
test:
	@docker rm tf_pritunl_acc_test -f || true
	@docker run --name tf_pritunl_acc_test --hostname pritunl.local --platform linux/amd64 --rm -d --privileged \
		-p 1194:1194/udp \
		-p 1194:1194/tcp \
		-p 80:80/tcp \
		-p 443:443/tcp \
		-p 27017:27017/tcp \
		ghcr.io/jippi/docker-pritunl:1.32.4469.94

	# Wait for MongoDB to be ready inside the container
	@chmod +x ./tools/wait-for-mongo.sh
	./tools/wait-for-mongo.sh tf_pritunl_acc_test 60

	# enables an api access for the pritunl user, updates an api token and secret
	@docker exec -i tf_pritunl_acc_test mongo pritunl < ./tools/mongo.js

	# Wait for Pritunl web server
	@chmod +x ./tools/wait-for-it.sh
	./tools/wait-for-it.sh localhost:443 -t 60 -- echo "pritunl web server is up"

	# Wait for API to be ready with credentials
	@chmod +x ./tools/wait-for-api.sh
	./tools/wait-for-api.sh https://localhost/state 60

	TF_ACC=1 \
	PRITUNL_URL="https://localhost/" \
	PRITUNL_INSECURE="true" \
	PRITUNL_TOKEN=tfacctest_token \
	PRITUNL_SECRET=tfacctest_secret \
	go test -v -cover -count 1 ./internal/provider

	@docker rm tf_pritunl_acc_test -f
