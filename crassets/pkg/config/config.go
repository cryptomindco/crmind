package config

import "github.com/spf13/viper"

type Config struct {
	Port                       string  `mapstructure:"PORT"`
	DBUrl                      string  `mapstructure:"DB_URL"`
	AllowAssets                string  `mapstructure:"ALLOW_ASSETS"`
	BtcMainetPort              string  `mapstructure:"BTC_MAINET_PORT"`
	BtcTestnetPort             string  `mapstructure:"BTC_TESTNET_PORT"`
	BtcRpcHost                 string  `mapstructure:"BTC_RPC_HOST"`
	BtcRpcUser                 string  `mapstructure:"BTC_RPC_USER"`
	BtcRpcPass                 string  `mapstructure:"BTC_RPC_PASS"`
	BtcWalletPassphrase        string  `mapstructure:"BTC_WALLET_PASSPHRASE"`
	BtcSystemWallet            string  `mapstructure:"BTC_SYSTEM_WALLET"`
	BtcTestnetWalletPassphrase string  `mapstructure:"BTC_TESTNET_WALLET_PASSPHRASE"`
	BtcSystemTestnetWallet     string  `mapstructure:"BTC_SYSTEM_TESTNET_WALLET"`
	DcrMainnetPort             string  `mapstructure:"DCR_MAINET_PORT"`
	DcrTestnetPort             string  `mapstructure:"DCR_TESTNET_PORT"`
	DcrWalletMainnetPort       string  `mapstructure:"DCR_WALLET_MAINNET_PORT"`
	DcrWalletTestnetPort       string  `mapstructure:"DCR_WALLET_TESTNET_PORT"`
	DcrRpcHost                 string  `mapstructure:"DCR_RPC_HOST"`
	DcrRpcUser                 string  `mapstructure:"DCR_RPC_USER"`
	DcrRpcPass                 string  `mapstructure:"DCR_RPC_PASS"`
	LtcMainetPort              string  `mapstructure:"LTC_MAINET_PORT"`
	LtcTestnetPort             string  `mapstructure:"LTC_TESTNET_PORT"`
	LtcRpcHost                 string  `mapstructure:"LTC_RPC_HOST"`
	LtcRpcUser                 string  `mapstructure:"LTC_RPC_USER"`
	LtcRpcPass                 string  `mapstructure:"LTC_RPC_PASS"`
	LtcWalletPassphrase        string  `mapstructure:"LTC_WALLET_PASSPHRASE"`
	LtcSystemWallet            string  `mapstructure:"LTC_SYSTEM_WALLET"`
	LtcTestnetWalletPassphrase string  `mapstructure:"LTC_TESTNET_WALLET_PASSPHRASE"`
	LtcSystemTestnetWallet     string  `mapstructure:"LTC_SYSTEM_TESTNET_WALLET"`
	SystemSyncTime             string  `mapstructure:"SYSTEM_SYNC_TIME"`
	Exchange                   string  `mapstructure:"EXCHANGE"`
	PriceSpread                float64 `mapstructure:"PRICE_SPREAD"`
}

func LoadConfig() (Config, error) {
	var config Config
	viper.AddConfigPath("../pkg/config/envs")
	viper.SetConfigName("dev")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return Config{}, err
	}
	err := viper.Unmarshal(&config)
	return config, err
}
