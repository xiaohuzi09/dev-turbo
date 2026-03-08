package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
	dataDir  string
	filePath string
	key      []byte
}

// NewKeyService 创建新的 KeyService
func NewKeyService() *KeyService {
	// 获取应用数据目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	dataDir := filepath.Join(homeDir, ".dev-turbo")
	filePath := filepath.Join(dataDir, "keys.dat")

	// 确保目录存在
	os.MkdirAll(dataDir, 0755)

	// 生成或加载加密密钥
	key := loadOrGenerateKey(dataDir)

	return &KeyService{
		dataDir:  dataDir,
		filePath: filePath,
		key:      key,
	}
}

// loadOrGenerateKey 加载或生成加密密钥
func loadOrGenerateKey(dataDir string) []byte {
	keyFile := filepath.Join(dataDir, ".key")

	// 尝试加载已有密钥
	if data, err := os.ReadFile(keyFile); err == nil && len(data) == 32 {
		return data
	}

	// 生成新密钥
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		// 如果随机生成失败，使用基于机器信息的派生密钥
		hostname, _ := os.Hostname()
		hash := sha256.Sum256([]byte(hostname + "dev-turbo-salt"))
		return hash[:]
	}

	// 保存密钥
	os.WriteFile(keyFile, key, 0600)
	return key
}

// encrypt 加密数据
func (k *KeyService) encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(k.key)
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

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decrypt 解密数据
func (k *KeyService) decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(k.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// GetAllKeys 获取所有密钥
func (k *KeyService) GetAllKeys() ([]KeyItem, error) {
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

	return os.Rename(tempFile, k.filePath)
}

// AddKey 添加新密钥
func (k *KeyService) AddKey(item KeyItem) (KeyItem, error) {
	keys, err := k.GetAllKeys()
	if err != nil {
		return KeyItem{}, err
	}

	// 检查名称是否重复
	for _, key := range keys {
		if key.Name == item.Name {
			return KeyItem{}, fmt.Errorf("密钥名称已存在")
		}
	}

	now := time.Now().UnixMilli()
	item.ID = fmt.Sprintf("key_%d_%d", now, randInt())
	item.CreatedAt = now
	item.UpdatedAt = now

	keys = append(keys, item)

	if err := k.SaveKeys(keys); err != nil {
		return KeyItem{}, err
	}

	return item, nil
}

// UpdateKey 更新密钥
func (k *KeyService) UpdateKey(item KeyItem) (KeyItem, error) {
	keys, err := k.GetAllKeys()
	if err != nil {
		return KeyItem{}, err
	}

	found := false
	for i, key := range keys {
		if key.ID == item.ID {
			// 检查新名称是否与其他密钥冲突
			for j, other := range keys {
				if j != i && other.Name == item.Name {
					return KeyItem{}, fmt.Errorf("密钥名称已存在")
				}
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
	}

	if !found {
		return KeyItem{}, fmt.Errorf("密钥不存在")
	}

	if err := k.SaveKeys(keys); err != nil {
		return KeyItem{}, err
	}

	return item, nil
}

// DeleteKey 删除密钥
func (k *KeyService) DeleteKey(id string) error {
	keys, err := k.GetAllKeys()
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

	return k.SaveKeys(newKeys)
}

// randInt 生成随机整数
func randInt() int {
	b := make([]byte, 4)
	rand.Read(b)
	return int(b[0])<<24 | int(b[1])<<16 | int(b[2])<<8 | int(b[3])
}
