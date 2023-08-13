package index

import (
	"html/template"
	"time"
)

type DataYaml struct {
	PageTitle                     string        `yaml:"pageTitle"`
	ServerDescription             template.HTML `yaml:"serverDescription"`
	MCAddress                     string        `yaml:"mcAddress"`
	MCPort                        uint16        `yaml:"mcPort"`
	MCType                        string        `yaml:"mcType"`
	MCProtocolVersion             int           `yaml:"mcProtocolVersion"`
	MCClientGUID                  int64         `yaml:"mcClientGUID"`
	MCTimeout                     time.Duration `yaml:"mcTimeout"`
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
}