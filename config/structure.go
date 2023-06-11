package config

type Config struct {
	ScansOutputDir string `json:"scansOutputDir"`
	RunPreviously  bool   `json:"runPreviously"`
}
