package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// KeyItem 表示一个密钥项
type KeyItem struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Value       string `json:"value"`
	Description string `json:"description"`
	CreatedAt   int64  `json:"createdAt"`
	UpdatedAt   int64  `json:"updatedAt"`
}

// KeyService 提供密钥管理服务
type KeyService struct {
	mu       sync.Mutex
	dataDir  string
	filePath string
	key      []byte
	aead     cipher.AEAD
}

// NewKeyService 创建新的 KeyService
func NewKeyService() (*KeyService, error) {
	// 获取应用数据目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("get user home dir: %w", err)
	}

	dataDir := filepath.Join(homeDir, ".dev-turbo")
	filePath := filepath.Join(dataDir, "keys.dat")

	// 确保目录存在，权限仅所有者可访问
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}

	// 生成或加载加密密钥
	key, err := loadOrGenerateKey(dataDir)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create gcm: %w", err)
	}

	return &KeyService{
		dataDir:  dataDir,
		filePath: filePath,
		key:      key,
		aead:     aead,
	}, nil
}

// loadOrGenerateKey 从系统密钥环加载加密密钥；若不存在则生成并保存。
// 若存在旧版本地明文 .key 文件，会将其迁移到密钥环并删除。
func loadOrGenerateKey(dataDir string) ([]byte, error) {
	keyFile := filepath.Join(dataDir, ".key")

	// 优先从系统密钥环读取
	if key, ok, err := loadKeyFromStore(); err != nil {
		return nil, err
	} else if ok {
		// 密钥已在密钥环中，顺手清理可能残留的旧明文文件
		_ = os.Remove(keyFile)
		return key, nil
	}

	// 密钥环中没有，尝试迁移旧版本地文件
	if data, err := os.ReadFile(keyFile); err == nil {
		if len(data) != 32 {
			return nil, fmt.Errorf("existing key file has invalid length %d", len(data))
		}
		if err := saveKeyToStore(data); err != nil {
			return nil, err
		}
		_ = os.Remove(keyFile)
		return data, nil
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("read legacy key file: %w", err)
	}

	// 生成新密钥并写入密钥环
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("generate encryption key: %w", err)
	}

	if err := saveKeyToStore(key); err != nil {
		return nil, err
	}

	return key, nil
}

// encrypt 加密数据
func (k *KeyService) encrypt(plaintext string) (string, error) {
	nonce := make([]byte, k.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := k.aead.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decrypt 解密数据
func (k *KeyService) decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	nonceSize := k.aead.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := k.aead.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// GetAllKeys 获取所有密钥
func (k *KeyService) GetAllKeys() ([]KeyItem, error) {
	k.mu.Lock()
	defer k.mu.Unlock()

	data, err := os.ReadFile(k.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []KeyItem{}, nil
		}
		return nil, err
	}

	// 解密数据
	decrypted, err := k.decrypt(string(data))
	if err != nil {
		return nil, err
	}

	var keys []KeyItem
	if err := json.Unmarshal([]byte(decrypted), &keys); err != nil {
		return nil, err
	}

	return keys, nil
}

// SaveKeys 保存所有密钥
func (k *KeyService) SaveKeys(keys []KeyItem) error {
	data, err := json.Marshal(keys)
	if err != nil {
		return err
	}

	// 加密数据
	encrypted, err := k.encrypt(string(data))
	if err != nil {
		return err
	}

	// 写入临时文件，然后重命名以确保原子性
	tempFile := k.filePath + ".tmp"
	if err := os.WriteFile(tempFile, []byte(encrypted), 0600); err != nil {
		return err
	}

	if err := os.Rename(tempFile, k.filePath); err != nil {
		_ = os.Remove(tempFile)
		return err
	}

	return nil
}

// AddKey 添加新密钥
func (k *KeyService) AddKey(item KeyItem) (KeyItem, error) {
	item.Name = strings.TrimSpace(item.Name)
	if item.Name == "" {
		return KeyItem{}, fmt.Errorf("密钥名称不能为空")
	}
	if strings.TrimSpace(item.Value) == "" {
		return KeyItem{}, fmt.Errorf("密钥值不能为空")
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	keys, err := k.loadKeysLocked()
	if err != nil {
		return KeyItem{}, err
	}

	// 检查名称是否重复
	for _, key := range keys {
		if key.Name == item.Name {
			return KeyItem{}, fmt.Errorf("密钥名称已存在")
		}
	}

	rnd, err := randInt()
	if err != nil {
		return KeyItem{}, fmt.Errorf("generate key id: %w", err)
	}

	now := time.Now().UnixMilli()
	item.ID = fmt.Sprintf("key_%d_%d", now, rnd)
	item.CreatedAt = now
	item.UpdatedAt = now

	keys = append(keys, item)

	if err := k.saveKeysLocked(keys); err != nil {
		return KeyItem{}, err
	}

	return item, nil
}

// UpdateKey 更新密钥
func (k *KeyService) UpdateKey(item KeyItem) (KeyItem, error) {
	item.Name = strings.TrimSpace(item.Name)
	if item.Name == "" {
		return KeyItem{}, fmt.Errorf("密钥名称不能为空")
	}
	if strings.TrimSpace(item.Value) == "" {
		return KeyItem{}, fmt.Errorf("密钥值不能为空")
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	keys, err := k.loadKeysLocked()
	if err != nil {
		return KeyItem{}, err
	}

	found := false
	nameIndex := make(map[string]int, len(keys))
	for i, key := range keys {
		nameIndex[key.Name] = i
	}

	for i, key := range keys {
		if key.ID != item.ID {
			continue
		}

		// 检查新名称是否与其他密钥冲突
		if j, exists := nameIndex[item.Name]; exists && j != i {
			return KeyItem{}, fmt.Errorf("密钥名称已存在")
		}

		keys[i].Name = item.Name
		keys[i].Type = item.Type
		keys[i].Value = item.Value
		keys[i].Description = item.Description
		keys[i].UpdatedAt = time.Now().UnixMilli()
		item = keys[i]
		found = true
		break
	}

	if !found {
		return KeyItem{}, fmt.Errorf("密钥不存在")
	}

	if err := k.saveKeysLocked(keys); err != nil {
		return KeyItem{}, err
	}

	return item, nil
}

// DeleteKey 删除密钥
func (k *KeyService) DeleteKey(id string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	keys, err := k.loadKeysLocked()
	if err != nil {
		return err
	}

	found := false
	newKeys := make([]KeyItem, 0, len(keys))
	for _, key := range keys {
		if key.ID == id {
			found = true
			continue
		}
		newKeys = append(newKeys, key)
	}

	if !found {
		return fmt.Errorf("密钥不存在")
	}

	return k.saveKeysLocked(newKeys)
}

// loadKeysLocked 在已加锁的情况下读取密钥列表
func (k *KeyService) loadKeysLocked() ([]KeyItem, error) {
	data, err := os.ReadFile(k.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []KeyItem{}, nil
		}
		return nil, err
	}

	decrypted, err := k.decrypt(string(data))
	if err != nil {
		return nil, err
	}

	var keys []KeyItem
	if err := json.Unmarshal([]byte(decrypted), &keys); err != nil {
		return nil, err
	}

	return keys, nil
}

// saveKeysLocked 在已加锁的情况下保存密钥列表
func (k *KeyService) saveKeysLocked(keys []KeyItem) error {
	data, err := json.Marshal(keys)
	if err != nil {
		return err
	}

	encrypted, err := k.encrypt(string(data))
	if err != nil {
		return err
	}

	tempFile := k.filePath + ".tmp"
	if err := os.WriteFile(tempFile, []byte(encrypted), 0600); err != nil {
		return err
	}

	if err := os.Rename(tempFile, k.filePath); err != nil {
		_ = os.Remove(tempFile)
		return err
	}

	return nil
}

// randInt 生成随机整数
func randInt() (int, error) {
	var n uint32
	if err := binary.Read(rand.Reader, binary.BigEndian, &n); err != nil {
		return 0, err
	}
	return int(n), nil
}
