package defaults

import (
	"github.com/sahib/config"
)

// DefaultsV0 is the default config validation for brig
var DefaultsV0 = config.DefaultMapping{
	"server": config.DefaultMapping{
		"port": config.DefaultEntry{
			Default:      5000,
			NeedsRestart: true,
			Docs:         "Port of the http server",
			Validator:    config.IntRangeValidator(1, 655356),
		},
		"certfile": config.DefaultEntry{
			Default:      "",
			NeedsRestart: false,
			Docs:         "Path to an existing certificate file",
		},
		"keyfile": config.DefaultEntry{
			Default:      "",
			NeedsRestart: false,
			Docs:         "Path to an existing key file",
		},
		"domain": config.DefaultEntry{
			Default:      "localhost",
			NeedsRestart: false,
			Docs:         "",
		},
	},
	"database": config.DefaultMapping{
		"sqlite_path": config.DefaultEntry{
			Default:      "./data.db",
			NeedsRestart: true,
			Docs:         "Where the sqlite3 database is stored",
		},
		"session_cache": config.DefaultEntry{
			Default:      "./session.cache",
			NeedsRestart: false,
			Docs:         "Where the session cache is stored",
		},
	},
	"auth": config.DefaultMapping{
		"expire_time": config.DefaultEntry{
			Default:      "48h",
			NeedsRestart: false,
			Docs:         "",
		},
	},
	"mail": config.DefaultMapping{
		"from": config.DefaultEntry{
			Default:      "christopher@ira-kunststoffe.de",
			NeedsRestart: false,
			Docs:         "",
		},
		"smtp_host": config.DefaultEntry{
			Default:      "smtp.1und1.de",
			NeedsRestart: false,
			Docs:         "",
		},
		"smtp_port": config.DefaultEntry{
			Default:      465,
			NeedsRestart: false,
			Docs:         "",
		},
		"smtp_password": config.DefaultEntry{
			Default:      "",
			NeedsRestart: false,
			Docs:         "",
		},
	},
}
