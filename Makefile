all: pre pre_web
	$(shell pwd)/build/build -o $(shell pwd)/build -i $(shell pwd)/cmd/novelpackager -f novelpackager

pre:
	mkdir -p ./build;
	go build -o $(shell pwd)/build/build -ldflags "-s -w" $(shell pwd)/script/build.go;

zip:
	mkdir -p ./_pack;
	zip -r _pack/novelpackager.zip  build -x build/build

pre_web:
	cd ./pkg/web/frontend && npm install && npm run build;

web: pre_web
	go run ./cmd/novelpackager web

webD: pre_web
	go run ./cmd/novelpackager web --view