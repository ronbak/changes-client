package runner

import (
	"os"
	"sync"
)

const (
	STATUS_QUEUED      = "queued"
	STATUS_IN_PROGRESS = "in_progress"
	STATUS_FINISHED    = "finished"
)

func publishArtifacts(reporter *Reporter, cID string, workspace string, artifacts []string) {
	if len(artifacts) == 0 {
		return
	}

	matches, err := GlobTree(workspace, artifacts)
	if err != nil {
		panic("Invalid artifact pattern" + err.Error())
	}

	reporter.PushArtifacts(cID, matches)
}

func RunAllCmds(reporter *Reporter, config *Config, logsource *LogSource) string {
	result := "passed"

	wg := sync.WaitGroup{}

	for _, cmd := range config.Cmds {
		reporter.PushStatus(cmd.Id, STATUS_IN_PROGRESS, -1)
		wc, err := NewWrappedScriptCommand(cmd.Script, cmd.Id)
		if err != nil {
			reporter.PushStatus(cmd.Id, STATUS_FINISHED, 255)
			result = "failed"
			break
		}

		env := os.Environ()
		for k, v := range cmd.Env {
			env = append(env, k+"="+v)
		}
		wc.Cmd.Env = env

		if len(cmd.Cwd) > 0 {
			wc.Cmd.Dir = cmd.Cwd
		}

		// Aritifacts can do out-of-band but we want to send logs synchronously.
		sem := make(chan bool)
		go func() {
			logsource.reportChunks(wc.ChunkChan)
			sem <- true
		}()

		pState, err := wc.Run()
		if err != nil {
			reporter.PushStatus(cmd.Id, STATUS_FINISHED, 255)
			result = "failed"
			break
		} else {
			if pState.Success() {
				reporter.PushStatus(cmd.Id, STATUS_FINISHED, 0)
			} else {
				reporter.PushStatus(cmd.Id, STATUS_FINISHED, 1)
				result = "failed"
				break
			}
		}

		// Wait for all the logs to be sent to reporter
		<-sem

		wg.Add(1)
		go func(artifacts []string) {
			publishArtifacts(reporter, config.JobstepID, config.Workspace, artifacts)
			wg.Done()
		}(cmd.Artifacts)
	}

	wg.Wait()

	return result
}

func RunBuildPlan(reporter *Reporter, config *Config) {
	logsource := &LogSource{
		Name:      "console",
		JobstepID: config.JobstepID,
		Reporter:  reporter,
	}

	reporter.PushJobStatus(config.JobstepID, STATUS_IN_PROGRESS, "")

	result := RunAllCmds(reporter, config, logsource)

	reporter.PushJobStatus(config.JobstepID, STATUS_FINISHED, result)
}
