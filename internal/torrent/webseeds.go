package torrent

import "net/url"

func supportedWebSeeds(raw []string) []string {
	var seeds []string
	for _, address := range raw {
		u, err := url.Parse(address)
		if err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Hostname() != "" && u.User == nil {
			seeds = append(seeds, address)
		}
	}
	return seeds
}
