package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

const stateFile = "/tmp/omlxstate.json"
const configFile = "omlx-waybar.yaml"

var config = loadConfig()

func loadConfig() Config {
	cfg := Config{
		APIURL:   "http://omlx.ai.lucia.friedmicro-lab.org",
		Backends: []string{"small", "mid", "big"},
	}

	exe, err := os.Executable()
	if err != nil {
		return cfg
	}
	path := filepath.Join(filepath.Dir(exe), configFile)
	b, err := os.ReadFile(path)
	if err != nil {
		return cfg
	}

	var fileCfg Config
	if err := yaml.Unmarshal(b, &fileCfg); err != nil {
		fmt.Fprintf(os.Stderr, "error parsing %s: %v\n", configFile, err)
		return cfg
	}

	if fileCfg.APIURL != "" {
		cfg.APIURL = fileCfg.APIURL
	}
	if len(fileCfg.Backends) > 0 {
		cfg.Backends = fileCfg.Backends
	}
	return cfg
}

func main() {
	machines := fetchOMLXData()
	outText := ""
	var tooltip string
	for _, key := range config.Backends {
		machine, ok := machines[key]
		if !ok {
			outText += "⚪"
			if tooltip != "" {
				tooltip += "\n"
			}
			tooltip += fmt.Sprintf("%s: not available", key)
			continue
		}

		icon := processIconStatus(machine)
		tooltip += processTooltip(machine, key)
		outText += icon
	}

	out := Output{
		Text:    outText,
		Tooltip: tooltip,
		Class:   processCSS(machines),
	}

	b, _ := json.Marshal(out)
	fmt.Println(string(b))
}

func processIconStatus(machine Machine) string {
	if machine.Maintenance || machine.Busy {
		return "🔴"
	}
	return "🟢"
}

func processStatus(machine Machine) string {
	if machine.Maintenance {
		return "maintenance"
	} else if machine.Busy {
		return "busy"
	} else {
		return "ok"
	}
}

func processTooltip(machine Machine, key string) string {
	status := processStatus(machine)
	tooltip := ""
	if tooltip != "" {
		tooltip += "\n"
	}
	tooltip += fmt.Sprintf("%s: %s", key, status)
	tooltip += fmt.Sprintf("\n  model: %s", machine.Model)

	activeTask := getActiveTask(machine)
	if activeTask != nil {
		tooltip += processActiveTask(*activeTask)
	}

	tooltip += fmt.Sprintf("\n  memory: %s / %s", formatMemory(machine.ModelMemoryUsed), formatMemory(machine.ModelMemoryMax))
	return tooltip
}

func getActiveTask(machine Machine) *ActiveTask {
	if len(machine.Prefilling) > 0 {
		return &machine.Prefilling[0]
	} else if len(machine.Waiting) > 0 {
		return &machine.Waiting[0]
	} else if len(machine.Generating) > 0 {
		return &machine.Generating[0]
	}
	return nil
}

func processActiveTask(activeTask ActiveTask) string {
	eta := activeTask.ETA()
	tooltip := ""
	if eta > 0 {
		tooltip += fmt.Sprintf("\n  eta: %s", formatETA(eta))
	}
	tooltip += fmt.Sprintf("\n  phase: %s", activeTask.Phase())
	processed, total, pct := activeTask.Progress()
	tooltip += fmt.Sprintf("\n  progress: %d/%d (%.1f%%)", processed, total, pct)
	tooltip += fmt.Sprintf("\n  speed: %.1f tok/s", activeTask.Speed())
	return tooltip
}

func processCSS(machines APIResponse) string {
	for _, key := range config.Backends {
		machine, ok := machines[key]
		if !ok || machine.Maintenance {
			return "unhealthy"
		}
	}
	return "ok"
}
