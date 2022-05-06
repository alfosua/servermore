package options

import (
	"flag"
	"fmt"
)

type HostOptions struct {
	WorkerDir string
	AppDir    string
	GuestEnv  string
}

func ParseArgs() (HostOptions, error) {
	var args HostOptions

	flag.StringVar(&args.WorkerDir, "workerdir", "", "worker directory for scripts runtime")
	flag.StringVar(&args.AppDir, "appdir", "", "app directory for scripts runtime")
	flag.StringVar(&args.GuestEnv, "guestenv", "deno", "guest environment")
	flag.Parse()

	if args.WorkerDir == "" || args.AppDir == "" {
		return args, fmt.Errorf("workerdir and appdir are required")
	}

	return args, nil
}
