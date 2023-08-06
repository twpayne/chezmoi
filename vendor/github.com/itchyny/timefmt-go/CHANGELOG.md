# Changelog
## [v0.1.5](https://github.com/itchyny/timefmt-go/compare/v0.1.4..v0.1.5) (2022-12-01)
* support parsing time zone offset with name using both `%z` and `%Z`

## [v0.1.4](https://github.com/itchyny/timefmt-go/compare/v0.1.3..v0.1.4) (2022-09-01)
* improve documents
* drop support for Go 1.16

## [v0.1.3](https://github.com/itchyny/timefmt-go/compare/v0.1.2..v0.1.3) (2021-04-14)
* implement `ParseInLocation` for configuring the default location

## [v0.1.2](https://github.com/itchyny/timefmt-go/compare/v0.1.1..v0.1.2) (2021-02-22)
* implement parsing/formatting time zone offset with colons (`%:z`, `%::z`, `%:::z`)
* recognize `Z` as UTC on parsing time zone offset (`%z`)
* fix padding on formatting time zone offset (`%z`)

## [v0.1.1](https://github.com/itchyny/timefmt-go/compare/v0.1.0..v0.1.1) (2020-09-01)
* fix overflow check in 32-bit architecture

## [v0.1.0](https://github.com/itchyny/timefmt-go/compare/2c02364..v0.1.0) (2020-08-16)
* initial implementation
