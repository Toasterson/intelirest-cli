rest-cli:
	go build -o rest-cli main.go

clean:
	rm -f rest-cli

all: clean rest-cli