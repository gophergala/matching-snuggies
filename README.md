Matching Snuggies
=================

A remote 3D printing slicer.

Remote Slicing
--------------

Matching Snuggies exposes a backend slicing program (currently slic3r) via
an HTTP API. It also provides a command line slicing tool that can
interface with existing host software like Repetier-Host.

The intended target for Matching Snuggies is to integrate with OctoPrint,
to make resource constrained devices like Raspberry Pi more practical as a
host device.  However I do not think it is immediately possible.

Documentation
=============

Install
-------

First install [slic3r](http://slic3r.org/download), the slic3r Matching
Snuggies has chosen to support initially.

./build.sh

Slicing API
-----------

A REST API is exposed to schedule slicing jobs, retrieve resulting gcode, and
get periodic status updates while slicing is in progress.

```
./bin/snuggied -slic3r.configs=testdata
```

Set the API [doc](API.md) for information about each endpoint.

Host Integration
----------------

```
./bin/snuggier -preset=hq -o FirstCube.gcode testdata/FirstCube.amf
```

Goals
-----

- daemon exposing slic3r over HTTP (authenticated)
- "slicing program" -- a client that acts as a normal slicer would
  (STL/AMF in, G-code out)
- a slicing queue that may be consumed by a pool of workers (shared
  configuration; dropbox?)
- cluster health/monitoring dashboard
