package pipewire

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

type PipeWireNode struct {
	ID   int                    `json:"id"`
	Type string                 `json:"type"`
	Info map[string]interface{} `json:"info"`
}

type PipeWireOutput struct {
	ID       int
	NodeName string
	Props    map[string]interface{}
}

// CheckCombinedOutput checks if a "combined_output" node exists in PipeWire
func CheckCombinedOutput() (*PipeWireOutput, error) {
	log.Printf("PipeWire: Checking for combined_output node")

	// Execute pw-dump command
	cmd := exec.Command("pw-dump")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute pw-dump: %w", err)
	}

	// Parse JSON output
	var nodes []PipeWireNode
	if err := json.Unmarshal(output, &nodes); err != nil {
		return nil, fmt.Errorf("failed to parse pw-dump output: %w", err)
	}

	// Search for combined_output node
	for _, node := range nodes {
		if node.Type != "PipeWire:Interface:Node" {
			continue
		}

		// Check if this node has info and props
		if node.Info == nil {
			continue
		}

		props, ok := node.Info["props"].(map[string]interface{})
		if !ok {
			continue
		}

		// Check for node.name property
		nodeName, ok := props["node.name"].(string)
		if !ok {
			continue
		}

		if nodeName == "combined_output" {
			log.Printf("PipeWire: Found combined_output node with ID %d", node.ID)
			return &PipeWireOutput{
				ID:       node.ID,
				NodeName: nodeName,
				Props:    props,
			}, nil
		}
	}

	return nil, errors.New("combined_output node not found")
}

// GetAllAudioNodes returns all audio nodes from PipeWire
func GetAllAudioNodes() ([]PipeWireOutput, error) {
	log.Printf("PipeWire: Getting all audio nodes")

	// Execute pw-dump command
	cmd := exec.Command("pw-dump")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute pw-dump: %w", err)
	}

	// Parse JSON output
	var nodes []PipeWireNode
	if err := json.Unmarshal(output, &nodes); err != nil {
		return nil, fmt.Errorf("failed to parse pw-dump output: %w", err)
	}

	var audioNodes []PipeWireOutput

	// Search for audio nodes
	for _, node := range nodes {
		if node.Type != "PipeWire:Interface:Node" {
			continue
		}

		// Check if this node has info and props
		if node.Info == nil {
			continue
		}

		props, ok := node.Info["props"].(map[string]interface{})
		if !ok {
			continue
		}

		// Check for node.name property
		nodeName, ok := props["node.name"].(string)
		if !ok {
			continue
		}

		// Check if it's an audio node
		mediaClass, ok := props["media.class"].(string)
		if ok && (strings.Contains(mediaClass, "Audio/Sink") ||
			strings.Contains(mediaClass, "Audio/Source") ||
			strings.Contains(nodeName, "audio") ||
			strings.Contains(nodeName, "output") ||
			strings.Contains(nodeName, "input")) {

			audioNodes = append(audioNodes, PipeWireOutput{
				ID:       node.ID,
				NodeName: nodeName,
				Props:    props,
			})
		}
	}

	log.Printf("PipeWire: Found %d audio nodes", len(audioNodes))
	return audioNodes, nil
}
