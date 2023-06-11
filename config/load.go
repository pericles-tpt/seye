package config

import (
	"encoding/json"
	"os"

	"github.com/joomcode/errorx"
)

var (
	cfg        *Config
	configPath = "./config.json"
)

func Load() error {
	f, err := os.Open(configPath)
	if os.IsNotExist(err) {
		f, err = os.Create(configPath)
		if err != nil {
			return errorx.Decorate(err, "unable to create file `%s`, after not detected", configPath)
		}
	} else if err != nil {
		return errorx.Decorate(err, "unable to open file `%s`", configPath)
	}

	st, err := f.Stat()
	if err != nil {
		return errorx.Decorate(err, "unable to stat file `%s`", configPath)
	}

	if st.Size() > 0 {
		jd := json.NewDecoder(f)
		return jd.Decode(&cfg)
	} else {
		cfg = &Config{}
	}

	return nil
}

func (s *Config) Flush() error {
	f, err := os.OpenFile(configPath, os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return errorx.Decorate(err, "unable to open file '%s' for 'flush'", configPath)
	}

	je := json.NewEncoder(f)
	return je.Encode(*cfg)
}
