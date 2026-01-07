package repository

import (
	"os"

	"specfirst/internal/domain"

	"github.com/spf13/viper"
)

// LoadConfig reads configuration from disk.
func LoadConfig(path string) (domain.Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		if os.IsNotExist(err) {
			return domain.Config{}, nil
		}
		return domain.Config{}, err
	}

	var cfg domain.Config
	if err := v.Unmarshal(&cfg); err != nil {
		return domain.Config{}, err
	}

	return cfg, nil
}
