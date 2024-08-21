package config

import "github.com/spf13/viper"

type Config struct {
	Port              string `mapstructure:"PORT"`
	DBUrl             string `mapstructure:"DB_URL"`
	JWTSecretKey      string `mapstructure:"JWT_SECRET_KEY"`
	PasskeyHost       string `mapstructure:"PASSKEY_HOST"`
	OriginalHost      string `mapstructure:"ORIGINAL_HOST"`
	AliveSessionHours string `mapstructure:"ALIVE_SESSION_HOURS"`
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
