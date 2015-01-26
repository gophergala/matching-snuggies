Matching Snuggies
=================

Matching Snuggies is 'slicing' software that converts 3D models into G-code
that can be printed by common 3D printers.

Matching Snuggies exposes a backend slicing program (currently
[Slic3r](http://slic3r.org/)) through an HTTP API.  A command line slicing tool
is provided for ease of use and to support eventual integration with host
software like Repetier-Host and OctoPrint.

Matching Snuggies is well suited for integration with host-software that may
run in a resource constrained environment, such as a Raspberry Pi.

Documentation
=============

Install
-------

NOTE: OS X users should symlink the executable at
/Applications/Slic3r.app/Contents/MacOS/slic3r into their environment's PATH
(e.g. to /usr/bin/slic3r).

./build.sh

On your server machine, install [slic3r](http://slic3r.org/download), the
backend slicing software, onto the machine you want to act as your server.
Then create a directory containing the Slic3r configuration files you would
like clients to have make available.  See the Slic3r [doc](slic3r/README.md)
for more information.

Slicing Server
--------------

The `snuggied` command runs a REST API for scheduling slicing jobs, retrieving
gcode, and getting periodic status updates while slicing is in progress.

```
./bin/snuggied -slic3r.configs=./slic3r
```

See the snuggied documentation on
[godoc.org](http://godoc.org/github.com/gophergala/matching-snuggies/cmd/snuggied)
or the API [doc](API.md) for information about each endpoint.

Command line tool
-----------------

After starting, the server can be sent files to slice using the command line
tool.

```
./bin/snuggier -o FirstCube.gcode testdata/FirstCube.amf
```

When `snuggied` is running on another host specify the server when calling `snuggier`.

```
./bin/snuggier -server=10.0.10.123:8888 -preset=hq -o FirstCube.gcode testdata/FirstCube.amf
```

See the snuggier command documentation on godoc.org
[godoc.org](http://godoc.org/github.com/gophergala/matching-snuggies/cmd/snuggier).

Long term goals
---------------

- API authorization
- integration with other backend slicers (Cura)
- a slicing queue that may be consumed by a pool of workers (shared
  configuration; dropbox?)
- cluster health/monitoring dashboard
