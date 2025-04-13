package manifest

import (
	"encoding/json"
	"os"
)

type Chunk struct {
	Sequence   int   `json:"sequence"`
	Partitions []int `json:"partitions"`
}

type FileBlock struct {
	FilePath string  `json:"filePath"`
	Chunks   []Chunk `json:"chunks"`
}

type NodeManifest struct {
	Contents []FileBlock
}

func LoadManifest(path string) (*NodeManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var blocks []FileBlock
	if err := json.Unmarshal(data, &blocks); err != nil {
		return nil, err
	}
	return &NodeManifest{Contents: blocks}, nil
}

type Shard struct {
	Shard int
	File  string
}

func (m *NodeManifest) GetLocalShards(workerID int) []Shard {
	var result []Shard
	for _, block := range m.Contents {
		for _, chunk := range block.Chunks {
			for _, pid := range chunk.Partitions {
				if pid == workerID {
					result = append(result, Shard{chunk.Sequence, block.FilePath})
				}
			}
		}
	}

	return result
}
