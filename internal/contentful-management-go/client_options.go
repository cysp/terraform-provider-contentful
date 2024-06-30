package client

//nolint:ireturn
func WithUserAgent(userAgent string) ClientOption {
	return optionFunc[clientConfig](func(cfg *clientConfig) {
		cfg.Client = wrapClientWithUserAgent(cfg.Client, userAgent)
	})
}
