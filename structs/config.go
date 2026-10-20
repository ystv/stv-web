package structs

// Config is a structure containing global website configuration.
//
// See the comments for Server and PageContext for more details.
type (
	Config struct {
		Server   Server   `toml:"server"`
		Keycloak Keycloak `toml:"keycloak"`
		Mail     Mail     `toml:"mail"`
	}

	Server struct {
		Debug                 bool   `toml:"debug"`
		Port                  string `toml:"port"`
		Address               string `toml:"address"`
		DomainName            string `toml:"domain_name"`
		ForceResetURLEndpoint string `toml:"force_reset_url_endpoint"`
		Commit                string `toml:"commit,omitempty"`
		Version               string `toml:"version,omitempty"`
	}

	Keycloak struct {
		URL         string `toml:"keycloak_url"`
		Realm       string `toml:"keycloak_realm"`
		ClientID    string `toml:"keycloak_clientID"`
		Secret      string `toml:"keycloak_clientsecret"`
		RedirectURL string `toml:"keycloak_redirectURL"`
		Address     string `toml:"address,omitempty"`
	}

	Mail struct {
		Host      string `toml:"mail_host"`
		User      string `toml:"mail_username"`
		Password  string `toml:"mail_password"`
		Port      int    `toml:"mail_port"`
		DefaultTo string `toml:"mail_default_to"`
	}
)
