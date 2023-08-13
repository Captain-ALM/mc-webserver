package index

import (
	"errors"
	"github.com/mcstatus-io/mcutil"
	"github.com/mcstatus-io/mcutil/options"
	"github.com/mcstatus-io/mcutil/response"
	"html/template"
	"strings"
)

type Marshal struct {
	Data          DataYaml
	Queried       MC
	PlayersShown  bool
	ModsShown     bool
	ExtendedShown bool
	Parameters    template.URL
	Light         bool
	Online        bool
}

func (m Marshal) NewMC() (MC, error) {
	switch strings.ToLower(m.Data.MCType) {
	case "java":
		r, err := mcutil.Status(m.Data.MCAddress, m.Data.MCPort, options.JavaStatus{
			EnableSRV:       m.ExtendedShown && m.Data.AllowDisplayActualAddress,
			Timeout:         m.Data.MCTimeout,
			ProtocolVersion: m.Data.MCProtocolVersion,
		})
		if err != nil {
			return MC{}, err
		}
		r2, err := mcutil.StatusRaw(m.Data.MCAddress, m.Data.MCPort, options.JavaStatus{
			Timeout:         m.Data.MCTimeout,
			ProtocolVersion: m.Data.MCProtocolVersion,
		})
		if err != nil {
			return MC{}, err
		}
		return MC{
			Version:                &r.Version.NameClean,
			ProtocolVersion:        &r.Version.Protocol,
			Address:                m.Data.MCAddress,
			Port:                   m.Data.MCPort,
			Port6:                  nil,
			PlayerCount:            r.Players.Online,
			MaxPlayers:             r.Players.Max,
			Players:                CollectPlayers(r.Players.Sample),
			MOTD:                   r.MOTD.Clean,
			ActualHost:             CollectSRVHost(r.SRVResult),
			ActualPort:             CollectSRVPort(r.SRVResult),
			Favicon:                CollectFavicon(r.Favicon),
			Edition:                CollectModEdition(r.ModInfo),
			ModCount:               CollectModCount(r.ModInfo),
			Mods:                   CollectMods(r.ModInfo),
			SecureProfilesEnforced: CollectSecureProfileEnforcement(r2),
			PreviewChatEnforced:    CollectPreviewChatEnforcement(r2),
		}, nil
	case "legacy", "legacyjava", "javalegacy", "legacy java", "java legacy", "legacy_java", "java_legacy":
		r, err := mcutil.StatusLegacy(m.Data.MCAddress, m.Data.MCPort, options.JavaStatusLegacy{
			EnableSRV:       m.ExtendedShown && m.Data.AllowDisplayActualAddress,
			Timeout:         m.Data.MCTimeout,
			ProtocolVersion: m.Data.MCProtocolVersion,
		})
		if err != nil {
			return MC{}, err
		}
		return MC{
			Version:                &r.Version.NameClean,
			ProtocolVersion:        &r.Version.Protocol,
			Address:                m.Data.MCAddress,
			Port:                   m.Data.MCPort,
			Port6:                  nil,
			PlayerCount:            &r.Players.Online,
			MaxPlayers:             &r.Players.Max,
			Players:                nil,
			MOTD:                   r.MOTD.Clean,
			ActualHost:             CollectSRVHost(r.SRVResult),
			ActualPort:             CollectSRVPort(r.SRVResult),
			Favicon:                nil,
			Edition:                nil,
			ModCount:               0,
			Mods:                   nil,
			SecureProfilesEnforced: nil,
			PreviewChatEnforced:    nil,
		}, nil
	case "bedrock":
		r, err := mcutil.StatusBedrock(m.Data.MCAddress, m.Data.MCPort, options.BedrockStatus{
			EnableSRV:  m.ExtendedShown && m.Data.AllowDisplayActualAddress,
			Timeout:    m.Data.MCTimeout,
			ClientGUID: m.Data.MCClientGUID,
		})
		if err != nil {
			return MC{}, err
		}
		return MC{
			Version:                r.Version,
			ProtocolVersion:        r.ProtocolVersion,
			Address:                m.Data.MCAddress,
			Port:                   m.CollectIPv4Port(r.PortIPv4),
			Port6:                  r.PortIPv6,
			PlayerCount:            r.OnlinePlayers,
			MaxPlayers:             r.MaxPlayers,
			Players:                nil,
			MOTD:                   r.MOTD.Clean,
			ActualHost:             CollectSRVHost(r.SRVResult),
			ActualPort:             CollectSRVPort(r.SRVResult),
			Favicon:                nil,
			Edition:                r.Edition,
			ModCount:               0,
			Mods:                   nil,
			SecureProfilesEnforced: nil,
			PreviewChatEnforced:    nil,
		}, nil
	default:
		return MC{}, errors.New("Invalid MCType")
	}
}

func CollectPlayers(sampleArray []response.SamplePlayer) []string {
	if sampleArray == nil {
		return nil
	}
	toReturn := make([]string, len(sampleArray))
	for i := 0; i < len(sampleArray); i++ {
		toReturn[i] = sampleArray[i].NameClean
	}
	return toReturn
}

func CollectSRVHost(srv *response.SRVRecord) *string {
	if srv == nil {
		return nil
	}
	return &srv.Host
}

func CollectSRVPort(srv *response.SRVRecord) *uint16 {
	if srv == nil {
		return nil
	}
	return &srv.Port
}

func CollectFavicon(favicon *string) *template.HTML {
	if favicon == nil {
		return nil
	}
	toReturn := template.HTML(*favicon)
	return &toReturn
}

func CollectModEdition(mod *response.ModInfo) *string {
	if mod == nil {
		return nil
	}
	return &mod.Type
}

func CollectModCount(mod *response.ModInfo) int64 {
	if mod == nil {
		return 0
	}
	return int64(len(mod.Mods))
}

func CollectMods(mod *response.ModInfo) []string {
	if mod == nil {
		return nil
	}
	toReturn := make([]string, len(mod.Mods))
	for i := 0; i < len(mod.Mods); i++ {
		toReturn[i] = mod.Mods[i].ID + " (" + mod.Mods[i].Version + ")"
	}
	return toReturn
}

func CollectSecureProfileEnforcement(data map[string]interface{}) *bool {
	val, ok := data["enforcesSecureChat"]
	if ok {
		toReturn := val.(bool)
		return &toReturn
	}
	return nil
}

func CollectPreviewChatEnforcement(data map[string]interface{}) *bool {
	val, ok := data["previewsChat"]
	if ok {
		toReturn := val.(bool)
		return &toReturn
	}
	return nil
}

func (m Marshal) CollectIPv4Port(port *uint16) uint16 {
	if port == nil {
		return m.Data.MCPort
	}
	return *port
}
