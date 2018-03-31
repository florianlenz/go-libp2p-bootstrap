deps:
	go get -u github.com/whyrusleeping/gx
	go get -u github.com/whyrusleeping/gx-go
	go get -u github.com/stretchr/testify/require
install:
	gx install
deps_hack:
	gx-go rewrite
deps_hack_revert:
	gx-go uw
test:
	go test -cover