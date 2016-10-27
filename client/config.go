package client

type Config struct {
	Addrs []string `toml:"addrs"`
}

//equals return true if two configs are same.
func (c *Config) equals(config Config) bool {
	if len(c.Addrs) != len(config.Addrs) {
		return false
	}

	length := len(c.Addrs)
	for i := 0; i < length; i++ {
		if c.Addrs[i] != config.Addrs[i] {
			return false
		}
	}
	return true
}
