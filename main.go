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

var slashesc = strings.NewReplacer("/", "∕", `\`, "⧵").Replace

var (
	trim   string
	prefix string
	output string = "-" // stdout default
	extens string
	simple bool
)

func init() {
	os.Args[0] = filepath.Base(os.Args[0]) // cleaner

	pflag.Usage = func() {
		log.Println("Usage:", os.Args[0], "[flags...] path")
		log.Println("  Path can be a directory, in which the application will make a file")
		log.Println("  with the playlist name as the filename and sanitize it by substituting")
		log.Println("  slashes.")
		log.Println("")

		log.Println("Flags:")
		pflag.PrintDefaults()
	}
}

func main() {
	pflag.StringVarP(&trim, "trim", "t", trim, "Prefix to trim from URI")
	pflag.StringVarP(&prefix, "prefix", "p", prefix, "Prefix to prepend to URI")
	pflag.StringVarP(&output, "output", "o", output, "Path to write to, - for stdout")
	pflag.StringVarP(&extens, "extension", "e", extens, "Extension to override in filename")
	pflag.BoolVarP(&simple, "simple", "s", simple, "Output simple M3U format")
	pflag.Parse()

	if len(pflag.Args()) != 1 {
		pflag.Usage()
		os.Exit(2)
	}

	f, err := os.Open(pflag.Arg(0))
	if err != nil {
		log.Fatalln("Failed to open file:", err)
	}

	name, tracks := convert(f)
	p := m3u.Playlist(tracks)

	f.Close()

	var w = os.Stdout
	if output != "-" && output != "" {
		// Is the output is a directory, then use the playlist name instead.
		if isdir(output) {
			output = filepath.Join(output, slashesc(name))
		}

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

func convert(r io.Reader) (string, []m3u.Track) {
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

		var path = prefix + strings.TrimPrefix(track.URI, trim)
		if extens != "" {
			path = convertExt(path, extens)
		}

		tracks = append(tracks, m3u.Track{
			Path:  path,
			Title: track.Title,
			Time:  l,
		})
	}

	return p.Name, tracks
}

func convertExt(file string, ext string) string {
	oldExt := filepath.Ext(file)
	return file[:len(file)-len(oldExt)] + "." + ext
}

func isdir(path string) bool {
	if len(path) > 0 && path[len(path)-1] == os.PathSeparator {
		return true
	}

	if s, err := os.Stat(path); err == nil && s.IsDir() {
		return true
	}

	return false
}
