package cmd

var version = "v0.6.4"

func SetVersion(v string) {
	if v != "" {
		version = v
		rootCmd.Version = v
	}
}
