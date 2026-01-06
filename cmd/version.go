package cmd

var version = "v0.6.2"

func SetVersion(v string) {
	if v != "" {
		version = v
		rootCmd.Version = v
	}
}
