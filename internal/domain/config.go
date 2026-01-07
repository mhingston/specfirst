package domain

type Config struct {
	ProjectName string            `mapstructure:"project_name"`
	Protocol    string            `mapstructure:"protocol"`
	Language    string            `mapstructure:"language"`
	Framework   string            `mapstructure:"framework"`
	CustomVars  map[string]string `mapstructure:"custom_vars"`
	Constraints map[string]string `mapstructure:"constraints"`
}
