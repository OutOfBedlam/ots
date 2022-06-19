pname="ots"

osm-data-source="./tmp/south-korea-2022-04-18.osm.pbf"
// osm-data-source="tcp://127.0.0.1:1918"
bind="127.0.0.1"
port=1919

/////// redering server
cache-size=2000
show-watermark = true
show-labels = true

grpc {
    max-recv-msg-size=100
    max-send-msg-size=100
}

log {
    console=true
    filename="-"
    default-prefix-width=10
    default-level="TRACE"
}

http {
    console-color=true
    debug-mode=true
}

httplog {
    console=true
    filename="-"
    default-prefix-width=10
    default-level="TRACE"
}