package pageHandler

import (
	"net/url"
	"time"
)

type PageProvider interface {
	GetPath() string
	GetLastModified() time.Time
	GetCacheIDExtension(urlParameters url.Values) string
	GetContents(urlParameters url.Values) (contentType string, contents []byte, canCache bool)
	PurgeTemplate()
}
