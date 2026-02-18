build:
	go build -o signet .

test:
	go test ./...

deploy-137: build
	scp signet root@192.168.120.137:/usr/local/bin/signet

deploy-138: build
	scp signet root@192.168.120.138:/usr/local/bin/signet

deploy-all: deploy-137 deploy-138
