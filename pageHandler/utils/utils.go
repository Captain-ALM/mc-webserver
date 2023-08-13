package utils

import (
	"net/http"
	"strconv"
	"time"
)

func SetNeverCacheHeader(header http.Header) {
	header.Set("Cache-Control", "max-age=0, no-cache, no-store, must-revalidate")
	header.Set("Pragma", "no-cache")
}

func SetLastModifiedHeader(header http.Header, modTime time.Time) {
	if !modTime.IsZero() {
		header.Set("Last-Modified", modTime.UTC().Format(http.TimeFormat))
	}
}

func SetCacheHeaderWithAge(header http.Header, maxAge uint, modifiedTime time.Time) {
	header.Set("Cache-Control", "max-age="+strconv.Itoa(int(maxAge))+", must-revalidate")
	if maxAge > 0 {
		checkerSecondsBetween := int64(time.Now().UTC().Sub(modifiedTime.UTC()).Seconds())
		if checkerSecondsBetween < 0 {
			checkerSecondsBetween *= -1
		}
		header.Set("Age", strconv.FormatUint(uint64(checkerSecondsBetween)%uint64(maxAge), 10))
	}
}

func SwitchToNonCachingHeaders(header http.Header) {
	SetNeverCacheHeader(header)
	if header.Get("Last-Modified") != "" {
		header.Del("Last-Modified")
	}
	if header.Get("Age") != "" {
		header.Del("Age")
	}
	if header.Get("Expires") != "" {
		header.Del("Expires")
	}
	if header.Get("ETag") != "" {
		header.Del("ETag")
	}
}
