package credentials

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"
)

// ErrNotFound is returned by Load when the given ref does not exist in the store.
var ErrNotFound = errors.New("credentials: ref not found")

// Store is the interface for saving and loading encrypted credentials.
type Store interface {
	Save(ref, password string) error
	Load(ref string) (string, error)
	Delete(ref string) error
	List() ([]string, error)
}

// FileStore is a Store backed by an encrypted local file.
// Passwords are encrypted with AES-256-GCM using a randomly-generated key
// that is stored alongside the credentials file. The goal is to prevent
// passwords from appearing in plaintext profile JSON files — not enterprise-
// grade secrets management.
type FileStore struct {
	mu      sync.Mutex
	dir     string
	keyPath string
	dbPath  string
}

// NewFileStore returns a FileStore that reads and writes files in dir.
func NewFileStore(dir string) *FileStore {
	return &FileStore{
		dir:     dir,
		keyPath: filepath.Join(dir, "key.bin"),
		dbPath:  filepath.Join(dir, "credentials.bin"),
	}
}

// Save encrypts password and stores it under ref.
func (s *FileStore) Save(ref, password string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key, err := s.loadOrCreateKey()
	if err != nil {
		return err
	}

	db, err := s.readDB()
	if err != nil {
		return err
	}

	encrypted, err := encrypt(key, []byte(password))
	if err != nil {
		return err
	}

	db[ref] = encrypted
	return s.writeDB(db)
}

// Load retrieves and decrypts the password stored under ref.
// Returns ErrNotFound if the ref does not exist.
func (s *FileStore) Load(ref string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key, err := s.loadOrCreateKey()
	if err != nil {
		return "", err
	}

	db, err := s.readDB()
	if err != nil {
		return "", err
	}

	encrypted, ok := db[ref]
	if !ok {
		return "", ErrNotFound
	}

	plaintext, err := decrypt(key, encrypted)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// Delete removes the entry for ref. It is not an error if ref does not exist.
func (s *FileStore) Delete(ref string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	db, err := s.readDB()
	if err != nil {
		return err
	}

	delete(db, ref)
	return s.writeDB(db)
}

// List returns all stored ref keys (not their passwords).
func (s *FileStore) List() ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	db, err := s.readDB()
	if err != nil {
		return nil, err
	}

	keys := make([]string, 0, len(db))
	for k := range db {
		keys = append(keys, k)
	}
	return keys, nil
}

// ── internal helpers ──────────────────────────────────────────────────────────

// loadOrCreateKey reads key.bin or generates and saves a new 32-byte key.
func (s *FileStore) loadOrCreateKey() ([]byte, error) {
	data, err := os.ReadFile(s.keyPath)
	if err == nil {
		key, decErr := hex.DecodeString(string(data))
		if decErr == nil && len(key) == 32 {
			return key, nil
		}
	}

	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}
	if err := os.WriteFile(s.keyPath, []byte(hex.EncodeToString(key)), 0600); err != nil {
		return nil, err
	}
	return key, nil
}

// readDB loads credentials.bin into a map. Returns an empty map if the file
// does not exist yet.
func (s *FileStore) readDB() (map[string]string, error) {
	db := map[string]string{}
	data, err := os.ReadFile(s.dbPath)
	if errors.Is(err, os.ErrNotExist) {
		return db, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &db); err != nil {
		return nil, err
	}
	return db, nil
}

// writeDB serialises db and overwrites credentials.bin.
func (s *FileStore) writeDB(db map[string]string) error {
	data, err := json.Marshal(db)
	if err != nil {
		return err
	}
	return os.WriteFile(s.dbPath, data, 0600)
}

// encrypt encrypts plaintext with AES-256-GCM and returns a hex-encoded string
// containing the nonce prepended to the ciphertext.
func encrypt(key, plaintext []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return hex.EncodeToString(ciphertext), nil
}

// decrypt reverses encrypt. Returns the original plaintext.
func decrypt(key []byte, encoded string) ([]byte, error) {
	data, err := hex.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("credentials: ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
