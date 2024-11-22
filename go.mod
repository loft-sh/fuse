module github.com/nestybox/sysbox-fs/bazil

go 1.13

require (
	bazil.org/fuse v0.0.0-00010101000000-000000000000
	golang.org/x/net v0.23.0
	golang.org/x/sys v0.18.0
)

replace bazil.org/fuse => ./
