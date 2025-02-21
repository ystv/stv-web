package structs

// Config is a structure containing global website configuration.
//
// See the comments for Server and PageContext for more details.
type (
	Config struct {
		Server Server `toml:"server"`
		AD     AD     `toml:"ad"`
		Mail   Mail   `toml:"mail"`
	}

	Server struct {
		Debug                 bool   `toml:"debug"`
		Address               string `toml:"address"`
		DomainName            string `toml:"domain_name"`
		EncryptionKey         string `toml:"encryption_key"`
		ForceResetURLEndpoint string `toml:"force_reset_url_endpoint"`
		Commit                string
		Version               string
	}

	AD struct {
		BypassUsername string `toml:"ad_bypass_username"`
		BypassPassword string `toml:"ad_bypass_password"`
		Server         string `toml:"ad_server"`
		Port           int    `toml:"ad_port"`
		BaseDN         string `toml:"ad_base_dn"`
		Security       int    `toml:"ad_security"`
		Bind           ADBind `toml:"bind"`
	}

	ADBind struct {
		Username string `toml:"ad_bind_username"`
		Password string `toml:"ad_bind_password"`
	}

	Mail struct {
		Host      string `toml:"mail_host"`
		User      string `toml:"mail_username"`
		Password  string `toml:"mail_password"`
		Port      int    `toml:"mail_port"`
		DefaultTo string `toml:"mail_default_to"`
	}
)
