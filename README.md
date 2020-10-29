**This is still a work in progress**

New develop take place in branch `dev`.

If you want to develop `go-ilno`, fell free to [contact](https://wrong.wang/about/) me.

# go-ilno

![go-ilno_s.png](https://i.loli.net/2018/10/16/5bc556ea1ae9a.png)

a commenting server similar to Disqus, while keeping completely API compatible with [ilno](https://posativ.org/ilno/)


## why another ilno

`ilno` is good, but it's hard to be installed or customized.
What' more, the frontend part of `ilno` use library that is no longer updated.

`go-ilno` is different from `ilno`:

* Written in Go (Golang)
* Works with Sqlite3 but easy to add other database support.
* Doesn't use any ORM
* Doesn't use any complicated framework
* Use only modern vanilla Javascript (ES6 and Fetch API)
* Single binary compiled statically without dependency

### Why choose Golang as a programming language?

Go is probably the best choice for self-hosted software:

* Go is a simple programming language.
* Running code concurrently is part of the language.
* It’s faster than a scripting language like PHP or Python.
* The final application is a binary compiled statically without any dependency.
* You just need to drop the executable on your server to deploy the application.
* You don’t have to worry about what version of PHP/Python is installed on your machine.
* Packaging the software using RPM/Debian/Docker is straightforward.

## Roadmap

1. rewrite ilno backend part
2. Pray that someone will help me rewrite the front part of ilno.

## Getting Started

### Prerequisites

go-ilno is commenting server written in Go language.

Make sure you have [go installed](https://golang.org/doc/install).

### Developing

Download the code: `git clone https://github.com/budui/go-ilno`

run `go build`

and play with it!

## Work in progress

**This is still a work in progress** so there's still bugs to iron out and as this
is my first project in Go the code could no doubt use an increase in quality,
but I'll be improving on it whenever I find the time. If you have any feedback
feel free to [raise an issue](https://github.com/budui/go-ilno/issues)/[submit a PR](https://github.com/budui/go-ilno/pulls).


## Contributing

I know NOTHING about javascript. I need someone to HELP ME!!!

If you want to develop `go-ilno`, fell free to [contact](https://wrong.wang/about/) me.

## Authors

* **Ray Wong** - *Initial work* - [wrong.wang](https://wrong.wang/about/)

## Thanks

[ilno](https://posativ.org/ilno/) & [ilno's contributors](https://github.com/posativ/ilno/graphs/contributors).

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details