package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func fetchOMLXData() APIResponse {
	var data APIResponse
	resp, err := http.Get(config.APIURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		data = loadState().APIResponse
	} else {
		defer resp.Body.Close()
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			fmt.Fprintf(os.Stderr, "error decoding JSON: %v\n", err)
			data = loadState().APIResponse
		} else {
			state := OutputState{
				APIResponse: data,
				Machines:    processMachines(data),
			}
			saveState(state)
		}
	}
	return data
}

func saveState(data OutputState) {
	b, err := json.Marshal(data)
	if err != nil {
		return
	}
	os.WriteFile(stateFile, b, 0644)
}

func loadState() OutputState {
	var state OutputState
	b, err := os.ReadFile(stateFile)
	if err != nil {
		return state
	}
	json.Unmarshal(b, &state)
	return state
}
