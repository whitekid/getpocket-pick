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

func InitConfig() {
	// 	if cfgFile != "" {
	// 		// Use config file from the flag.
	// 		viper.SetConfigFile(cfgFile)
	// 	} else {
	// 		// Find home directory.
	// 		home, err := os.UserHomeDir()
	// 		cobra.CheckErr(err)

	// 		// Search config in home directory with name ".cobra" (without extension).
	// 		viper.AddConfigPath(home)
	// 		viper.SetConfigType("yaml")
	// 		viper.SetConfigName(".cobra")
	// }

	viper.SetEnvPrefix("pp")
	viper.AutomaticEnv()

	//	if err := viper.ReadInConfig(); err == nil {
	//		fmt.Println("Using config file:", viper.ConfigFileUsed())
	//	}
}

func InitRootFlags(fs *pflag.FlagSet) {
	flags.String(fs, keyBind, "bind", "B", "127.0.0.1:8000", "bind address")
	flags.String(fs, keyRootURL, "root-url", "r", "http://127.0.0.0:8000", "root url")
	flags.String(fs, keySecretKey, "secret-key", "", "", "encrypt secret key")
	flags.String(fs, keyConsumerKey, "consumer-key", "k", "", "getpocket consumer key")
	flags.String(fs, keyAccessToken, "access-token", "a", "", "getpocket access token")
	flags.Duration(fs, keyCookieTimeout, "cookie-timeout", "c", time.Hour*24*30*12, "cookie timeout")
	flags.Duration(fs, keyCacheTimeout, "cache-timeout", "", time.Hour, "timeout for cache favorite items")
}

// Config access functions
func BindAddr() string                    { return viper.GetString(keyBind) }
func RootURL() string                     { return viper.GetString(keyRootURL) }
func SecretKey() string                   { return viper.GetString(keySecretKey) }
func ConsumerKey() string                 { return cryptox.MustDecrypt(SecretKey(), viper.GetString(keyConsumerKey)) }
func AccessToken() string                 { return cryptox.MustDecrypt(SecretKey(), viper.GetString(keyAccessToken)) }
func CacheEvictionTimeout() time.Duration { return viper.GetDuration(keyCacheTimeout) }
func CookieTimeout() time.Duration        { return viper.GetDuration(keyCookieTimeout) }
