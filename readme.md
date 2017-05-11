
## mtmediasrv

A Minetest Media server implementation as fcgi server.

This program is useful to distribute minetest media (textures, models,
sounds) to minetest clients for multiplayer server owners that wish to
have their media hosted on a `remote media server` URL. Doing this as
a separate download removes some of the download bandwidth from the
actual game server and offloads it to a different HTTP server. This
will work for clients that have cURL support enabled.


### Requirements

- nginx or apache, with fcgi
- webserver must be able to access file sockets in /run/mtmediasrv
- systemd for controlling the service startup
- mod assets to serve

Optionally, if you want to compile the source yourself:

- go to build the service

Binary builds are available for Linux under `releases` on the github
project. Other OS builds can be added on request, but the code is
intended to run headless on a Linux server.

This program is intended to run as fcgi process and handle POST
requests for the `/index.mth` URI. It listens on a local unix domain
socket, and needs to read the media files in the media folder to
create sha1 hashes. It creates a hash list of files it has available
and keeps this in memory.

At startup, the program can optionally scan mods and subgames to find
and copy or hardlink (the default) all the assets into the webroot.
The hardlink method is better for space, but may fail if the media
is not on the same filesystem as the webroot.

When a client connects, the client POSTS their list of known sha1
hashes of files they need.

The program verifies for each needed hash that the server has the hash
listed, and returns the needed hashes only, and only hashes that it
has the files for.

The actual file content is not served by this program, for this you
need to have your web server serve that content as static files.


### Building

Note that binaries are available in the releases section of the
github project.

mtmediasrv uses `viper` to read configuration files. You must
`go get https://github.com/spf13/viper` before building.
Run `make` in this folder to create the binary `mtschemsrv`.


### Installation

Several example config and service units are provided as an example
on how to deploy this service.

Once you have configured your web server properly so that it serves
up the static content files by hash, point the mtmediasrv to the
same folder and it will talk to minetest clients that request media
automatically. If the content changes, you need to restart the program.

You should not have a file called `index.mth` in the media folder,
although this will not break anything, it will probably be confusing.

Copy and edit the `mtmediasrv.yaml` file and point it at the proper
webroot, socket path, and mediapath entries. Place it in /etc/ or
~/.config/.


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

