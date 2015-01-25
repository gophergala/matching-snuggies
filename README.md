Matching Snuggies
=================

Matching Snuggies is a slicing software that exposes a backend slicing program
(currently slic3r) through an HTTP API.  A command line slicing tool is
provided for ease of use and to support eventual integration with host software
like Repetier-Host and OctoPrint.

Matching Snuggies is well suited for integration with host-software that may
run in a Resource Constrained environment.

Documentation
=============

Install
-------

First install [slic3r](http://slic3r.org/download), the slic3r Matching
Snuggies has chosen to support initially.

NOTE: OS X users should symlink the executable at Slicer.app/MacOS/slicer into
their environment's PATH.

./build.sh

Slicing API
-----------

A REST API is exposed to schedule slicing jobs, retrieve resulting gcode, and
get periodic status updates while slicing is in progress.

```
./bin/snuggied -slic3r.configs=testdata
```

Set the API [doc](API.md) for information about each endpoint.

Command line tool
-----------------

After starting, the daemon can be sent files to slice using the command line
tool.

```
./bin/snuggier -preset=hq -o FirstCube.gcode testdata/FirstCube.amf
```

Goals
-----

- daemon exposing slic3r over HTTP (authenticated)
- integration with other backend slicers (Cura)
- a slicing queue that may be consumed by a pool of workers (shared
  configuration; dropbox?)
- cluster health/monitoring dashboard
