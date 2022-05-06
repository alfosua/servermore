package guests

import (
	"os"
	"os/exec"
	"path"
)

func StartDenoWorker(workerDir string, appDir string) (err error) {
	denoCmd := exec.Command(
		"deno",
		"run", "--allow-all",
		path.Join(workerDir, "main.ts"),
		"--appDir", appDir,
	)

	denoCmd.Stdout = os.Stdout
	denoCmd.Stderr = os.Stderr

	return denoCmd.Start()
}
