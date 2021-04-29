# Torrent-tool

This is a simple script that sorts and splits torrent files in go, It separates torrents into size based directories

## Installation

Using this package requires a working Go environment. [See the install instructions for Go](http://golang.org/doc/install.html).

Go Modules are required when using this package. [See the go blog guide on using Go Modules](https://blog.golang.org/using-go-modules).

## Flags

To run the script you need to provide 4 flags or the help flag

### [-h,--help]

Shows help and all flags

### [-d,--directory]

Provide the full path to the input directory where the torrents are located

### [-o,--output]

Provide the full path to the output directory where the subdirectories and torrents will be copied

### [-s,--size]

The size limit the torrents will be split into in NGB or NTB format

### [-t,--type]

Sorting type of the torrents, 1 to sort by size and 2 to sort by name

## Usage

Run the script using go

```
$ go run script.go -d<value> -o<value> -s<value> -t <value>
or
$ go run script.go -directory<value> -output<value> -size<value> -type <value>
```
