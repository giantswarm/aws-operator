package cloudflare

type Cloudflare struct {
	Domain string `json:"domain" yaml:"domain"`
	Token  string `json:"token" yaml:"token"`
}
