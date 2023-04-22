package main

import (
	"flag"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/bfontaine/jsons"
	"github.com/maxmind/mmdbwriter"
	"github.com/maxmind/mmdbwriter/inserter"
	"github.com/maxmind/mmdbwriter/mmdbtype"
)

var (
	writer *mmdbwriter.Tree
	suffix = "/32"
)

func main() {

	database := flag.String("d", "GeoLite2-Country.mmdb", "database")
	ips := flag.String("i", "ips.ndjson", "ips")
	output := flag.String("o", "geoip.mmdb", "output")
	dbtype := flag.String("t", "GeoIP-City", "type")

	flag.Parse()

	//Init source
	var err error
	if *database != "" {
		// Load the database we wish to enrich.
		writer, err = mmdbwriter.Load(*database, mmdbwriter.Options{IncludeReservedNetworks: true})

	} else {
		writer, err = mmdbwriter.New(mmdbwriter.Options{IncludeReservedNetworks: true, IPVersion: 4, DatabaseType: *dbtype})
	}
	if err != nil {
		log.Fatal(err)
	} else {
		err = nil
	}

	// Read IPs
	reader := jsons.NewFileReader(*ips)
	if err := reader.Open(); err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	for {
		var entry map[string]string
		if err := reader.Next(&entry); err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}
		var ipv4Net *net.IPNet
		var err error
		var floor string
		if !strings.Contains(entry["ip"], "/") {
			_, ipv4Net, err = net.ParseCIDR(entry["ip"] + suffix)
		} else {
			_, ipv4Net, err = net.ParseCIDR(entry["ip"])
		}
		if err != nil {
			log.Fatal(err)
		}
		lat, err := strconv.Atoi(entry["lat"])
		lon, err := strconv.Atoi(entry["lon"])

		if val, ok := entry["floor"]; ok {
			floor = val
		}

		locData := mmdbtype.Map{
			"continent_name":   mmdbtype.String("Europe"),
			"continent_code":   mmdbtype.String("EU"),
			"country_name":     mmdbtype.String("Germany"),
			"country_iso_code": mmdbtype.String("DE"),
			"city_name":        mmdbtype.String("Göttingen"),
			"name":             mmdbtype.String(entry["name"]),
			"latitude":         mmdbtype.Float64(lat),
			"longitude":        mmdbtype.Float64(lon),
			"floor":            mmdbtype.String(floor),
		}

		err = writer.InsertFunc(ipv4Net, inserter.TopLevelMergeWith(locData))
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Inserted %s, name '%s': {lat: %f, lon:%f}, floor: %s", ipv4Net, entry["name"], lat, lon, floor)
	}

	// Write results
	fh, err := os.Create(*output)
	if err != nil {
		log.Fatal(err)
	}
	_, err = writer.WriteTo(fh)
	if err != nil {
		log.Fatal(err)
	}

}
