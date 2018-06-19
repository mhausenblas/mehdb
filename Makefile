mehdb_version := 0.6

.PHONY: build clean

build :
	GOOS=linux GOARCH=amd64 go build -o ./mehdb main.go
	@docker build -t quay.io/mhausenblas/mehdb:$(mehdb_version) .
	@docker push quay.io/mhausenblas/mehdb:$(mehdb_version)

clean :
	@rm mehdb
