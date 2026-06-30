package service

import (
	"strings"
	"testing"
)

func TestKeyServiceKeyringRoundTrip(t *testing.T) {
	// 使用真实系统密钥环和用户数据目录。该测试会触发一次性的 .key 文件迁移。
	svc, err := NewKeyService()
	if err != nil {
		t.Fatalf("NewKeyService failed: %v", err)
	}

	item, err := svc.AddKey(KeyItem{
		Name:        "__smoke_test_key__",
		Type:        "secret",
		Value:       "super-secret-value",
		Description: "created by go test",
	})
	if err != nil {
		t.Fatalf("AddKey failed: %v", err)
	}

	keys, err := svc.GetAllKeys()
	if err != nil {
		t.Fatalf("GetAllKeys failed: %v", err)
	}

	found := false
	for _, k := range keys {
		if k.ID == item.ID {
			found = true
			if k.Value != item.Value {
				t.Fatalf("value mismatch: got %q, want %q", k.Value, item.Value)
			}
		}
	}
	if !found {
		t.Fatalf("added key not found in list")
	}

	if err := svc.DeleteKey(item.ID); err != nil {
		t.Fatalf("DeleteKey failed: %v", err)
	}

	// 简单验证 UpdateKey 不破坏数据
	_, err = svc.UpdateKey(KeyItem{
		ID:    item.ID,
		Name:  "__smoke_test_key__updated",
		Type:  "secret",
		Value: "updated-value",
	})
	if err == nil || !strings.Contains(err.Error(), "密钥不存在") {
		t.Fatalf("UpdateKey after delete should report not found, got: %v", err)
	}
}
