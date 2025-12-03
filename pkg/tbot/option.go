package tbot

type Option func(*Bot)

// WithToken - sets token
func WithToken(token string) Option {
	return func(bot *Bot) {
		bot.token = token
	}
}

// WithDebug - sets debug mode
func WithDebug(debug bool) Option {
	return func(bot *Bot){
		bot.debug = debug
	}
}

// WithReadTimeout - sets readTimeout
func WithReadTimeout(readTimeout int) Option{
	return func(bot *Bot){
		bot.readTimeout = readTimeout
	}
}