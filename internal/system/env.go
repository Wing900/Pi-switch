package system

import "os"

type EnvCheckResult struct {
	Name    string `json:"name"`
	Found   bool   `json:"found"`
	Message string `json:"message"`
}

func CheckEnvVar(name string) EnvCheckResult {
	value, ok := os.LookupEnv(name)
	if ok && value != "" {
		return EnvCheckResult{
			Name:    name,
			Found:   true,
			Message: "环境变量已找到",
		}
	}
	return EnvCheckResult{
		Name:    name,
		Found:   false,
		Message: "环境变量未找到",
	}
}

func ResolveAPIKey(envName, literal string) (string, EnvCheckResult) {
	if envName != "" {
		result := CheckEnvVar(envName)
		if result.Found {
			value, _ := os.LookupEnv(envName)
			result.Message = "环境变量：" + envName
			return value, result
		}
		result.Message = "环境变量：" + envName + " 未找到"
		return "", result
	}
	return literal, EnvCheckResult{
		Name:    "",
		Found:   literal != "",
		Message: "本地模式：使用字面量密钥",
	}
}
