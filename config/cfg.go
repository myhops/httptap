package config

type TapHandler struct {
	ListenAddress string               `yaml:"listenAddress"`
	Logging       *Logger              `yaml:"logging"`
	Upstream      string               `yaml:"upstream"`
	Header        HeaderIncludeExclude `yaml:"header"`

	Taps []*Tap `yaml:"taps"`
}

type HeaderIncludeExclude struct {
	Exclude []string `yaml:"exclude"`
	Include []string `yaml:"include"`
}

type Logger struct {
	LogLevel string `yaml:"logLevel"`
	LogFile  string `yaml:"logFile"`
}

type Tap struct {
	Name     string               `yaml:"name"`
	Patterns []string             `yaml:"patterns"`
	Header   HeaderIncludeExclude `yaml:"header"`

	// Add more taps when they become available.
	LogTap      *LogTap      `yaml:"logTap"`
	TemplateTap *TemplateTap `yaml:"templateTap"`

	Body       bool   `yaml:"body"`
	BodyFilter string `yaml:"bodyFilter"`
}

type LogTap struct {
	LogFile string `yaml:"logFile"`
}

type TemplateTap struct {
	Template string `yaml:"template"`
	LogFile  string `yaml:"logFile"`
}
