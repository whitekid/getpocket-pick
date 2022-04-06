package config

import (
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/whitekid/goxp/cryptox"
	"github.com/whitekid/goxp/flags"
)

const (
	keyBind          = "bind_addr"
	keyRootURL       = "root_url"
	keySecretKey     = "secret"
	keyConsumerKey   = "consumer_key"
	keyAccessToken   = "access_token"
	keyCookieTimeout = "cookie_timeout"
	keyCacheTimeout  = "favorite_cache_timeout"
)

var configs = map[string][]flags.Flag{
	"pocket-pick": {
		{keyBind, "B", "127.0.0.1:8000", "bind address"},
		{keyRootURL, "r", "http://127.0.0.0:8000", "root url"},
		{keySecretKey, "", "", "encrypt secret key"},
		{keyConsumerKey, "k", "", "getpocket consumer key"},
		{keyAccessToken, "a", "", "getpocket access token"},
		{keyCookieTimeout, "c", time.Hour * 24 * 30 * 12, "cookie timeout"},
		{keyCacheTimeout, "", time.Hour, "timeout for cache favorite items"},
	},
}

func init() {
	viper.SetEnvPrefix("pp")
	viper.AutomaticEnv()

	flags.InitDefaults(nil, configs)
}

func InitFlagSet(use string, fs *pflag.FlagSet) { flags.InitFlagSet(nil, configs, use, fs) }

// Config access functions
func BindAddr() string                    { return viper.GetString(keyBind) }
func RootURL() string                     { return viper.GetString(keyRootURL) }
func SecretKey() string                   { return viper.GetString(keySecretKey) }
func ConsumerKey() string                 { return cryptox.MustDecrypt(SecretKey(), viper.GetString(keyConsumerKey)) }
func AccessToken() string                 { return cryptox.MustDecrypt(SecretKey(), viper.GetString(keyAccessToken)) }
func CacheEvictionTimeout() time.Duration { return viper.GetDuration(keyCacheTimeout) }
func CookieTimeout() time.Duration        { return viper.GetDuration(keyCookieTimeout) }
