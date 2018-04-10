all: nodeos

nodeos:
	rm -f nodeos
	go build programs/nodeos/nodeos.go

