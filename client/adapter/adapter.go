package adapter

import (
	"fmt"

	"github.com/dropbox/changes-client/client"
)

type Adapter interface {
	// Init should be called before any other methods, and no more than once.
	Init(*client.Config) error
	// Prepare must be called no more than once, and must return successfully
	// before any method other than Init is called. Any returned metrics
	// will be reported via the active Reporter
	Prepare(*client.Log) (client.Metrics, error)
	Run(*client.Command, *client.Log) (*client.CommandResult, error)
	Shutdown(*client.Log) (client.Metrics, error)
	CaptureSnapshot(string, *client.Log) error
	GetRootFs() string
	CollectArtifacts([]string, *client.Log) ([]string, error)
	// Get absolute path to directory in which artifacts are searched
	GetArtifactRoot() string
}

func FormatUUID(uuid string) string {
	return fmt.Sprintf("%s-%s-%s-%s-%s", uuid[0:8], uuid[8:12], uuid[12:16], uuid[16:20], uuid[20:])
}
