# ilno

`isso` with lightning

## What it does

It includes a JavaScript-powered, Disqus-like comment box in your HTML pages. The code is based on on [Isso](https://posativ.org/isso/) and [go-isso](https://github.com/budui/go-isso).

## How is it different?

1. It doesn't store cookies or does any kind of antispam, instead it requires [lnurl-auth](https://github.com/btcontract/lnurl-rfc/blob/master/lnurl-auth.md) logins for commenters. [Read more](https://github.com/fiatjaf/awesome-lnurl).
2. In the future it will allow requiring Lightning payments for commenting as antispam measure and a way to fund the host.

## Installation

Depends on [sqlite3](https://www.sqlite.org/index.html) being installed (better install with your default `apt`/`rpm`/`pacman`/whatever dependency management system).

Then install with [Go](https://golang.org/dl/):

```
go get github.com/fiatjaf/ilno
```

Or download a binary directly from the [releases page](https://github.com/fiatjaf/ilno/releases/).

### Running the server

And run with

```
./ilno
```

It will start listening at `0.0.0.0:11140` and create an sqlite database on `./comments.db`. And by default it will accept requests from any domain and will have no admin. To change these things set the following environment variables:

```
HOST=<ip address>
PORT=<port number>
DATABASE=<file path to the db>
ALLOWED_ORIGINS=<comma-separated list of domains, generally just a single domain>
ADMIN_KEY=<the lnurl-auth key for the admin>
```

To get your lnurl-auth key first run the server without admin, login and **double-click the abbreviated key** to get the full key.

You can set only the values for which you want to change the defaults.

Setting environment variables can be done in multiple ways, for example, calling `ilno` with `PORT=... ADMIN_KEY=... ./ilno`.

### Embedding in the HTML

In your HTML pages you must include the following snippet:

```
<script data-ilno="https://domain.that.is.serving.your.ilno"
        src="https://domain.that.is.serving.your.ilno/js/embed.js"></script>

<section id="ilno-thread"></section>
```

## Server setup and proxy tips (ignore if you know what you're doing)(if you have no idea of what you're doing these tips will probably not be enough anyway)

If you have

1. A Linux server with a public IP
2. A domain name

You can

1. run `ilno` from your public server, then
2. use [Caddy](https://caddyserver.com/) or something else to serve `ilno` in a subdomain of the domain name you use (like `comments.mydomain.com`) with a simple [`reverse_proxy`](https://caddyserver.com/docs/quick-starts/reverse-proxy) rule (Caddy will deal with setting up automatic HTTPS for you if it is running as a system process, as root), and finally
3. write an `A` rule from your DNS provider (generally the place where you bought the domain) to the IP of your Linux server

## Sponsored by

- [Bitcoin Audible](https://bitcoinaudible.com/)
