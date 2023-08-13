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
	Favicon                *template.HTML
	Edition                *string
	ModCount               int64
	Mods                   []string
	SecureProfilesEnforced *bool
	PreviewChatEnforced    *bool
}
