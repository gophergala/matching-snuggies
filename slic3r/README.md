This directory contains Slic3r configuration files in the INI format natively
exported by Slic3r.

The `snuggied` server will look for INI files in a directory and expose them as
"presets" that clients can use to control the quality and speed of their
prints. For example the "default.ini" and "hq.ini" files in this directory will
be exported as "default" and "hq".

    curl http://localhost:8888/slicer/jobs -F slicer=slic3r -F preset=default -F meshfile=@testdata/FirstCube.stl
    curl http://localhost:8888/slicer/jobs -F slicer=slic3r -F preset=hq -F meshfile=@testdata/FirstCube.stl

In the slic3r GUI you can save your current settings as an INI file by
selecting menu options "File" > "Export Config...".
