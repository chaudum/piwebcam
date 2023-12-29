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
	"time"

	"github.com/gorilla/mux"
)

var defaultLimit int = 100

type Conf struct {
	Delay       int `json:"delay"`
	ImageWidth  int `json:"width"`
	ImageHeight int `json:"height"`
	ImageISO    int `json:"iso"`
}

type Args struct {
	Capture   bool
	Width     int
	Height    int
	ISO       int
	Delay     int
	Port      int
	WebRoot   string
	RootDir   string
	ImageDir  string
	StaticDir string
}

type Image struct {
	Name    string    `json:"name"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"modtime"`
}

func MaxInt(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func capture(args *Args) {
	fn := filepath.Join(args.RootDir, args.ImageDir, "%d.png")
	cmd := exec.Command(
		"raspistill",
		"--timelapse", strconv.Itoa(args.Delay*1000),
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
	flag.BoolVar(&args.Capture, "capture", false, "Should program capture images")
	flag.IntVar(&args.Delay, "delay", 60, "Capture delay")
	flag.IntVar(&args.Width, "img.width", 1920, "Image width")
	flag.IntVar(&args.Height, "img.height", 1080, "Image height")
	flag.IntVar(&args.ISO, "img.iso", 100, "Image ISO")
	flag.IntVar(&args.Port, "http.port", 8765, "HTTP listen port")
	flag.StringVar(&args.WebRoot, "http.webroot", "/", "HTTP web root")
	flag.StringVar(&args.RootDir, "dir.root", cwd, "Root directory")
	flag.StringVar(&args.ImageDir, "dir.images", "images", "Image directory")
	flag.StringVar(&args.StaticDir, "dir.static", "static", "Static directory")
	flag.Parse()

	return args
}

func main() {
	var args = getArgs()

	// if args.Capture {
	// 	go capture(&args)
	// }

	prefix := args.WebRoot
	if prefix == "/" {
		prefix = ""
	}

	router := mux.NewRouter()
	router = router.PathPrefix(args.WebRoot).Subrouter()

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(args.RootDir, args.StaticDir, "index.html"))
	})

	router.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK")
	})

	router.HandleFunc("/conf", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		conf := Conf{
			Delay:       args.Delay,
			ImageWidth:  args.Width,
			ImageHeight: args.Height,
			ImageISO:    args.ISO,
		}
		json.NewEncoder(w).Encode(conf)
	})

	router.HandleFunc("/images.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		files, err := ioutil.ReadDir(filepath.Join(args.RootDir, args.ImageDir))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("[]")) // empty response
			return
		}
		sort.Slice(files, func(i, j int) bool {
			return files[i].ModTime().After(files[j].ModTime())
		})
		images := make([]Image, 0, 32)
		for idx, f := range files {
			if filepath.Ext(f.Name()) == ".png" {
				img := Image{
					Name:    fmt.Sprintf("%s/images/%s", prefix, f.Name()),
					Size:    f.Size(),
					ModTime: f.ModTime(),
				}
				images = append(images, img)
			}
			if idx >= defaultLimit-1 {
				break
			}
		}
		json.NewEncoder(w).Encode(images)
	})

	fsImg := http.FileServer(http.Dir(filepath.Join(args.RootDir, args.ImageDir)))
	router.PathPrefix("/images/").Handler(http.StripPrefix(fmt.Sprintf("%s/images/", prefix), fsImg))

	fsStatic := http.FileServer(http.Dir(filepath.Join(args.RootDir, args.StaticDir)))
	router.PathPrefix("/static/").Handler(http.StripPrefix(fmt.Sprintf("%s/static/", prefix), fsStatic))

	fmt.Printf("Serving HTTP at port %d ...\n", args.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", args.Port), router)
}
