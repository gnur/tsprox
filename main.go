package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/sethvargo/go-envconfig"
	"tailscale.com/client/tailscale"
	"tailscale.com/ipn"

	"tailscale.com/tsnet"
)

type config struct {
	TailscaleControlHost string `env:"TS_HOST"`
	ClientID             string `env:"OAUTH_CLIENT_ID"`
	ClientSecret         string `env:"OAUTH_CLIENT_SECRET"`
	TailnetName          string `env:"TAILNET_NAME"`
	HostName             string `env:"HOSTNAME"`
	ProxyHost            string `env:"PROXY_HOST"`
	Verbose              bool   `env:"VERBOSE"`
}

var tsClient *tailscale.LocalClient

func main() {

	ctx := context.Background()
	var cfg config
	if err := envconfig.Process(ctx, &cfg); err != nil {
		log.Fatal(err)
	}

	tsToken, err := getAuthToken(cfg.ClientID, cfg.ClientSecret, cfg.TailnetName)
	if err != nil {
		log.Fatal(err)
	}

	srv := &tsnet.Server{
		ControlURL: cfg.TailscaleControlHost,
		Hostname:   cfg.HostName,
		AuthKey:    tsToken,
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

	http.HandleFunc("/", proxy.ServeHTTP)

	if cfg.TailscaleControlHost == "" {
		cfg.TailscaleControlHost = ipn.DefaultControlURL
	}

	tsClient, _ = srv.LocalClient()
	l80, err := srv.Listen("tcp", ":80")
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Serving http://%s/ ...", cfg.HostName)
	if err := http.Serve(l80, nil); err != nil {
		log.Fatal(err)
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
