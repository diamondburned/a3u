package main

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/diamondburned/audpl"
	"github.com/spf13/pflag"
	"github.com/ushis/m3u"
)

func init() {
	log.SetFlags(0)
	log.SetPrefix("")
}

var (
	trim   string
	prefix string
	output string = "-" // stdout default
	simple bool
)

func init() {
	os.Args[0] = filepath.Base(os.Args[0]) // cleaner
}

func main() {
	pflag.StringVarP(&trim, "trim", "t", trim, "Prefix to trim from URI")
	pflag.StringVarP(&prefix, "prefix", "p", prefix, "Prefix to prepend to URI")
	pflag.StringVarP(&output, "output", "o", output, "Path to write to, - for stdout")
	pflag.BoolVarP(&simple, "simple", "s", simple, "Output simple M3U format")
	pflag.Parse()

	var p m3u.Playlist

	if len(pflag.Args()) == 0 {
		p = append(p, convert(os.Stdin)...)
	} else {
		// If there are arguments, treat them as files.
		for _, file := range pflag.Args() {
			f, err := os.Open(file)
			if err != nil {
				log.Fatalln("Failed to open file:", err)
			}

			p = append(p, convert(f)...)
			f.Close()
		}
	}

	var err error

	var w = os.Stdout
	if output != "-" {
		w, err = os.OpenFile(output, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0664)
		if err != nil {
			log.Fatalln("Failed to open file to write to:", err)
		}
	}

	if simple {
		_, err = p.WriteSimpleTo(w)
	} else {
		_, err = p.WriteTo(w)
	}

	if err != nil {
		log.Fatalln("Failed to write playlist:", err)
	}
}

func convert(r io.Reader) []m3u.Track {
	p, err := audpl.Parse(r)
	if err != nil {
		log.Fatalln("Failed to parse:", err)
	}

	var tracks = make([]m3u.Track, 0, len(p.Tracks))
	for _, track := range p.Tracks {
		l, err := strconv.ParseInt(track.Length, 10, 64)
		if err != nil {
			log.Println("Invalid track length", track.Length, "with title", track.Title)
			continue
		}

		tracks = append(tracks, m3u.Track{
			Path:  prefix + strings.TrimPrefix(track.URI, trim),
			Title: track.Title,
			Time:  l,
		})
	}

	return tracks
}
