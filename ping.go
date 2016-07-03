package multiping

import (
	"net/http"
	"net/url"
	"time"
	"net"
	"fmt"
	"github.com/tatsushid/go-fastping"
	"os"
)

type Multiping struct {
	Timeout int32     // Timeout in Seconds
	URL     []url.URL // URL to Ping
}

type MultipingResult struct {
	err error
}

func PingOnURL(urlToUse url.URL) error {
	switch urlToUse.Scheme {
	case "icmp": {
		pinger := fastping.NewPinger()

		ra, err := net.ResolveIPAddr("ip4:icmp", urlToUse.Host)

		if err != nil {
			return err
		}

		pinger.AddIPAddr(ra)

		pinger.MaxRTT = time.Second * 15

		returnError := fmt.Errorf("Timed out")

		pinger.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
			returnError = nil
		}

		err = pinger.Run()

		if err != nil {
			return err
		} else {
			return returnError
		}
	}
	case "tcp": {
		conn, err := net.DialTimeout("tcp4", urlToUse.Host, 15 * time.Second)

		if nil != err {
			return err
		}

		conn.Close()

		return nil
	}
	case "http", "https": {
		client := &http.Client{}
		client.Timeout = 15 * time.Second
		resp, err := client.Get(urlToUse.String())

		if err != nil {
			fmt.Fprintf(os.Stderr, "%s / Err: %#v / %#v\n", urlToUse.String(), err, err.Error())

			return err
		} else {
			fmt.Fprint(os.Stderr, "%s / Status Text: %s\n", urlToUse.String(), resp.Status)
			if resp.StatusCode >= 400 {
				return fmt.Errorf("Server returned status code '%d'", resp.StatusCode)
			} else if resp.StatusCode < 400 {
				return nil
			}
		}
	}
	}
	return fmt.Errorf("Unreachable code")
}

func (mp*Multiping) RunPingLoop() bool {
	results := make(map[url.URL]error)

	for _, v := range mp.URL {
		results[v] = nil
	}

	timeoutsAt := time.Now().Add(time.Duration(mp.Timeout) * time.Second)

	done := false

	for !done {
		for k, _ := range results {
			results[k] = PingOnURL(k)

			fmt.Printf("%s: %#v\n", k.String(), results[k])

			if nil == results[k] {
				delete(results, k)
			}
		}

		done = (0 == len(results)) || (time.Now().After(timeoutsAt))

		if (! done) {
			time.Sleep(time.Second)
		}
	}

	if (0 == len(results)) {
		return false
	} else {
		return true
	}
}
