
terminal:
	go run .

debug:
	go run . --debug

trace:
	go run . --trace

dev:
	go run . --ser --cors

update:
	git pull origin master
	go build .
	rsync 


