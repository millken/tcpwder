package config

/**
 * Config file top-level object
 */
type Config struct {
	Logging  LoggingConfig     `toml:"logging" json:"logging"`
	Api      ApiConfig         `toml:"api" json:"api"`
	Defaults ConnectionOptions `toml:"defaults" json:"defaults"`
	Servers  map[string]Server `toml:"servers" json:"servers"`
}

/**
 * Logging config section
 */
type LoggingConfig struct {
	Level  string `toml:"level" json:"level"`
	Output string `toml:"output" json:"output"`
}

/**
 * Api config section
 */
type ApiConfig struct {
	Enabled   bool                `toml:"enabled" json:"enabled"`
	Bind      string              `toml:"bind" json:"bind"`
	BasicAuth *ApiBasicAuthConfig `toml:"basic_auth" json:"basic_auth"`
	Tls       *ApiTlsConfig       `toml:"tls" json:"tls"`
	Cors      bool                `toml:"cors" json:"cors"`
}

/**
 * Api Basic Auth Config
 */
type ApiBasicAuthConfig struct {
	Login    string `toml:"login" json:"login"`
	Password string `toml:"password" json:"password"`
}

/**
 * Api TLS server Config
 */
type ApiTlsConfig struct {
	CertPath string `toml:"cert_path" json:"cert_path"`
	KeyPath  string `toml:"key_path" json:"key_path"`
}

/**
 * Default values can be overridden in server
 */
type ConnectionOptions struct {
	MaxConnections           *int    `toml:"max_connections" json:"max_connections"`
	ClientIdleTimeout        *string `toml:"client_idle_timeout" json:"client_idle_timeout"`
	BackendIdleTimeout       *string `toml:"backend_idle_timeout" json:"backend_idle_timeout"`
	BackendConnectionTimeout *string `toml:"backend_connection_timeout" json:"backend_connection_timeout"`
	ChinaIpdbPath            string  `toml:"china_ipdb_path" json:"china_ipdb_path"`
}

type Upstream []string

/**
 * Server section config
 */
type Server struct {
	ConnectionOptions

	// hostname:port
	Bind string `toml:"bind" json:"bind"`

	// tcp | udp | tls
	Protocol string `toml:"protocol" json:"protocol"`

	// weight | leastconn | roundrobin
	Balance string `toml:"balance" json:"balance"`

	//upstream
	Upstream []string `toml:"upstream" json:"upstream"`

	// Optional configuration for server name indication
	Sni *Sni `toml:"sni" json:"sni"`

	// Optional configuration for protocol = tls
	Tls *Tls `toml:"tls" json:"tls"`

	// Optional configuration for backend_tls_enabled = true
	BackendsTls *BackendsTls `toml:"backends_tls" json:"backends_tls"`

	// Optional configuration for protocol = udp
	Udp *Udp `toml:"udp" json:"udp"`

	// Filter limit_max_connection_filter configuration
	MaxConnections *int `toml:"max_connections" json:"max_connections"`

	// Filter limit_perip_connection_filter configuration
	PerIpConnections *uint `toml:"per_ip_connections" json:"per_ip_connections"`

	LimitReconnectRate          *LimitReconnectRate    `toml:"limit_reconnect_rate" json:"limit_reconnect_rate"`
	LimitPeripRate              *LimitPeripRate        `toml:"limit_per_ip_rate" json:"limit_per_ip_rate"`
	LimitChinaAccessDefault     string                 `toml:"limit_china_access_default" json:"limit_china_access_default"`
	LimitChinaAccess            []LimitChinaAccess     `toml:"limit_china_access" json:"limit_china_access"`
	FilterRequestContentDefault string                 `toml:"filter_request_content_default" json:"filter_request_content_default"`
	FilterRequestContent        []FilterRequestContent `toml:"filter_request_content" json:"filter_request_content"`

	// Healthcheck configuration
	Healthcheck *HealthcheckConfig `toml:"healthcheck" json:"healthcheck"`
}

/**
 * Server Sni options
 */
type Sni struct {
	HostnameMatchingStrategy   string `toml:"hostname_matching_strategy" json:"hostname_matching_strategy"`
	UnexpectedHostnameStrategy string `toml:"unexpected_hostname_strategy" json:"unexpected_hostname_strategy"`
	ReadTimeout                string `toml:"read_timeout" json:"read_timeout"`
}

/**
 * Common part of Tls and BackendTls types
 */
type tlsCommon struct {
	Ciphers             []string `toml:"ciphers" json:"ciphers"`
	PreferServerCiphers bool     `toml:"prefer_server_ciphers" json:"prefer_server_ciphers"`
	MinVersion          string   `toml:"min_version" json:"min_version"`
	MaxVersion          string   `toml:"max_version" json:"max_version"`
	SessionTickets      bool     `toml:"session_tickets" json:"session_tickets"`
}

/**
 * Server Tls options
 * for protocol = "tls"
 */
type Tls struct {
	CertPath string `toml:"cert_path" json:"cert_path"`
	KeyPath  string `toml:"key_path" json:"key_path"`
	tlsCommon
}

type BackendsTls struct {
	IgnoreVerify   bool    `toml:"ignore_verify" json:"ignore_verify"`
	RootCaCertPath *string `toml:"root_ca_cert_path" json:"root_ca_cert_path"`
	CertPath       *string `toml:"cert_path" json:"cert_path"`
	KeyPath        *string `toml:"key_path" json:"key_path"`
	tlsCommon
}

/**
 * Server udp options
 * for protocol = "udp"
 */
type Udp struct {
	MaxRequests  uint64 `toml:"max_requests" json:"max_requests"`
	MaxResponses uint64 `toml:"max_responses" json:"max_responses"`
}

/**
 * filter limit_reconnect_rate configuration
 */
type LimitReconnectRate struct {
	Interval   string `toml:"interval" json:"interval"`
	Reconnects int    `toml:"reconnects" json:"reconnects"`
}

/**
 * filter limit_per_ip_rate configuration
 */
type LimitPeripRate struct {
	Interval   string `toml:"interval" json:"interval"`
	ReadBytes  uint   `toml:"readbytes" json:"readbytes"`
	WriteBytes uint   `toml:"writebytes" json:"writebytes"`
}

/**
 * filter limit_china_access configuration
 */
type LimitChinaAccess struct {
	Area   string `toml:"area" json:"area"`
	Region string `toml:"region" json:"region"`
	Isp    string `toml:"isp" json:"isp"`
	Access string `toml:"access" json:"access"`
}

/**
 * filter filter_request_content configuration
 */
type FilterRequestContent struct {
	Mode    string `toml:"mode" json:"mode"`
	Content string `toml:"content" json:"content"`
	Access  string `toml:"access" json:"access"`
}

/**
 * Healthcheck configuration
 */
type HealthcheckConfig struct {
	Kind     string `toml:"kind" json:"kind"`
	Interval string `toml:"interval" json:"interval"`
	Passes   int    `toml:"passes" json:"passes"`
	Fails    int    `toml:"fails" json:"fails"`
	Timeout  string `toml:"timeout" json:"timeout"`

	/* Depends on Kind */

	*PingHealthcheckConfig
	*ExecHealthcheckConfig
}

type PingHealthcheckConfig struct{}

type ExecHealthcheckConfig struct {
	ExecCommand                string `toml:"exec_command" json:"exec_command,omitempty"`
	ExecExpectedPositiveOutput string `toml:"exec_expected_positive_output" json:"exec_expected_positive_output"`
	ExecExpectedNegativeOutput string `toml:"exec_expected_negative_output" json:"exec_expected_negative_output"`
}
