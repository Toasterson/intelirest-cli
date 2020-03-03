# InteliJ Rest CLI

This little client implements the InteliJ Rest File format and allows running them on the CLI.

## Installation

```shell script
go install github.com/toasterson/intelirest-cli
``` 

or alternate

```shell script
git clone https://github.com/toasterson/intelirest-cli
cd intelirest-cli
make all
```

## Usage
**NOTE:** this utility requires you to have a working InteliJ rest client files.

```shell script
rest-cli [-e ENVIRONMENT] FILE
```

```shell script
rest-cli -e development test.http
```

## Development
Currently no Javascript Client support is available. However it can be implemented easily using [Otto](https://github.com/robertkrimen/otto) https://github.com/robertkrimen/otto

No formal requirements yet.

## Sponsoring
This package is part of my Github Sponsor Profile. If you want to support development of this monetarily you can do so here https://github.com/sponsors/Toasterson