GeoIP Tool
==========

# Purpose

This tool can be used to create a [MMDB database](https://maxmind.github.io/MaxMind-DB/) which can be used to feed custom IP adresse into the [Elasticsearch GeoIP Processor](https://www.elastic.co/guide/en/elasticsearch/reference/current/geoip-processor.html).

# Building

## Docker

Use docker to built a image containing the tool as `/go/bingeoip-tool`.

```
docker buildx build . -t ghcr.io/cmahnke/geoip-tool:latest
```

## Golang

User your local [Go](https://go.dev/) installation:

```
go build
```

# Usage

## Options

```
/usr/local/bin/geoip-tool -d "" -i /tmp/ips.ndjson -o /usr/share/elasticsearch/config/ingest-geoip/fowi.mmdb
```

| Parameter       | Default               | Description                                                                                                    |
|-----------------|-----------------------|----------------------------------------------------------------------------------------------------------------|
| -d / --database | GeoLite2-Country.mmdb | Update existing database, use "" if new database should be created                                             |
| -i / --ips      | ips.ndjson              | Important: **The tool uses the NDJSON format as input**, see below                                           |
| -o / --output   | geoip.mmdb            | File name of output file                                                                                       |
| -t / --type     | GeoIP-City            | Database type, required by the Elasticsearch GeoIP processor, make sure to use a suffix as provided by MaxMind |

## Input format

The input file format is [NDJSON](http://ndjson.org/). Each line should be it's own JSON document.

```
{"ip": "", "name": "", "lat": "", "lon":"", "floor": "", "accuracy_radius": ""}
```

| Field           | Type   | Description                                                                           |
|-----------------|--------|---------------------------------------------------------------------------------------|
| ip              | String | IP adress or sub net in CIDR notation, if no host identifier is given `/32`is assumed |
| name            | String | Name of IP or sub net                                                                 |
| lat             | Number | Lattitute                                                                             |
| lon             | Number | Longitute                                                                             |
| floor           | String | Floor / Name of Floor                                                                 |
| accuracy_radius | Number | Accuracy radius                                                                       |

## As a Elasticsearch pipeline

Copy the output file to the `/config/ingest-geoip` of Elasticsearch, create the subdirectory if it doesn't exists.

```
{
  "processors": [
    {
      "geoip": {
        "field": "request.ip",
        "ignore_missing": true,
        "ignore_failure": true,
        "database_file": "geoip.mmdb",
        "properties": ["region_name", "region_iso_code", "location", "city_name", "country_name"]
      }
    }
  ]
}
```

## Running

Using Docker:

```
docker run -v "`pwd`:`pwd`" -w "`pwd`" ghcr.io/cmahnke/geoip-tool:latest /usr/local/bin/geoip-tool [args]
```

See above for `[args]`.

## Query the database

One can use the [`mmdbinspect`](https://github.com/maxmind/mmdbinspect) tool to query the database.

### Installation

```
go install github.com/maxmind/mmdbinspect/cmd/mmdbinspect@latest
```

### Query

```
$(go env GOPATH)/bin/mmdbinspect -db geoip.mmdb 1.2.3.4
```

Replace `1.2.3.4` with the IP adress to query