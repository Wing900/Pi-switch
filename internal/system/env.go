package system

type EnvCheckResult struct {
	Name    string `json:"name"`
	Found   bool   `json:"found"`
	Message string `json:"message"`
}

func CheckEnvVar(name string) EnvCheckResult {
	return EnvCheckResult{
		Name:    name,
		Found:   false,
		Message: "环境变量模式：已停用",
	}
}

func ResolveAPIKey(envName, literal string) (string, EnvCheckResult) {
	if literal != "" {
		return literal, EnvCheckResult{
			Name:    envName,
			Found:   true,
			Message: "本地密钥：已设置",
		}
	}
	return literal, EnvCheckResult{
		Name:    envName,
		Found:   literal != "",
		Message: "本地密钥：未设置",
	}
}
