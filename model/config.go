package model

import (
	"os"
	"strconv"
	"strings"

	"github.com/naiba/nezha/pkg/utils"
	"github.com/spf13/viper"
	"sigs.k8s.io/yaml"
)

var Languages = map[string]string{
	"zh-CN": "简体中文",
	"zh-TW": "繁體中文",
	"en-US": "English",
	"es-ES": "Español",
}

var Themes = map[string]string{
	"default":       "Default",
	"daynight":      "JackieSung DayNight",
	"mdui":          "Neko Mdui",
	"hotaru":        "Hotaru",
	"angel-kanade":  "AngelKanade",
	"server-status": "ServerStatus",
	"custom":        "Custom(local)",
}

var DashboardThemes = map[string]string{
	"default": "Default",
	"custom":  "Custom(local)",
}

const (
	ConfigTypeGitHub     = "github"
	ConfigTypeGitee      = "gitee"
	ConfigTypeGitlab     = "gitlab"
	ConfigTypeJihulab    = "jihulab"
	ConfigTypeGitea      = "gitea"
	ConfigTypeCloudflare = "cloudflare"
	ConfigTypeOidc       = "oidc"
)

const (
	ConfigCoverAll = iota
	ConfigCoverIgnoreAll
)

type Config struct {
	Debug bool // debug模式开关

	Language       string // 系统语言，默认 zh-CN
	SiteName       string
	JWTSecretKey   string
	AgentSecretKey string
	ListenPort     uint
	InstallHost    string
	TLS            bool
	Location       string // 时区，默认为 Asia/Shanghai

	EnablePlainIPInNotification bool // 通知信息IP不打码

	// IP变更提醒
	EnableIPChangeNotification bool
	IPChangeNotificationTag    string
	Cover                      uint8  // 覆盖范围（0:提醒未被 IgnoredIPNotification 包含的所有服务器; 1:仅提醒被 IgnoredIPNotification 包含的服务器;）
	IgnoredIPNotification      string // 特定服务器IP（多个服务器用逗号分隔）

	IgnoredIPNotificationServerIDs map[uint64]bool // [ServerID] -> bool(值为true代表当前ServerID在特定服务器列表内）
	AvgPingCount                   int
	DNSServers                     string

	v *viper.Viper
}

// Read 读取配置文件并应用
func (c *Config) Read(path string) error {
	c.v = viper.New()
	c.v.SetConfigFile(path)
	err := c.v.ReadInConfig()
	if err != nil {
		return err
	}

	err = c.v.Unmarshal(c)
	if err != nil {
		return err
	}

	if c.ListenPort == 0 {
		c.ListenPort = 8008
	}
	if c.Language == "" {
		c.Language = "zh-CN"
	}
	if c.EnableIPChangeNotification && c.IPChangeNotificationTag == "" {
		c.IPChangeNotificationTag = "default"
	}
	if c.Location == "" {
		c.Location = "Asia/Shanghai"
	}
	if c.AvgPingCount == 0 {
		c.AvgPingCount = 2
	}
	if c.JWTSecretKey == "" {
		c.JWTSecretKey, err = utils.GenerateRandomString(1024)
		if err != nil {
			return err
		}
		if err = c.Save(); err != nil {
			return err
		}
	}

	if c.AgentSecretKey == "" {
		c.AgentSecretKey, err = utils.GenerateRandomString(32)
		if err != nil {
			return err
		}
		if err = c.Save(); err != nil {
			return err
		}
	}

	c.updateIgnoredIPNotificationID()
	return nil
}

// updateIgnoredIPNotificationID 更新用于判断服务器ID是否属于特定服务器的map
func (c *Config) updateIgnoredIPNotificationID() {
	c.IgnoredIPNotificationServerIDs = make(map[uint64]bool)
	splitedIDs := strings.Split(c.IgnoredIPNotification, ",")
	for i := 0; i < len(splitedIDs); i++ {
		id, _ := strconv.ParseUint(splitedIDs[i], 10, 64)
		if id > 0 {
			c.IgnoredIPNotificationServerIDs[id] = true
		}
	}
}

// Save 保存配置文件
func (c *Config) Save() error {
	c.updateIgnoredIPNotificationID()
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(c.v.ConfigFileUsed(), data, 0600)
}
