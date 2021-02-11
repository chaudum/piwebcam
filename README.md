# piwebcam

_A  webcam tool for Raspberry Pi with a web interface written in Go._

> This project was created to learn Go basics.

## Installation

Requires **Go >= 1.11**.
On Raspian, for example, install using `apt`:

```
sudo apt install golang
```

Clone this Git repository on your Raspberry Pi in the home directory of the
user `pi`:

```
git clone https://github.com/chaudum/piwebcam.git
cd piwebcam
go build piwebcam.go
```

The `piwebcam` binary can be invoked within the same folder.

### systemd service

The project contains a template for a systemd service file to run the program
in the background. Edit `systemd/piwebcam.service` to adjust configuration
parameters. Make sure the `-http.root` is set correctly!

```
vi systemd/piwebcam.service
```

Thereafter, enable the file and start the service.

```
sudo systemctl enable `readlink -f systemd/piwebcam.service`
sudo systemctl start piwebcam
```
