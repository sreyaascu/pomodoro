build:
	go build -o pomo

install: build
	sudo cp pomo /usr/local/bin/ 

run:
	go run . start
