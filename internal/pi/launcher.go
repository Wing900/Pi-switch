package pi

type LaunchPreview struct {
	Command   string   `json:"command"`
	Checklist []string `json:"checklist"`
}

func BuildCommand(piCommand, providerID, modelID string) string {
	return piCommand + " --provider " + providerID + " --model " + modelID
}

func OpenCommandInTerminal(command string, workingDir string) error {
	return openCommandInTerminal(command, workingDir)
}
