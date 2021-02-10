package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

var defaultLimit int = 100

type Conf struct {
	Delay       int `json:"delay"`
	ImageWidth  int `json:"width"`
	ImageHeight int `json:"height"`
	ImageISO    int `json:"iso"`
}

type Args struct {
	Width int
	Height int
	ISO int
	Delay int
	Root string
	Port int
}

type Image struct {
	Name string `json:"name"`
	Size int64 `json:"size"`
	ModTime time.Time `json:"modtime"`
}

func MaxInt(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func capture(args *Args) {
	fn := filepath.Join(args.Root, "images", "img-%d.jpg")
	cmd := exec.Command(
		"raspistill",
	  "--timelapse", strconv.Itoa(args.Delay * 1000),
	  "--timeout", "0",
		"--width", strconv.Itoa(args.Width),
		"--height", strconv.Itoa(args.Height),
		"--quality", "100",
		"--ISO", strconv.Itoa(args.ISO),
		"--annotate", "12",
		"--output", fn,
	  "--timestamp")
	_, err := cmd.Output()
	if err != nil {
		log.Print(err)
	}
}

func getArgs() Args {
	cwd, _ := os.Getwd()

	var args Args
	flag.IntVar(&args.Width, "width", 1024, "Image width")
	flag.IntVar(&args.Height, "height", 768, "Image height")
	flag.IntVar(&args.ISO, "iso", 100, "Image ISO")
	flag.IntVar(&args.Delay, "delay", 60, "Capture delay")
	flag.StringVar(&args.Root, "http.root", cwd, "HTTP root directory")
	flag.IntVar(&args.Port, "http.port", 8765, "HTTP listen port")
	flag.Parse()

	return args
}

func main() {
	var args = getArgs()

	go capture(&args)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(args.Root, "static", "index.html"))
	})

	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK")
	})

	http.HandleFunc("/conf", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		conf := Conf{
			Delay:       args.Delay,
			ImageWidth:  args.Width,
			ImageHeight: args.Height,
			ImageISO:    args.ISO,
		}
		json.NewEncoder(w).Encode(conf)
	})

	http.HandleFunc("/images.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		files, err := ioutil.ReadDir(filepath.Join(args.Root, "images"))
		if err != nil {
			log.Fatal(err)
		}
		sort.Slice(files, func(i, j int) bool {
			if files[i].ModTime().Unix() > files[j].ModTime().Unix() {
				return true
			} else {
				return false
			}
		})
		var images []Image
		for idx, f := range files {
			if filepath.Ext(f.Name()) == ".jpg" && strings.HasPrefix(f.Name(), "img-") {
				var img = Image{
					Name: fmt.Sprintf("/images/%s", f.Name()),
					Size: f.Size(),
					ModTime: f.ModTime(),
				}
				images = append(images, img)
			}
			if idx >= defaultLimit - 1 {
				break
			}
		}
		json.NewEncoder(w).Encode(images)
	})

	fsImg := http.FileServer(http.Dir(filepath.Join(args.Root, "images")))
	http.Handle("/images/", http.StripPrefix("/images", fsImg))

	fsStatic := http.FileServer(http.Dir(filepath.Join(args.Root, "static")))
	http.Handle("/static/", http.StripPrefix("/static", fsStatic))

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", args.Port), nil))
}
