package config

type Config struct {
	MaxDepth   int
	IgnoreDirs []string
}

var Default = Config{
	MaxDepth: 4,
	IgnoreDirs: []string{
		".git",
	},
}
