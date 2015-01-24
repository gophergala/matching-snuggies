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

Slicing API
-----------

A REST API is exposed to schedule slicing jobs, retrieve resulting gcode,
and get periodic status updates while slicing is in progress.

```
snuggied -http=:8080
```

Host Integration
----------------

```
matching-snuggies -preset=hq-printrbot -o composition.gcode composition.amf
```

Goals
-----

- daemon exposing slic3r over HTTP (authenticated)
- "slicing program" -- a client that acts as a normal slicer would
  (STL/AMF in, G-code out)
- a slicing queue that may be consumed by a pool of workers (shared
  configuration; dropbox?)
- cluster health/monitoring dashboard
