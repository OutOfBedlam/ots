
# OpenStreetMap Tile Server (OTS)

OTS is a server that loads OpenStreetMap data file (`*.osm.pbf`) into memory then generates map tiles for HTTP requests.

![](docs/screenshot_00.jpg)

## Compile & Packaging

- checkout source codes
```sh
$ git clone https://github.com/OutOfBedlam/ots.git

$ cd ots
```

- compiled executable binary will be created in './tmp'

```sh
$ make
```

- create package as `./packages/ots-{version}-{os}-{arch}.zip`

```sh
$ make package
```

### Start OTS as tile-rendering-server and OSM data-server in a process

```
./tmp/ots server -p 1918 -i ./tmp/my-area.osm.pbf 
```

```mermaid
sequenceDiagram
autonumber
   Note over Browser, OTS: HTTP (ex: leaflet.js)
   Browser ->>+ OTS : http://server/{z}/{x}/{y}.png
   OTS->>OTS: Convert coords
   OTS->>OTS: Retrieve map data
   OTS->>OTS: Gen image
   OTS ->>- Browser: .png
```

### Start tile-rendering-server and data-server

- start a process as a data-server

```
./tmp/ots server -p 1918 -i ./tmp/my-area.osm.pbf 
```

- start tile-rendering-server

> Specify the endpoint of data-server instead of `*.osm.pbf` file with `-i`

```
./tmp/ots server -p 1919 -i tcp://127.0.0.1:1918
```

- then web browser can retrieve map tiles as following flow

`Web browser` --> `127.0.0.1:1919` (rendering server) --> `127.0.0.1:1918` (data server)


```mermaid
sequenceDiagram
autonumber
   Note over Browser, Render-Server: HTTP (ex: leaflet.js)
   Browser ->>+ Render-Server : http://server/{z}/{x}/{y}.png
   Render-Server->>Render-Server: Convert coords
   Render-Server ->>+ Data-Server : request map data with lat/lon bounds
   Note over Render-Server, Data-Server: GRPC
   Data-Server ->>- Render-Server: map data (osm features)
   Render-Server->>Render-Server: Gen image
   Render-Server ->>- Browser: .png
```

### 서버 설정 파일로 구동하는 방법

`server-config-sample.hcl`을 수정하여 설정파일을 생성하고 `*.hcl` 확장자로 저장하고, `-c` 옵션으로 경로를 지정한다.

```
./tmp/ots server -c <config_file.hcl>
```

*config file items*

| 항목          | 설명         | 예   | 비고|
| -------------| ------------| --- | -------- |
| `pname`      | process instance name | `"ots01"` | |
| `osm-data-source` | data file or data server address | `"./tmp/south-korea-2022-04-18.osm.pbf"` `"tcp://127.0.0.1:1918"` |  |
| `bind`       | listening address | `"127.0.0.1"` | |
| `port`       | listening port    | 1919 | |
| `grpc.max-recv-msg-size` | grpc limit (MB) | 100 | |
| `grpc.max-send-msg-size` | grpc limit (MB) | 100 | |
| `log.console`       | log output to console       | `true` 또는 `false` | |
| `log.filename`      | log file path               |     | |
| `log.default-prefix-width` |            | 10 | |
| `log.default-level` | log level         | `"TRACE"`| | 
| `http.console-color`| colored http log | `true` 또는 `false` | |
| `http.debug-mode`   | debug http log   | `true` 또는 `false` | |
| `httplog.console`   |||
| `httplog.filename`  |||
| `httplog.default-prefix-width`|||
| `httplog.default-level`|||

*rendering server only items*

| 항목          | 설명         | 예   | 비고|
| -------------| ------------| --- | -------- |
| `cache-size` | size of lru cache | 2000 | |
| `show-watermark` | watermark (tile coordinates) on tiles | true false | |
| `show-labels` | enable labels | true false | |


### Where to download `*.osm.pdf` files

한국 지역 osm.pbf 파일 다운로드
- https://download.geofabrik.de/asia/south-korea.html
- OSM data up to 2022-04-17T20:21:54Z. File size: 148 MB;

- https://wiki.openstreetmap.org/wiki/Planet.osm#Country_and_area_extracts

### References about `osm.pbf` files

- https://wiki.openstreetmap.org/wiki/Elements
- https://wiki.openstreetmap.org/wiki/PBF_Format

- Nodes     : https://wiki.openstreetmap.org/wiki/Node
- Ways      : https://wiki.openstreetmap.org/wiki/Way
- Relations : https://wiki.openstreetmap.org/wiki/Relation