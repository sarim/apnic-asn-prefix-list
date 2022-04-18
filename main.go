package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type ASN struct {
	AS       string
	prefixes []string
}

var routeRe = regexp.MustCompile(`oute:\s+([\d./]+)`)
var asRe = regexp.MustCompile(`origin:\s+(AS\d+)`)

func getAsnMap(url string) (map[string]*[]*ASN, map[string]*ASN) {
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	asnMapByCountry := make(map[string]*[]*ASN)
	asnMapByAS := make(map[string]*ASN)

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if line[0] == '#' {
			continue
		}
		parts := strings.Split(scanner.Text(), "|")
		if parts[2] == "asn" {
			country := parts[1]
			asn := "AS" + parts[3]
			arr, ok := asnMapByCountry[country]

			_ASN := &ASN{asn, []string{}}
			if ok {
				*arr = append(*arr, _ASN)
			} else {
				asnMapByCountry[country] = &[]*ASN{_ASN}
			}
			asnMapByAS[asn] = _ASN
		}
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}

	return asnMapByCountry, asnMapByAS
}

func mergeDbRoute(url string, asnMap map[string]*ASN) {
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	gzipReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(gzipReader)

	split := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if i := bytes.Index(data, []byte("route:")); i >= 0 {
			return i + 1, data[0:i], nil
		}

		if atEOF {
			return len(data), data, nil
		}
		// Request more data.
		return 0, nil, nil
	}

	scanner.Split(split)
	i := 0
	if scanner.Scan() {
		for scanner.Scan() {
			chunk := scanner.Text()
			i++

			routeSubmatch := routeRe.FindStringSubmatch(chunk)
			asSubmatch := asRe.FindStringSubmatch(chunk)

			if len(routeSubmatch) == 2 && len(asSubmatch) == 2 {
				prefix := routeSubmatch[1]
				as := asSubmatch[1]

				asn, ok := asnMap[as]
				if ok {
					asn.prefixes = append(asn.prefixes, prefix)
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}

}

func printPrefixes(asn *ASN) {
	for _, prefix := range asn.prefixes {
		fmt.Println(prefix)
	}
}

func main() {

	as := flag.String("as", "", "provide an ASN to print its ip prefixes")
	country := flag.String("country", "", "provide a country to all ASNs' ip prefixes in this country")

	flag.Parse()

	if *as == "" && *country == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *as != "" && *country != "" {
		fmt.Fprintln(os.Stderr, "Please provide only one argument")
		flag.PrintDefaults()
		os.Exit(1)
	}

	asnMapByCountry, asnMapByAS := getAsnMap("http://ftp.apnic.net/apnic/stats/apnic/delegated-apnic-latest")
	_ = asnMapByCountry
	mergeDbRoute("http://ftp.apnic.net/apnic/whois/apnic.db.route.gz", asnMapByAS)

	if *country != "" {
		if asnArr, ok := asnMapByCountry[*country]; ok {
			for _, asn := range *asnArr {
				printPrefixes(asn)
			}
		} else {
			fmt.Fprintln(os.Stderr, "country not found")
		}
	}

	if *as != "" {
		if asn, ok := asnMapByAS[*as]; ok {
			printPrefixes(asn)
		} else {
			fmt.Fprintln(os.Stderr, "ASN not found")
		}
	}
}
