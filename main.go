package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"github.com/sethvargo/go-envconfig"
	"tailscale.com/client/tailscale"
	"tailscale.com/ipn"

	"tailscale.com/tsnet"
)

type config struct {
	TailscaleControlHost string   `env:"TS_HOST"`
	ClientID             string   `env:"OAUTH_CLIENT_ID"`
	ClientSecret         string   `env:"OAUTH_CLIENT_SECRET"`
	TailscaleToken       string   `env:"TAILSCALE_TOKEN"`
	TailnetName          string   `env:"TAILNET_NAME"`
	TailscaleTags        []string `env:"TAILSCALE_TAGS"`
	HostName             string   `env:"HOSTNAME"`
	ProxyHost            string   `env:"PROXY_HOST"`
	Verbose              bool     `env:"VERBOSE"`

	EnableCapture bool `env:"ENABLE_CAPTURE"`
	MaxCaptures   int  `env:"MAX_CAPTURES"`
}

var tsClient *tailscale.LocalClient

func main() {

	ctx := context.Background()
	var cfg config
	err := envconfig.Process(ctx, &cfg)
	if err != nil {
		log.Fatal(err)
	}

	tsToken := cfg.TailscaleToken
	if tsToken == "" {
		tsToken, err = getAuthToken(cfg.ClientID, cfg.ClientSecret, cfg.TailnetName, cfg.TailscaleTags)
		if err != nil {
			log.Fatal(err)
		}
	}

	if cfg.TailscaleControlHost == "" {
		cfg.TailscaleControlHost = ipn.DefaultControlURL
	}

	srv := &tsnet.Server{
		ControlURL: cfg.TailscaleControlHost,
		Hostname:   cfg.HostName,
		AuthKey:    tsToken,
		Ephemeral:  true,
		Logf:       func(format string, args ...any) {},
	}
	if cfg.Verbose {
		srv.Logf = log.Printf
	}

	url, err := url.Parse(cfg.ProxyHost)
	if err != nil {
		log.Fatal(err)
	}

	proxy := &httputil.ReverseProxy{
		Director: func(r *http.Request) {
			u, _ := currentUser(r)
			r.Header.Add("X-Forwarded-Host", url.Hostname())
			r.Header.Add("X-Origin-Host", cfg.HostName)
			r.Header.Add("X-Origin-User", u)
			r.Host = url.Host
			r.URL.Host = url.Host
			r.URL.Scheme = url.Scheme

		}, Transport: &http.Transport{
			TLSHandshakeTimeout: 10 * time.Second,
			IdleConnTimeout:     90 * time.Second,
			MaxIdleConns:        100,
			Dial: (&net.Dialer{
				Timeout:   6 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
		}}

	capSrv := NewCaptureService(cfg.MaxCaptures)
	hdr := NewRecorderHandler(capSrv, proxy.ServeHTTP)

	tsClient, _ = srv.LocalClient()
	l80, err := srv.Listen("tcp", ":80")
	if err != nil {
		log.Fatal(err)
	}

	if cfg.EnableCapture && cfg.MaxCaptures > 0 {
		go func() {
			l81, err := srv.Listen("tcp", ":81")
			if err != nil {
				log.Fatal(err)
			}
			if err := http.Serve(l81, NewDashboardHandler(hdr, capSrv, cfg)); err != nil {
				log.Fatal(err)
				os.Exit(1)
			}
		}()
	} else {
		hdr = proxy.ServeHTTP
	}

	log.Printf("Serving http://%s/ ...", cfg.HostName)
	if err := http.Serve(l80, hdr); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	fmt.Println("Hello, World!")
}

func currentUser(r *http.Request) (string, error) {
	login := ""
	res, err := tsClient.WhoIs(r.Context(), r.RemoteAddr)
	if err != nil {
		return "", err
	}
	login = res.UserProfile.LoginName

	return login, nil
}
