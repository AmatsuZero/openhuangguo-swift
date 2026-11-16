package sdk

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

// Session 是 token 快照。Origin 绑定 + Generation 防止跨 base 误用旧 token。
type Session struct {
	Origin     string          `json:"origin"`
	Token      string          `json:"token"`
	Generation int64           `json:"generation"`
	User       json.RawMessage `json:"user"`
}

// TokenStore 是 Go SDK 内部 token 存储接口（多返回值，不受 gomobile 限制）。
// MemStore / FileStore 实现在 Task 3。
type TokenStore interface {
	Load() (Session, error)
	Save(Session) error
	Clear() error
}

// MemStore 是线程安全的内存 TokenStore（移动端默认；原生层经 core 的 token 桥接回写持久化）。
type MemStore struct {
	mu sync.Mutex
	s  Session
}

func NewMemStore() *MemStore { return &MemStore{} }

func (m *MemStore) Load() (Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.s, nil
}

func (m *MemStore) Save(s Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.s = s
	return nil
}

func (m *MemStore) Clear() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.s = Session{}
	return nil
}

// FileStore 是文件 TokenStore：0600 权限、tmp+rename 原子写、明确错误传播。
type FileStore struct {
	mu   sync.Mutex
	path string
}

func NewFileStore(path string) *FileStore { return &FileStore{path: path} }

func (f *FileStore) Load() (Session, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	b, err := os.ReadFile(f.path)
	if errors.Is(err, os.ErrNotExist) {
		return Session{}, nil
	}
	if err != nil {
		return Session{}, err
	}
	var s Session
	if err := json.Unmarshal(b, &s); err != nil {
		return Session{}, err
	}
	return s, nil
}

func (f *FileStore) Save(s Session) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}
	dir := filepath.Dir(f.path)
	tmp, err := os.CreateTemp(dir, ".huangguo-session-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(0600); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, f.path)
}

// Clear 删除文件；删除失败（非"不存在"）必须返回错误，绝不静默成功。
func (f *FileStore) Clear() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	err := os.Remove(f.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}
