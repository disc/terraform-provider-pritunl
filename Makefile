.PHONY: build
build:
	go build -gcflags="all=-N -l" -o ~/.terraform.d/plugins/registry.terraform.io/disc/pritunl/0.0.1/darwin_amd64/terraform-provider-pritunl_v0.0.1 main.go

.PHONY: test
test:
	@docker rm tf_pritunl_acc_test -f || true
	@docker run --name tf_pritunl_acc_test --hostname pritunl.local --rm -d --privileged \
		-p 1194:1194/udp \
		-p 1194:1194/tcp \
		-p 80:80/tcp \
		-p 443:443/tcp \
		-p 27017:27017/tcp \
		ghcr.io/jippi/docker-pritunl:1.32.4469.94

	@chmod +x ./tools/wait-for-it.sh
	./tools/wait-for-it.sh localhost:27017 -t 60 -- echo "mongodb is up"
	./tools/wait-for-it.sh localhost:443 -t 60 -- echo "pritunl web server is up"

	# enables an api access for the pritunl user, updates an api token and secret
	@docker exec -i tf_pritunl_acc_test mongo pritunl < ./tools/mongo.js

	# Wait for API to be ready with credentials
	@echo "Waiting for Pritunl API to be ready..."
	@for i in 1 2 3 4 5 6 7 8 9 10 11 12; do \
		if curl -sk https://localhost/state >/dev/null 2>&1; then \
			echo "Pritunl API is ready"; \
			break; \
		fi; \
		echo "Attempt $$i: API not ready, waiting 5s..."; \
		sleep 5; \
	done

	TF_ACC=1 \
	PRITUNL_URL="https://localhost/" \
	PRITUNL_INSECURE="true" \
	PRITUNL_TOKEN=tfacctest_token \
	PRITUNL_SECRET=tfacctest_secret \
	go test -v -cover -count 1 ./internal/provider

	@docker rm tf_pritunl_acc_test -f
