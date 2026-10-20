package main

import (
	"fmt"
	"math"
)

func processMachines(data APIResponse) []MachineState {
	var states []MachineState
	for _, key := range config.Backends {
		machine, ok := data[key]
		if !ok {
			states = processUnavailable(states)
			continue
		}
		etaStr := processPrefil(machine)
		if etaStr == "" {
			etaStr = processWaiting(machine)
		}
		if etaStr == "" {
			etaStr = processGenerating(machine)
		}

		states = append(states, MachineState{
			Status:     processStatus(machine),
			Icon:       processIconStatus(machine),
			Model:      machine.Model,
			ETA:        etaStr,
			MemoryUsed: formatMemory(machine.ModelMemoryUsed),
			MemoryMax:  formatMemory(machine.ModelMemoryMax),
		})
	}
	return states
}

func processUnavailable(states []MachineState) []MachineState {
	states = append(states, MachineState{Status: "unavailable", Icon: "⚪", Model: "-", ETA: "-"})
	return states
}

func processPrefil(machine Machine) string {
	for _, t := range machine.Prefilling {
		if eta := t.ETA(); eta > 0 {
			return formatETA(eta)
		}
	}
	return ""
}

func processWaiting(machine Machine) string {
	for _, t := range machine.Waiting {
		if eta := t.ETA(); eta > 0 {
			return formatETA(eta)
		}
	}
	return ""
}

func processGenerating(machine Machine) string {
	for _, t := range machine.Generating {
		if eta := t.ETA(); eta > 0 {
			return formatETA(eta)
		}
	}
	return ""
}

func formatETA(seconds float64) string {
	if seconds <= 0 {
		return "N/A"
	}
	hours := int(seconds / 3600)
	minutes := int(math.Mod(seconds, 3600)) / 60
	secs := int(math.Mod(seconds, 60))

	if hours > 0 {
		return fmt.Sprintf("%dh%dm", hours, minutes)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm%ds", minutes, secs)
	}
	return fmt.Sprintf("%ds", secs)
}

func formatMemory(bytes float64) string {
	gb := bytes / (1024 * 1024 * 1024)
	if gb >= 1 {
		return fmt.Sprintf("%.1f GB", gb)
	}
	mb := bytes / (1024 * 1024)
	return fmt.Sprintf("%.1f MB", mb)
}
