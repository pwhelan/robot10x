module github.com/pwhelan/robot10x

go 1.22

require (
	github.com/elastic/gosigar v0.14.2
	github.com/rubiojr/go-usbmon v1.0.0
)

require (
	github.com/jkeiser/iter v0.0.0-20200628201005-c8aa0ae784d1 // indirect
	github.com/jochenvg/go-udev v0.0.0-20240801134859-b65ed646224b // indirect
	golang.org/x/sys v0.0.0-20220412211240-33da011f77ad // indirect
)

replace github.com/rubiojr/go-usbmon => github.com/pwhelan/go-usbmon v0.0.0-20251109162029-905bf426e8e2
