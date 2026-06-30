package service

import (
	"fmt"
	"runtime"

	"github.com/zalando/go-keyring"
)

const (
	keyringService = "dev-turbo"
	keyringUser    = "encryption-key"
)

// loadKeyFromStore 从系统密钥环读取加密密钥。
// 若密钥不存在，返回 (nil, false, nil)。
func loadKeyFromStore() ([]byte, bool, error) {
	data, err := keyring.Get(keyringService, keyringUser)
	if err == keyring.ErrNotFound {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("read key from %s keyring: %w", runtime.GOOS, err)
	}

	key := []byte(data)
	if len(key) != 32 {
		return nil, false, fmt.Errorf("keyring entry has invalid length %d", len(key))
	}

	return key, true, nil
}

// saveKeyToStore 将加密密钥写入系统密钥环。
func saveKeyToStore(key []byte) error {
	if len(key) != 32 {
		return fmt.Errorf("invalid key length %d", len(key))
	}

	if err := keyring.Set(keyringService, keyringUser, string(key)); err != nil {
		return fmt.Errorf("save key to %s keyring: %w", runtime.GOOS, err)
	}
	return nil
}

// deleteKeyFromStore 从系统密钥环删除加密密钥。
func deleteKeyFromStore() error {
	if err := keyring.Delete(keyringService, keyringUser); err != nil && err != keyring.ErrNotFound {
		return fmt.Errorf("delete key from %s keyring: %w", runtime.GOOS, err)
	}
	return nil
}
