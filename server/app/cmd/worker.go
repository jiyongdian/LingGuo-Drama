package cmd

import (
	"spiritFruit/app/jobs"
	"spiritFruit/pkg/appctx"
	myAsynq "spiritFruit/pkg/asynq"
	"spiritFruit/pkg/console"

	"github.com/spf13/cobra"
)

var CmdWorker = &cobra.Command{
	Use:   "worker",
	Short: "Start async task worker",
	Run:   runWorker,
	Args:  cobra.NoArgs,
}

func runWorker(cmd *cobra.Command, args []string) {
	console.Success("Worker starting...")
	myAsynq.StartServer(appctx.GetContext(), jobs.NewServeMux())
	console.Success("Worker exited properly")
}
