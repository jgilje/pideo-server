# pideo-server
Raspberry Pi video server. The server will capture video using raspivid, wrapped in a MPEG-TS stream
and served to your network. It registers itself using zeroconf, so it should be auto-detectable by clients.

## Connecting
* An android application for connecting to the server is available at http://github.com/jgilje/pideo
* Use vlc - `cvlc tcp://server:12345`

## Binaries
Prebuilt binaries for linux/arm available at http://jgilje.net/pideo/

## Dependencies
`sudo apt-get install gstreamer1.0-tools gstreamer1.0-plugins-bad`
