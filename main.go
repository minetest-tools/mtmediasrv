//
//    mtmediasrv - a Minetest Media server implementation done right
//
//    Copyright (C) 2017 - Auke Kok <sofar@foo-projects.org>
//
//    This program is free software: you can redistribute it and/or modify
//    it under the terms of the GNU Affero General Public License as
//    published by the Free Software Foundation, either version 3 of the
//    License, or (at your option) any later version.
//
//    This program is distributed in the hope that it will be useful,
//    but WITHOUT ANY WARRANTY; without even the implied warranty of
//    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//    GNU Affero General Public License for more details.
//
//    You should have received a copy of the GNU Affero General Public License
//    along with this program.  If not, see <http://www.gnu.org/licenses/>.
//

package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/spf13/viper"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"
	"path/filepath"
	"strings"
)

var (
	Version string
	Build   string

	newmedia int

	arr map[string]bool
)

// Change our logging to be journalctl friendly
type logWriter struct {
}

func (writer logWriter) Write(bytes []byte) (int, error) {
	return fmt.Print(string(bytes))
}

type FastCGIServer struct{}

func (s FastCGIServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	header := make([]byte, 4)
	version := make([]byte, 2)

	req.Body.Read(header)
	req.Body.Read(version)
	r := req.Referer()
	// omit port numbers from peers
	ip, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		http.Error(w, err.Error(), 500)
		log.Print("Request: unable to identify peer\n")
		return
	}
	remoteip := net.ParseIP(ip).String()

	if r == "" {
		// reject all clients without a referrer URL set
		http.Error(w, "A referer URL is required", 403)
		log.Print(remoteip, " '", req.UserAgent(), "' REJECT 0 -", r)
		return
	}

	if req.Method != "POST" {
		w.Header().Set("Access-Control-Allow-Headers", "POST")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		log.Printf("Invalid GET from %v\n", remoteip)
		return
	}

	if !bytes.Equal(header, []byte("MTHS")) {
		log.Print(remoteip, ": invalid MTHS header")
		return
	}
	if !bytes.Equal(version, []byte{0, 1}) {
		log.Print(remoteip, ": unsupported MTHS version\n")
		return
	}

	// read client needed hashes
	clientarr := make([]string, 0)
	for {
		h := make([]byte, 20)
		_, err := req.Body.Read(h)
		if err != nil {
			break
		}
		clientarr = append(clientarr, hex.EncodeToString(h))
	}

	// Iterate over client hashes and remove hashes that we don't have from it
	resultmap := map[string]bool{}
	for _, v := range clientarr {
		if arr[v] {
			resultmap[v] = true
		}
	}

	// formulate response
	headers := w.Header()
	headers.Add("Content-Type", "octet/stream")
	headers.Add("Content-Length", fmt.Sprintf("%d", 6+(len(resultmap)*20)))

	c1, _ := w.Write([]byte(header))
	c2, _ := w.Write([]byte(version))
	c := c1 + c2
	for k := range resultmap {
		b, _ := hex.DecodeString(k)
		c3, _ := w.Write([]byte(b))
		c = c + c3
	}

	// log transaction
	log.Print(remoteip, " '", req.UserAgent(), "' ", len(resultmap), "/", len(clientarr), " ", c, " ", r)
}

func getHash(path string) (string, error) {
	var hashStr string

	f, err := os.Open(path)
	if err != nil {
		return hashStr, err
	}
	defer f.Close()

	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		return hashStr, err
	}

	hashStr = hex.EncodeToString(h.Sum(nil)[:20])
	return hashStr, nil
}

func parseMedia(path string) {
	files, _ := ioutil.ReadDir(path)
	for _, f := range files {
		h, err := getHash(strings.Join([]string{path, "/", f.Name()}, ""))
		if err != nil {
			log.Print("parseMedia(): ", f.Name(), err)
			continue
		}

		arr[h] = true
	}
}

func collectMedia(l bool, c bool, e map[string]bool, w string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Print(err)
			return nil
		}
		if info.IsDir() {
			return nil
		}
		ext := filepath.Ext(path)
		if e[ext] {
			sha, err := getHash(path)
			if err != nil {
				return err
			}

			of := strings.Join([]string{w, sha}, "/")

			if l {
				err := os.Link(path, of)
				if err != nil {
					if os.IsExist(err) {
						return nil
					}
					return err
				}
				newmedia++
			} else if c {
				in, err := os.Open(path)
				if err != nil {
					return err
				}
				defer in.Close()
				os.Remove(of)
				out, err := os.Create(of)
				if err != nil {
					return err
				}
				defer out.Close()
				_, err = io.Copy(out, in)
				closeErr := out.Close()
				if err != nil {
					return err
				}
				newmedia++
				return closeErr
			}
		}
		return nil
	}
}

func main() {
	// log output format
	log.SetFlags(0)
	log.SetOutput(new(logWriter))

	// config stuff
	viper.SetConfigName("mtmediasrv")
	viper.SetConfigType("yaml")

	viper.AddConfigPath("/usr/share/defaults/etc")
	viper.AddConfigPath("/etc")
	viper.AddConfigPath("$HOME/.config")

	viper.SetDefault("socket", "/run/mtmediasrv/sock")
	viper.SetDefault("webroot", "/var/www/media")
	viper.SetDefault("mediapath", []string{})
	viper.SetDefault("mediascan", "true")
	viper.SetDefault("medialink", "true")
	viper.SetDefault("mediacopy", "false")
	viper.SetDefault("extensions", []string{
		// Image formats
		".png", ".jpg", ".jpeg", 
		// Sound
		".ogg",
		// Models
		".x", ".b3d", ".obj",
		// Translations
		".tr",
	})

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal("Error in confog file: ", err)
	}

	// step 1, collect media files
	w := viper.GetString("webroot")
	ext := viper.GetStringSlice("extensions")
	extmap := make(map[string]bool)
	for i := 0; i < len(ext); i++ {
		extmap[ext[i]] = true
	}
	if viper.GetBool("mediascan") {
		l := viper.GetBool("medialink")
		c := viper.GetBool("mediacopy")
		if !(l || c) {
			log.Fatal("mediascan enabled but both medialink and mediacopy are disabled!")
		}
		if len(viper.GetStringSlice("mediapath")) == 0 {
			log.Fatal("empty mediapath list, but mediascan was enabled!")
		}
		for _, v := range viper.GetStringSlice("mediapath") {
			log.Print("Scaning mediapath: ", v)
			err := filepath.Walk(v, collectMedia(l, c, extmap, w))
			if err != nil {
				log.Fatal(err)
			}
		}
		log.Print("mediascan linked/copied files: ", newmedia)
	}

	// step 2, fill our hash table `arr`
	arr = make(map[string]bool)
	parseMedia(w)
	log.Print("Number of media files: ", len(arr))

	s := viper.GetString("socket")
	os.Remove(s)

	listener, err := net.Listen("unix", s)
	if err != nil {
		log.Fatal("net.Listen: ", err)
	}
	os.Chmod(s, 0666)

	defer listener.Close()

	h := new(FastCGIServer)

	log.Print("version ", Version, " (", Build, ") started")

	err = fcgi.Serve(listener, h)
	if err != nil {
		log.Fatal("fcgi.Serve: ", err)
	}
}
