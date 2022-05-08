package guests

import (
	"os"
	"os/exec"
	"path"
)

func StartDenoWorker(workerDir string, appDir string) (cmd *exec.Cmd, err error) {
	cmd = exec.Command(
		"deno",
		"run", "--allow-all",
		path.Join(workerDir, "main.ts"),
		"--appDir", appDir,
		"--hostUrl", "http://localhost:8080",
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd, cmd.Start()
}
