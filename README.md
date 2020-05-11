# gocal
The Go Clean Architecture Linter checks whether your go imports are following the dependency rule.

Find out more about the Clean Architecture and its dependency rule in this [blog post](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html).

## Installation

Install simply via the Makefile:
```
make install
```
or with `go get`:
```
go get -u github.com/mheidinger/gocal
```

## Usage

For `gocal` to work, your application needs to be split up in layers that are represented by different packages.
These layer packages need to be listed from the inner to the outer layers in the `.gocal` file.
Gocal will check for a layer not having imports of any outer layers.
For example:
```
domain
application
adapter
plugin
```

With this configuration just run `gocal` in your project folder containing the `.gocal` and `go.mod` file.
Without any ouput your project contains no violations.
If there are any violations the output will look like the following:
```
~/Projects/gocal_test » gocal
domain/domain.go: Forbidden import of 'gocal_test/adapter'
domain/domain.go: Forbidden import of 'gocal_test/application'
```

If your gocal config file has a different name, use the name of the file as first argument:
```
~/Projects/gocal_test » gocal gocal.config
```