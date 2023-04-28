# TSPROX

tsprox is a proxy designed to be running as sidecar container within kubernetes. It will will add itself to your tailscale tailnet to provide a easily reachable address for you and other services living within your tailnet.

By design, it becomes a unique host in your tailnet, it will always listen on port 80.

## motivation

For my side projects I often don't need publicly reachable addresses, but it is really useful to be able to reach services running inside a kubernetes cluster. Or even between clusters. tsprox allows you to instantly have resolvable "magic" dns names for those services.

## authentication

It has 2 ways of authenticating itself to tailscale.

1. With an oauth client id and client secret, this will allow tsprox to provision a auth key at runtime. This auth key will be provisioned as ephemeral and expires after 30 seconds.
1. With a tailscale auth key you create yourself.

The advantage of the first method is that the client id and secret won't expire. The disadvantage is that you will have to provide tsprox with quite some broad permissions.  
The second approach allows you to use an auth key you provision yourself.

## identity aware proxy

tsprox resolves all ip addresses to actual "users". The resolved "user" (tailscale machine name or tags) will be available to the proxied service in the `X-Origin-User` header.


## captures

tsprox can optionally also expose a web interface that shows received requests and responses. This web interface is reachable on port `${HOSTNAME}:81`

## configuration

The following env vars configure tsprox:

```bash
HOSTNAME=tsprox-dev #"machine" name that will be the magic dns name within tailscale
VERBOSE=true
PROXY_HOST=http://localhost:8008 # address to proxy requests to, needs to be with protocol
ENABLE_CAPTURE=false # wether to enable the web requests capture funcionality that will be available at http://${HOSTNAME}:81
MAX_CAPTURES=10 # amount of requests to keep in memory

# to use OAUTH
OAUTH_CLIENT_ID=xxxxxxxCNTRL 
OAUTH_CLIENT_SECRET=tskey-client-xxxxxxxCNTRL-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
TAILNET_NAME=xxxxxxx #name of your tailnet, shown as "organization" in tailscale General Settings
TAILSCALE_TAGS=tag:service # tags to automatically assign to your tsprox instance

# to use auth key
TAILSCALE_TOKEN=tskey-auth-xxxxxxxCNTRL-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

## sidecar container config

Just add the following snippet to any deployment:

```yaml
    containers:
    - image: ghcr.io/gnur/tsprox:v0.9-1-geb13476
      imagePullPolicy: IfNotPresent
      name: tsprox
      envFrom:
        - configMapRef:
            name: tsprox-config
        - secretRef:
            name: tailscale-credentials
```

And also create the configmap and secret off course.
