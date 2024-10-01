package config

import (
	"log/slog"
	"os"
	"strconv"
)

func LoadEnv(c *Config) {
	slog.Info("Updating config using environment variables")
	opts := strOpts()
	opts = append(opts, intOpts()...)
	for _, opt := range opts {
		opt(c)
	}
}

func strOpts() []Option {
	var res []Option

	envToOpt := map[string]func(string) Option{
		"BHL_NAMES_DUMP_URL":     OptBHLDumpURL,
		"BHL_NAMES_URL":          OptBHLNamesURL,
		"BHL_NAMES_COL_DATA_URL": OptCoLDataURL,
		"BHL_NAMES_DB_DATABASE":  OptDbDatabase,
		"BHL_NAMES_DB_HOST":      OptDbHost,
		"BHL_NAMES_DB_USER":      OptDbUser,
		"BHL_NAMES_DB_PASS":      OptDbPass,
		"BHL_NAMES_ROOT_DIR":     OptRootDir,
	}

	for envVar, optFunc := range envToOpt {
		envVal := os.Getenv(envVar)
		if envVal != "" {
			res = append(res, optFunc(envVal))
		}
	}

	return res
}

func intOpts() []Option {
	var res []Option
	envToOpt := map[string]func(int) Option{
		"BHL_NAMES_JOBS_NUM":  OptJobsNum,
		"BHL_NAMES_PORT_REST": OptPortREST,
	}
	for envVar, optFunc := range envToOpt {
		if envVar == "" {
			continue
		}
		val := os.Getenv(envVar)
		i, err := strconv.Atoi(val)
		if err != nil {
			slog.Warn("Cannot convert to int", "env", envVar, "value", val)
			continue
		}
		res = append(res, optFunc(i))
	}
	return res
}
