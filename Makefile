.PHONY: run

run:
	go run .

piwebcam:
	go build -ldflags="-extldflags=-static" -o $@ .

clean:
	rm -rf piwebcam
