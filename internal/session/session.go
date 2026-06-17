package session

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"mia_p1_201800996/internal/disk"
	"mia_p1_201800996/internal/fs"
	"mia_p1_201800996/internal/mount"
	"mia_p1_201800996/internal/users"
)

type Session struct {
	Active         bool
	MountedID      string
	User           string
	UID            int32
	Group          string
	GID            int32
	DiskPath       string
	PartitionName  string
	PartitionStart int32
	PartitionSize  int32
}

type Manager struct {
	mu      sync.Mutex
	current Session
}

func NewManager() *Manager {
	return &Manager{}
}

var Global = NewManager()

func Login(user string, pass string, id string) (Session, error) {
	return Global.Login(user, pass, id)
}

func Logout() error {
	return Global.Logout()
}

func Current() (Session, bool) {
	return Global.Current()
}

func RequireActive() (Session, error) {
	return Global.RequireActive()
}

func RequireRoot() (Session, error) {
	return Global.RequireRoot()
}

func Clear() {
	Global.Clear()
}

func ClearIfMountedID(id string) bool {
	return Global.ClearIfMountedID(id)
}

func ClearIfDiskPath(path string) bool {
	return Global.ClearIfDiskPath(path)
}

func (m *Manager) Login(user string, pass string, id string) (Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if strings.TrimSpace(user) == "" {
		return Session{}, fmt.Errorf("login requiere -user")
	}
	if strings.TrimSpace(pass) == "" {
		return Session{}, fmt.Errorf("login requiere -pass")
	}
	if strings.TrimSpace(id) == "" {
		return Session{}, fmt.Errorf("login requiere -id")
	}
	if m.current.Active {
		return Session{}, fmt.Errorf("ya existe una sesion activa, debe ejecutar logout")
	}

	mounted, ok := mount.Global.GetMounted(id)
	if !ok {
		return Session{}, fmt.Errorf("no existe montaje con id %q", id)
	}

	file, _, err := disk.OpenReadWrite(mounted.DiskPath)
	if err != nil {
		return Session{}, err
	}
	defer file.Close()

	sb, err := fs.ReadSuperBlock(file, int64(mounted.Start))
	if err != nil {
		return Session{}, err
	}
	if sb.SMagic != 0xEF53 || sb.SFilesystemType != 2 {
		return Session{}, fmt.Errorf("la particion no esta formateada como EXT2")
	}

	content, err := fs.ReadRootUsersFile(file, sb)
	if err != nil {
		return Session{}, err
	}

	userRecord, groupRecord, err := users.Authenticate(content, user, pass)
	if err != nil {
		return Session{}, err
	}

	session := Session{
		Active:         true,
		MountedID:      mounted.ID,
		User:           userRecord.Username,
		UID:            userRecord.ID,
		Group:          groupRecord.Name,
		GID:            groupRecord.ID,
		DiskPath:       mounted.DiskPath,
		PartitionName:  mounted.PartitionName,
		PartitionStart: mounted.Start,
		PartitionSize:  mounted.Size,
	}
	m.current = session
	return session, nil
}

func (m *Manager) Logout() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.current.Active {
		return fmt.Errorf("no existe sesion activa")
	}
	m.current = Session{}
	return nil
}

func (m *Manager) Current() (Session, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.current, m.current.Active
}

func (m *Manager) RequireActive() (Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.current.Active {
		return Session{}, fmt.Errorf("necesita iniciar sesion")
	}
	return m.current, nil
}

func (m *Manager) RequireRoot() (Session, error) {
	current, err := m.RequireActive()
	if err != nil {
		return Session{}, err
	}
	if current.User != "root" || current.UID != 1 {
		return Session{}, fmt.Errorf("solo el usuario root puede ejecutar este comando")
	}
	return current, nil
}

func (m *Manager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.current = Session{}
}

func (m *Manager) ClearIfMountedID(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.current.Active || !strings.EqualFold(m.current.MountedID, id) {
		return false
	}
	m.current = Session{}
	return true
}

func (m *Manager) ClearIfDiskPath(path string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.current.Active {
		return false
	}
	if !samePath(m.current.DiskPath, path) {
		return false
	}
	m.current = Session{}
	return true
}

func samePath(a, b string) bool {
	absA, errA := filepath.Abs(a)
	absB, errB := filepath.Abs(b)
	if errA == nil {
		a = absA
	}
	if errB == nil {
		b = absB
	}
	return a == b
}
