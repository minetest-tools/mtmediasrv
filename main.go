
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
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"
	"strings"
)

var arr []string
var logger *log.Logger

type FastCGIServer struct{}

func (s FastCGIServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	header := make([]byte, 4)
	version := make([]byte, 2)

	req.Body.Read(header)
	req.Body.Read(version)

	if !bytes.Equal(header, []byte("MTHS")) {
		log.Print("Request: invalid header\n")
		return
	}
	if !bytes.Equal(version, []byte {0, 1}) {
		log.Print("Request: unsupported version\n")
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
	resultarr := make([]string, 0)
	for _, v := range clientarr {
		for _, w := range arr {
			if v == w {
				resultarr = append(resultarr, v)
				break
			}
		}
	}

	// formulate response
	headers := w.Header()
	headers.Add("Content-Type", "octet/stream")
	headers.Add("Content-Length", fmt.Sprintf("%d", 6 + (len(resultarr) * 20)))

	c1, _ := w.Write([]byte(header))
	c2, _ := w.Write([]byte(version))
	c := c1 + c2
	for _, v := range resultarr {
		b, _ := hex.DecodeString(v)
		c3, _ := w.Write([]byte(b))
		c = c + c3
	}

	// log transaction
	log.Print("mtmediasrv: ", req.RemoteAddr, " '", req.UserAgent(), "' ", len(resultarr), "/", len(clientarr), " ", c)
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
	arr = make([]string, 0)
	files, _ := ioutil.ReadDir(path)
	for _, f := range files {
		h, err := getHash(strings.Join([]string{path, "/" , f.Name()}, ""))
		if err != nil {
			log.Print("parseMedia(): ", f.Name(), err)
			continue
		}

		arr = append(arr, h)
	}
}


func main() {
	p := "/var/www/media"
	parseMedia(p)
	log.Print("mtmediasrv: Number of media files: ", len(arr))

	s := "/run/mtmediasrv/sock"
	os.Remove(s)

	listener, err := net.Listen("unix", s)
	if err != nil {
		log.Fatal("mtmediasrv: net.Listen: ", err)
	}
	os.Chmod(s, 666)

	defer listener.Close()

	h := new(FastCGIServer)

	log.Print("mtmediasrv: started")

	err = fcgi.Serve(listener, h)
}
