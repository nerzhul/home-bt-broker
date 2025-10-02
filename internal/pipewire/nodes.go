package pipewire

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/nerzhul/home-bt-broker/internal/database"
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

// UpdateCombinedSinks configures the combined_output node with new sink targets
func UpdateCombinedSinks(deviceMACs []string) error {
	log.Printf("PipeWire: Updating combined sinks with %d devices", len(deviceMACs))

	// First, find the combined_output node
	combinedOutput, err := CheckCombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to find combined_output node: %w", err)
	}

	// Convert MAC addresses to PipeWire node names
	targetSinks, err := getBluetoothSinkNames(deviceMACs)
	if err != nil {
		return fmt.Errorf("failed to get bluetooth sink names: %w", err)
	}

	if len(targetSinks) == 0 {
		log.Printf("PipeWire: No valid bluetooth sinks found for MACs: %v", deviceMACs)
		return nil
	}

	// Build stream.rules property value
	streamRules := buildStreamRules(targetSinks)
	log.Printf("PipeWire: Setting stream.rules to: %s", streamRules)

	// Use pw-metadata to set the stream.rules property
	cmd := exec.Command("pw-metadata", "-n", "settings",
		fmt.Sprintf("%d", combinedOutput.ID), "stream.rules", streamRules)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to set stream.rules via pw-metadata: %w, output: %s", err, string(output))
	}

	log.Printf("PipeWire: Successfully updated stream.rules")
	return nil
}

// getBluetoothSinkNames converts MAC addresses to PipeWire bluetooth sink node names
func getBluetoothSinkNames(deviceMACs []string) ([]string, error) {
	// Get all audio nodes
	nodes, err := GetAllAudioNodes()
	if err != nil {
		return nil, fmt.Errorf("failed to get audio nodes: %w", err)
	}

	var sinkNames []string

	// Convert MAC addresses to the format used in PipeWire node names
	// Bluetooth devices typically appear as: bluez_output.XX_XX_XX_XX_XX_XX.1
	for _, mac := range deviceMACs {
		// Convert MAC from XX:XX:XX:XX:XX:XX to XX_XX_XX_XX_XX_XX
		nodeMAC := strings.ReplaceAll(mac, ":", "_")
		expectedNodeName := fmt.Sprintf("bluez_output.%s.1", nodeMAC)

		// Look for this node in the available nodes
		for _, node := range nodes {
			if node.NodeName == expectedNodeName {
				sinkNames = append(sinkNames, expectedNodeName)
				log.Printf("PipeWire: Found bluetooth sink: %s for MAC: %s", expectedNodeName, mac)
				break
			}
		}
	}

	if len(sinkNames) != len(deviceMACs) {
		log.Printf("PipeWire: Warning - only found %d/%d bluetooth sinks", len(sinkNames), len(deviceMACs))
	}

	return sinkNames, nil
}

// buildStreamRules creates the stream.rules structure for matching bluetooth audio devices
func buildStreamRules(targetSinks []string) string {
	if len(targetSinks) == 0 {
		return "[]"
	}

	// Create a regex pattern to match all bluetooth output devices
	// Example: "~bluez_output.B3_47_26_42_F5_DE.1|bluez_output.AA_BB_CC_DD_EE_FF.1"
	nodeNamePattern := "~" + strings.Join(targetSinks, "|")

	rule := fmt.Sprintf(`[
        {
          matches = [
          {
            media.class = "Audio/Sink"
            node.name = "%s"
          }
          ]
          actions = {
            create-stream = {
            }
          }
        }
      ]`, nodeNamePattern)

	return rule
}

// ClearStreamRules removes all stream rules from the combined output
func ClearStreamRules() error {
	log.Printf("PipeWire: Clearing stream rules")

	// Find the combined_output node
	combinedOutput, err := CheckCombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to find combined_output node: %w", err)
	}

	// Clear the stream.rules property
	cmd := exec.Command("pw-metadata", "-n", "settings",
		fmt.Sprintf("%d", combinedOutput.ID), "stream.rules", "")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to clear stream.rules via pw-metadata: %w, output: %s", err, string(output))
	}

	log.Printf("PipeWire: Successfully cleared stream.rules")
	return nil
}

// GetStreamRules retrieves the current stream.rules value from PipeWire
func GetStreamRules() (string, error) {
	// Find the combined_output node
	combinedOutput, err := CheckCombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to find combined_output node: %w", err)
	}

	// Get the stream.rules property
	cmd := exec.Command("pw-metadata", "-n", "settings",
		fmt.Sprintf("%d", combinedOutput.ID), "stream.rules")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get stream.rules via pw-metadata: %w, output: %s", err, string(output))
	}

	// Parse the output to extract the value
	// pw-metadata output format: "update: id:123 key:'stream.rules' value:'[{matches:[{node.name:"~bluez_output..."}]}]' type:''"
	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		return "", nil // No value set
	}

	// Extract value between single quotes after "value:"
	valueStart := strings.Index(outputStr, "value:'")
	if valueStart == -1 {
		return "", nil // No value found
	}
	valueStart += 7 // Skip "value:'"

	valueEnd := strings.Index(outputStr[valueStart:], "'")
	if valueEnd == -1 {
		return "", fmt.Errorf("malformed pw-metadata output: %s", outputStr)
	}

	value := outputStr[valueStart : valueStart+valueEnd]
	return value, nil
}

// InitializePipeWire initializes PipeWire at startup and displays current state
func InitializePipeWire(db *sql.DB) error {
	log.Printf("PipeWire: Initializing at startup...")

	// Check if combined_output node exists and get its ID
	combinedOutput, err := CheckCombinedOutput()
	if err != nil {
		log.Printf("PipeWire: Warning - combined_output node not found: %v", err)
		log.Printf("PipeWire: Skipping initialization (combined output may not be configured)")
		return nil // Don't fail startup if combined output doesn't exist
	}

	log.Printf("PipeWire: Found combined_output node with ID: %d", combinedOutput.ID)

	// Get current stream.rules value
	currentRules, err := GetStreamRules()
	if err != nil {
		log.Printf("PipeWire: Warning - failed to get current stream.rules: %v", err)
		currentRules = "(error reading value)"
	}

	if currentRules == "" {
		log.Printf("PipeWire: Current stream.rules value: (empty/not set)")
	} else {
		log.Printf("PipeWire: Current stream.rules value: %s", currentRules)
	}

	// Load combined audio configuration from database and apply it
	return LoadAndApplyCombinedConfig(db)
}

// LoadAndApplyCombinedConfig loads the combined audio configuration from DB and applies it to PipeWire
func LoadAndApplyCombinedConfig(db *sql.DB) error {
	log.Printf("PipeWire: Loading combined audio configuration from database...")

	combinedConfig, err := database.GetCombinedAudioConfig(db)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("PipeWire: No combined audio configuration found in database")
			// Clear any existing configuration
			if err := ClearStreamRules(); err != nil {
				log.Printf("PipeWire: Warning - failed to clear stream rules: %v", err)
			}
			return nil
		}
		return fmt.Errorf("failed to load combined audio config from database: %w", err)
	}

	if len(combinedConfig.Devices) == 0 {
		log.Printf("PipeWire: Combined audio configuration is empty, clearing PipeWire settings")
		return ClearStreamRules()
	}

	log.Printf("PipeWire: Found %d devices in combined audio configuration: %v",
		len(combinedConfig.Devices), combinedConfig.Devices)

	// Apply the configuration to PipeWire
	if err := UpdateCombinedSinks(combinedConfig.Devices); err != nil {
		return fmt.Errorf("failed to apply combined audio configuration to PipeWire: %w", err)
	}

	log.Printf("PipeWire: Successfully applied combined audio configuration from database")
	return nil
}
