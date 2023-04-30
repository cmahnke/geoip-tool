package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/maxmind/mmdbwriter"
	"github.com/maxmind/mmdbwriter/inserter"
	"github.com/maxmind/mmdbwriter/mmdbtype"
)

var (
	writer                        *mmdbwriter.Tree
	suffix                                = "/32"
	name_field                            = "city_name"
	floor_as_timezone                     = true
	default_continent_name                = "Europe"
	default_continent_geonames_id         = 6255148
	default_continent_code                = "EU"
	default_country_name                  = "Germany"
	default_country_geonames_id           = 2921044
	default_country_code                  = "DE"
	default_is_in_european_union          = true
	default_city_name                     = "Göttingen"
	default_city_geoname_id               = 2918632
	default_lat                   float64 = 51.5441
	default_lon                   float64 = 9.9254
	default_accuracy              float32 = 5
)

type Ip struct {
	Ip          string     `json:"ip"`
	Name        string     `json:"name"`
	Floor       string     `json:"floor"`
	Accuracy    float32    `json:"accuracy_radius"`
	Lon         float64    `json:"lon"`
	Lat         float64    `json:"lat"`
	Coordinates [2]float64 `json:"coordinates"`
}

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
		description := map[string]string{"en": "Simple Database for GeoIP lookups"}
		writer, err = mmdbwriter.New(mmdbwriter.Options{IncludeReservedNetworks: true, IPVersion: 4, DatabaseType: *dbtype, Languages: []string{"en", "de"}, Description: description})
	}
	if err != nil {
		log.Fatal(err)
	} else {
		err = nil
	}

	reader, err := os.Open(*ips)
	if err != nil {
		log.Fatal(err)
	}

	defer reader.Close()

	decoder := json.NewDecoder(reader)
	for decoder.More() {
		ip := Ip{Accuracy: default_accuracy, Coordinates: [2]float64{default_lon, default_lat}}
		if err := decoder.Decode(&ip); err != nil {
			log.Fatal(err)
		}

		var ipv4Net *net.IPNet
		var err error
		if !strings.Contains(ip.Ip, "/") {
			_, ipv4Net, err = net.ParseCIDR(ip.Ip + suffix)
		} else {
			_, ipv4Net, err = net.ParseCIDR(ip.Ip)
		}
		if err != nil {
			log.Fatal(err)
		}

		location := mmdbtype.Map{}
		if ip.Lon != 0 && ip.Lat != 0 {
			location["longitude"] = mmdbtype.Float64(ip.Lon)
			location["latitude"] = mmdbtype.Float64(ip.Lat)
		} else {
			location["longitude"] = mmdbtype.Float64(ip.Coordinates[0])
			location["latitude"] = mmdbtype.Float64(ip.Coordinates[1])
		}

		if floor_as_timezone {
			location["time_zone"] = mmdbtype.String(ip.Floor)
		}

		location["accuracy_radius"] = mmdbtype.Int32(ip.Accuracy)

		locData := mmdbtype.Map{
			"continent": mmdbtype.Map{
				"geoname_id": mmdbtype.Uint32(default_continent_geonames_id),
				"code":       mmdbtype.String(default_continent_code),
				"names": mmdbtype.Map{
					"en": mmdbtype.String(default_continent_name),
				},
			},

			"country": mmdbtype.Map{
				"geoname_id":           mmdbtype.Uint32(default_country_geonames_id),
				"iso_code":             mmdbtype.String(default_country_code),
				"is_in_european_union": mmdbtype.Bool(default_is_in_european_union),
				"names": mmdbtype.Map{
					"en": mmdbtype.String(default_country_name),
				},
			},

			"city": mmdbtype.Map{
				"geoname_id": mmdbtype.Uint32(default_city_geoname_id),
				"names": mmdbtype.Map{
					"en": mmdbtype.String(default_city_name),
					"de": mmdbtype.String(default_city_name),
				}},

			"name":     mmdbtype.String(ip.Name),
			"location": location,
			"subdivisions": mmdbtype.Slice{
				mmdbtype.Map{
					"iso_code": mmdbtype.String(fmt.Sprintf("%2s", ip.Floor)),
					"name":     mmdbtype.String(ip.Name),
					"names": mmdbtype.Map{
						"en": mmdbtype.String(ip.Name),
						"de": mmdbtype.String(ip.Name),
					}}},
			"floor": mmdbtype.String(ip.Floor),
		}

		err = writer.InsertFunc(ipv4Net, inserter.TopLevelMergeWith(locData))
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Inserted %s, name '%s': {lat: %f, lon:%f}, floor: %s, accuracy_radius: %f", ipv4Net, ip.Name, ip.Lat, ip.Lon, ip.Floor, ip.Accuracy)
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
