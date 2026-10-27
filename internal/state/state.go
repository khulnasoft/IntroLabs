package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// InstanceState tracks the running state of a lab instance.
type InstanceState struct {
	LabID        string     `json:"labID"`
	Version      string     `json:"version"`
	Status       string     `json:"status"` // running, stopped, removed
	StartedAt    time.Time  `json:"startedAt"`
	StoppedAt    *time.Time `json:"stoppedAt,omitempty"`
	Ports        []int      `json:"ports"`
	URLs         []string   `json:"urls"`
	TemplatePath string     `json:"templatePath"`
	ComposeFile  string     `json:"composeFile"`
	ProjectName  string     `json:"projectName"`
}

// Store persists instance state to disk.
type Store struct {
	dir       string
	instances map[string]*InstanceState
	mu        sync.RWMutex
}

// NewStore creates a state store in the given directory.
func NewStore(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("creating state directory: %w", err)
	}

	s := &Store{
		dir:       dir,
		instances: make(map[string]*InstanceState),
	}

	if err := s.load(); err != nil {
		return nil, fmt.Errorf("loading state: %w", err)
	}

	return s, nil
}

// load reads all instance state files from disk.
func (s *Store) load() error {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(s.dir, entry.Name()))
		if err != nil {
			continue
		}

		var inst InstanceState
		if err := json.Unmarshal(data, &inst); err != nil {
			continue
		}

		s.instances[inst.LabID] = &inst
	}

	return nil
}

// save writes a single instance state to disk.
func (s *Store) save(inst *InstanceState) error {
	data, err := json.MarshalIndent(inst, "", "  ")
	if err != nil {
		return err
	}

	filename := safeFilename(inst.LabID) + ".json"
	if err := os.WriteFile(filepath.Join(s.dir, filename), data, 0644); err != nil {
		return err
	}

	return nil
}

// safeFilename sanitizes a lab ID for use as a filename.
func safeFilename(labID string) string {
	return strings.ReplaceAll(labID, "/", "_")
}

// Create records a new running instance.
func (s *Store) Create(labID, version, templatePath, composeFile, projectName string, ports []int, urls []string) *InstanceState {
	inst := &InstanceState{
		LabID:        labID,
		Version:      version,
		Status:       "running",
		StartedAt:    time.Now(),
		Ports:        ports,
		URLs:         urls,
		TemplatePath: templatePath,
		ComposeFile:  composeFile,
		ProjectName:  projectName,
	}

	s.mu.Lock()
	s.instances[labID] = inst
	s.mu.Unlock()

	if err := s.save(inst); err != nil {
		// Log but don't fail the run
		_ = err
	}

	return inst
}

// Get returns the instance state for a lab ID, or nil if not found.
func (s *Store) Get(labID string) *InstanceState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.instances[labID]
}

// List returns all tracked instances.
func (s *Store) List() []*InstanceState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	instances := make([]*InstanceState, 0, len(s.instances))
	for _, inst := range s.instances {
		instances = append(instances, inst)
	}
	return instances
}

// UpdateStatus changes the status of an instance.
func (s *Store) UpdateStatus(labID, status string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	inst, ok := s.instances[labID]
	if !ok {
		return fmt.Errorf("instance not found: %s", labID)
	}

	inst.Status = status
	now := time.Now()
	if status == "stopped" {
		inst.StoppedAt = &now
	} else if status == "running" {
		inst.StoppedAt = nil
	}

	return s.save(inst)
}

// Remove deletes an instance from state.
func (s *Store) Remove(labID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.instances[labID]; !ok {
		return fmt.Errorf("instance not found: %s", labID)
	}

	delete(s.instances, labID)

	filename := safeFilename(labID) + ".json"
	return os.Remove(filepath.Join(s.dir, filename))
}

// Running returns all instances with status "running".
func (s *Store) Running() []*InstanceState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var running []*InstanceState
	for _, inst := range s.instances {
		if inst.Status == "running" {
			running = append(running, inst)
		}
	}
	return running
}

// DefaultStateDir returns the default state directory path.
func DefaultStateDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".labkit-state"
	}
	return filepath.Join(home, ".labkit", "state")
}
