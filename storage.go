package main

// Storage for redis interface
type Storage interface {
	Shorten(url string, exp int64) (string, error)
	ShortlinkInfo(eid string) (interface{}, error)
	UnShorten(eid string) (string, error)
}
