install:
	go install thebasketcase...
	sudo setcap 'CAP_NET_RAW+eip CAP_NET_ADMIN+eip' $(GOPATH)/bin/gatherworker

clean:
	rm -rf pkg/

proto:
	protoc -I model/ model/*.proto --go_out=plugins=grpc:model/
