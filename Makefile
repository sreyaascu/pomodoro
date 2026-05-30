build:
	go build -o pomo

install:
	sudo cp pomo /usr/local/bin/ 

run:
	go run . start
