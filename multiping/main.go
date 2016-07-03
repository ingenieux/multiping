package main

import (
	"fmt"
	"github.com/docopt/docopt-go"
	"github.com/ingenieux/multiping"
	"os"
	"net/url"
	"strconv"
	"strings"
)

const USAGE = `
multiping.

Usage:
  multiping [--timeout=<timeout>] URL ...

Options:
  -t --timeout=<timeout>  Timeout in seconds [default: 300]

# NOTE: Needs sysctl net.ipv4.ping_group_range = 0   2147483647 for icmp to work
`

func main() {
	arguments, err := docopt.Parse(USAGE, nil, true, "multiping 1.0", true)

	if nil != err {
		fmt.Fprintf(os.Stderr, "Usage error: %s", err)

		os.Exit(127)
	}

	//fmt.Fprintf(os.Stderr, "arguments: %#v\n", arguments)

	urlArguments := arguments["URL"].([]string)

	mp := &multiping.Multiping{}

	timeout, err := strconv.Atoi(arguments["--timeout"].(string))

	if nil != err {
		fmt.Fprintf(os.Stderr, "Invalid timeout value '%s'\n", arguments["--timeout"].(string))

		os.Exit(127)
	}

	mp.Timeout = int32(timeout)

	mp.URL = make([]url.URL, len(urlArguments))

	i := 0
	for _, v := range urlArguments {
		newUrl, err := url.Parse(v)

		if nil != err {
			fmt.Fprintf(os.Stderr, "Error: %s", err)

			os.Exit(127)
		}

		switch newUrl.Scheme {
		case "icmp": break
		case "tcp": {
			hasError := false

			elements := strings.Split(newUrl.Host, ":")

			hasError = nil != err

			if (! hasError) {
				hasError = (2 != len(elements))
			}

			if (! hasError) {
				_, err := strconv.Atoi(elements[1])

				hasError = (nil != err)
			}

			if hasError {
				fmt.Fprintf(os.Stderr, "Invalid tcp host spec '%s'. Must be tcp://host:port/\n", newUrl.Host)

				os.Exit(127)
			}
		}
		case "http":
		case "https": break
		default:
			fmt.Fprintf(os.Stderr, "Unknown url scheme '%s' for URL '%s'\n", newUrl.Scheme, v)
			os.Exit(127)
		}

		mp.URL[i] = *newUrl

		i++
	}

	//fmt.Printf("mp: %#v\n", mp)

	result := mp.RunPingLoop()

	if (result) {
		os.Exit(128)
	} else {
		os.Exit(0)
	}
}