all: mochat wikiexplorer

mochat:
	go build --tags sqlite_vtable ./cmd/mochat

wikiexplorer:
	go build --tags sqlite_vtable ./cmd/wikiexplorer

clean:
	rm -f ./mochat
	rm -f ./wikiexplorer
