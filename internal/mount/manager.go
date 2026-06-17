package mount

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"mia_p1_201800996/internal/partition"
)

const carnetPrefix = "96"

type MountedPartition struct {
	ID             string
	DiskPath       string
	PartitionName  string
	PartitionIndex int
	PartitionType  byte
	Start          int32
	Size           int32
	DiskNumber     int
	Letter         byte
}

type Manager struct {
	mu          sync.Mutex
	mounts      map[string]MountedPartition
	diskNumbers map[string]int
	nextDiskNum int
}

func NewManager() *Manager {
	return &Manager{
		mounts:      map[string]MountedPartition{},
		diskNumbers: map[string]int{},
		nextDiskNum: 1,
	}
}

var Global = NewManager()

func (m *Manager) Mount(path, name string) (MountedPartition, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	absPath, err := filepath.Abs(path)
	if err != nil {
		return MountedPartition{}, err
	}
	if m.isMountedLocked(absPath, name) {
		return MountedPartition{}, fmt.Errorf("la particion %q ya esta montada", name)
	}

	part, index, err := partition.SearchPartition(absPath, name)
	if err != nil {
		return MountedPartition{}, err
	}
	if part.PartType != 'P' && part.PartType != 'E' {
		return MountedPartition{}, fmt.Errorf("solo se pueden montar particiones primarias o extendidas en esta fase")
	}

	diskNumber := m.diskNumberLocked(absPath)
	letter, err := m.nextLetterLocked(absPath)
	if err != nil {
		return MountedPartition{}, err
	}
	id := fmt.Sprintf("%s%d%c", carnetPrefix, diskNumber, letter)
	mounted := MountedPartition{
		ID:             id,
		DiskPath:       absPath,
		PartitionName:  name,
		PartitionIndex: index,
		PartitionType:  part.PartType,
		Start:          part.PartStart,
		Size:           part.PartSize,
		DiskNumber:     diskNumber,
		Letter:         letter,
	}
	m.mounts[strings.ToLower(id)] = mounted
	return mounted, nil
}

func (m *Manager) Unmount(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := strings.ToLower(id)
	if _, ok := m.mounts[key]; !ok {
		return fmt.Errorf("no existe montaje con id %q", id)
	}
	delete(m.mounts, key)
	return nil
}

func (m *Manager) GetMounted(id string) (MountedPartition, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	mounted, ok := m.mounts[strings.ToLower(id)]
	return mounted, ok
}

func (m *Manager) IsMounted(path, name string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	return m.isMountedLocked(absPath, name)
}

func (m *Manager) UnmountByDisk(path string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	absPath, err := filepath.Abs(path)
	if err != nil {
		return
	}
	for key, mounted := range m.mounts {
		if samePath(mounted.DiskPath, absPath) {
			delete(m.mounts, key)
		}
	}
}

func (m *Manager) List() []MountedPartition {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make([]MountedPartition, 0, len(m.mounts))
	for _, mounted := range m.mounts {
		result = append(result, mounted)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})
	return result
}

func (m *Manager) diskNumberLocked(path string) int {
	for diskPath, number := range m.diskNumbers {
		if samePath(diskPath, path) {
			return number
		}
	}
	number := m.nextDiskNum
	m.nextDiskNum++
	m.diskNumbers[path] = number
	return number
}

func (m *Manager) nextLetterLocked(path string) (byte, error) {
	used := map[byte]bool{}
	for _, mounted := range m.mounts {
		if samePath(mounted.DiskPath, path) {
			used[mounted.Letter] = true
		}
	}
	for letter := byte('A'); letter <= byte('Z'); letter++ {
		if !used[letter] {
			return letter, nil
		}
	}
	return 0, fmt.Errorf("no hay letras disponibles para el disco")
}

func (m *Manager) isMountedLocked(path, name string) bool {
	for _, mounted := range m.mounts {
		if samePath(mounted.DiskPath, path) && mounted.PartitionName == name {
			return true
		}
	}
	return false
}

func samePath(a, b string) bool {
	cleanA, errA := filepath.Abs(a)
	cleanB, errB := filepath.Abs(b)
	if errA == nil {
		a = cleanA
	}
	if errB == nil {
		b = cleanB
	}
	return a == b
}
