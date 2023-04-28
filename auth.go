package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"golang.org/x/oauth2/clientcredentials"
)

type createTokenBody struct {
	Capabilities struct {
		Devices struct {
			Create struct {
				Reusable      bool     `json:"reusable"`
				Ephemeral     bool     `json:"ephemeral"`
				Preauthorized bool     `json:"preauthorized"`
				Tags          []string `json:"tags"`
			} `json:"create"`
		} `json:"devices"`
	} `json:"capabilities"`
	ExpirySeconds int `json:"expirySeconds"`
}

type createTokenResponse struct {
	ID           string    `json:"id"`
	Key          string    `json:"key"`
	Created      time.Time `json:"created"`
	Expires      time.Time `json:"expires"`
	Capabilities struct {
		Devices struct {
			Create struct {
				Reusable      bool     `json:"reusable"`
				Ephemeral     bool     `json:"ephemeral"`
				Preauthorized bool     `json:"preauthorized"`
				Tags          []string `json:"tags"`
			} `json:"create"`
		} `json:"devices"`
	} `json:"capabilities"`
}

func getAuthToken(clientid, clientsecret, tailnet string) (string, error) {
	var oauthConfig = &clientcredentials.Config{
		ClientID:     clientid,
		ClientSecret: clientsecret,
		TokenURL:     "https://api.tailscale.com/api/v2/oauth/token",
	}

	client := oauthConfig.Client(context.Background())

	endpoint := fmt.Sprintf("https://api.tailscale.com/api/v2/tailnet/%s/keys", tailnet)

	req := createTokenBody{}
	//sometimes the token gets revoked prematurely somehow so allow reuse
	req.Capabilities.Devices.Create.Reusable = true
	req.Capabilities.Devices.Create.Ephemeral = true
	req.Capabilities.Devices.Create.Preauthorized = true
	req.Capabilities.Devices.Create.Tags = []string{"tag:service"}
	req.ExpirySeconds = 30

	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("error marshalling req body: %w", err)
	}

	resp, err := client.Post(endpoint, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Fatalf("error getting keys: %v", err)
	}

	var token createTokenResponse
	err = json.NewDecoder(resp.Body).Decode(&token)
	if err != nil {
		return "", fmt.Errorf("error unmarshalling response body: %w", err)
	}

	return token.Key, nil
}
