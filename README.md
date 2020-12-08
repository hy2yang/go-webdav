# webdav

As of golang.org/x/net v0.0.0-20200822124328-c89045814202 there's [a bug in the webdav implementation](https://github.com/golang/go/issues/16195), fix at https://go-review.googlesource.com/c/net/+/249797.

To incoporate the fix before release:
`go mod vendor`
update the verder go file following the fix
`go build -mod=vendor`

## Install

Please refer to the [Releases page](https://github.com/hy2yang/go-webdav/releases) for more information. There, you can either download the binaries or find the Docker commands to install WebDAV.

## Usage

```webdav``` command line interface is really easy to use so you can easily create a WebDAV server for your own user. By default, it runs on a random free port and supports JSON, YAML and TOML configuration. An example of a YAML configuration with the default configurations:

```yaml
# Server related settings
address: 0.0.0.0
port: 0
auth: true
tls: false
cert: cert.pem
key: key.pem

# Default user settings (will be merged)
scope: .
modify: true
rules: []

# CORS configuration
cors:
  enabled: true
  credentials: true
  allowed_headers:
    - Depth
  allowed_hosts:
    - http://localhost:8080
  allowed_methods:
    - GET
  exposed_headers:
    - Content-Length
    - Content-Range

users:
  - username: admin
    password: admin
    scope: /a/different/path
  - username: encrypted
    password: "{bcrypt}$2y$10$zEP6oofmXFeHaeMfBNLnP.DO8m.H.Mwhd24/TOX2MWLxAExXi4qgi"
  - username: "{env}ENV_USERNAME"
    password: "{env}ENV_PASSWORD"
  - username: basic
    password: basic
    modify:   false
    rules:
      - regex: false
        allow: false
        path: /some/file
```

There are more ways to customize how you run WebDAV through flags and environment variables. Please run `webdav --help` for more information on that.

### Systemd

An example of how to use this with `systemd` is on [webdav.service.example](/webdav.service.example).

### CORS

The `allowed_*` properties are optional, the default value for each of them will be `*`. `exposed_headers` is optional as well, but is not set if not defined. Setting `credentials` to `true` will allow you to:

1. Use `withCredentials = true` in javascript.
2. Use the `username:password@host` syntax.

## License

MIT © 
