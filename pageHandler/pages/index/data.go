package index

import (
	"html/template"
	"strings"
	"time"
)

type DataYaml struct {
	PageTitle                     string        `yaml:"pageTitle"`
	ServerDescription             template.HTML `yaml:"serverDescription"`
	Footer                        string        `yaml:"footer"`
	MCAddress                     string        `yaml:"mcAddress"`
	MCPort                        uint16        `yaml:"mcPort"`
	MCType                        string        `yaml:"mcType"`
	MCProtocolVersion             int           `yaml:"mcProtocolVersion"`
	MCClientGUID                  int64         `yaml:"mcClientGUID"`
	MCTimeout                     time.Duration `yaml:"mcTimeout"`
	MCQueryInterval               time.Duration `yaml:"mcQueryInterval"`
	AllowDisplayState             bool          `yaml:"allowDisplayState"`
	AllowDisplayVersion           bool          `yaml:"allowDisplayVersion"`
	AllowDisplayActualAddress     bool          `yaml:"allowDisplayActualAddress"`
	AllowPlayerCountDisplay       bool          `yaml:"allowPlayerCountDisplay"`
	AllowPlayerListing            bool          `yaml:"allowPlayerListing"`
	AllowMOTDDisplay              bool          `yaml:"allowMOTDDisplay"`
	AllowFaviconDisplay           bool          `yaml:"allowFaviconDisplay"`
	AllowSecureProfileModeDisplay bool          `yaml:"allowSecureProfileModeDisplay"`
	AllowPreviewChatModeDisplay   bool          `yaml:"allowPreviewChatModeDisplay"`
	AllowDisplayModded            bool          `yaml:"allowDisplayModded"`
	AllowModListing               bool          `yaml:"allowModListing"`
	ShowAnonymousPlayers          bool          `yaml:"showAnonymousPlayers"`
}

func (dy DataYaml) GetCleanMCType() string {
	return strings.Title(dy.MCType)
}
