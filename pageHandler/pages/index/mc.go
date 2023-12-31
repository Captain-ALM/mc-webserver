package index

import (
	"html/template"
	"time"
)

type MC struct {
	Timestamp              time.Time
	Version                *string
	ProtocolVersion        *int64
	Address                string
	Port                   uint16
	Port6                  *uint16
	PlayerCount            *int64
	MaxPlayers             *int64
	Players                []string
	MOTD                   string
	ActualHost             *string
	ActualPort             *uint16
	Favicon                *string
	Edition                *string
	ModCount               int64
	Mods                   []string
	SecureProfilesEnforced *bool
	PreviewChatEnforced    *bool
	PreventsChatReports    *bool
}

func (m MC) GetFaviconSRC() template.HTMLAttr {
	toReturn := "src=\""
	if m.Favicon != nil {
		toReturn += *m.Favicon
	}
	toReturn += "\""
	return template.HTMLAttr(toReturn)
}

func (m MC) GetTimestamp() string {
	return m.Timestamp.Format(time.RFC822)
}

func (m MC) GetPlayers(aap bool) []string {
	toReturn := make([]string, 0)
	for _, cpl := range m.Players {
		if aap || cpl != "Anonymous Player" {
			toReturn = append(toReturn, cpl)
		}
	}
	return toReturn
}
