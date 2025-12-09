
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
	rsync -av --delete . /home/jmguirao/aplicaciones/grampal/service --exclude .git
	sudo systemctl restart grampal-dic
	sudo systemctl restart grampal-des
	


