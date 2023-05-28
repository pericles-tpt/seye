package config

func GetRunPreviously() bool {
	return cfg.RunPreviously
}

func GetScansOutputDir() string {
	return cfg.ScansOutputDir
}

func SetRunPreviously(newVal bool) {
	cfg.RunPreviously = newVal
	cfg.Flush()
}

func SetScansOutputDir(newVal string) {
	cfg.ScansOutputDir = newVal
	cfg.Flush()
}
