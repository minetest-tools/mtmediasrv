
## mtmediasrv

A Minetest Media server implementation as fcgi server.


### Requirements

- nginx or apache, with fcgi
- webserver must be able to access file sockets in /run/mtmediasrv
- systemd for controlling the service startup
- go to build the service

This program is intended to run as fcgi process and handle POST
requests for the `/index.mth` URI. It listens on a local unix domain
socket, and needs to read the media files in the media folder to
create sha1 hashes. It creates a hash list of files it has available
and keeps this in memory.

When a client connects, the client POSTS their list of known sha1
hashes of files they need.

The program verifies for each needed hash that the server has the hash
listed, and returns the needed hashes only, and only hashes that it
has the files for.

The actual file content is not served by this program, for this you
need to have your web server serve that content as static files.


### Building

Run `go build` in this folder to create the binary `mtschemsrv`.

Please note the currently hardcoded values in the binary and
example configuration files.


### Installation

Several example config and service units are provided as an example
on how to deploy this service.

Once you have configured your web server properly so that it serves
up the static content files by hash, point the mtmediasrv to the
same folder and it will talk to minetest clients that request media
automatically. If the content changes, you need to restart the program.

You should not have a file called `index.mth` in the media folder,
although this will not break anything, it will probably be confusing.


### logging

The program logs all clients and writes out the following
data to the log output (journal/syslog):

- remote address
- user agent
- hashes needed/hashes given
- content-length of sent bytes

Example output:

```
mtmediasrv: 5.4.3.2:34567 'Minetest/0.4.15 (Linux/4.10.1 x86_64)' 64/64 1286
```

