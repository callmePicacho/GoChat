package util

import (
	"GoChat/config"
	"GoChat/pkg/db"
	"sync"
	"testing"
)

// 测试是否有序 and 重启后是否有序
func TestUid_Seq_NextId(t *testing.T) {
	// 初始化配置
	config.InitConfig("../../app.yaml")
	db.InitMySQL(config.GlobalConfig.MySQL.DNS)

	uid := newUid("TestUid_Seq_NextId")

	for i := 0; i < UidStep*2-4; i++ {
		id, err := uid.nextId()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		t.Log(id)
	}
}

func TestUid_NextId(t *testing.T) {
	// 初始化配置
	config.InitConfig("../../app.yaml")
	db.InitMySQL(config.GlobalConfig.MySQL.DNS)

	// Create a new UID
	uid := newUid("test")

	// Test with multiple goroutines
	var wg sync.WaitGroup
	n := 10
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := uid.nextId()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		}()
	}
	wg.Wait()
}

func TestUidGenerator_GetNextIds(t *testing.T) {
	// 初始化配置
	config.InitConfig("../../app.yaml")
	db.InitMySQL(config.GlobalConfig.MySQL.DNS)

	gen := NewGeneratorUid()
	businessIds := []string{"user", "order", "product"}

	// 测试获取一批 id
	ids, err := gen.GetNextIds(businessIds)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(ids) != len(businessIds) {
		t.Errorf("Expected %d ids, but got %d", len(businessIds), len(ids))
	}

	// 测试获取多批 id
	ids1, err := gen.GetNextIds(businessIds)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	ids2, err := gen.GetNextIds(businessIds)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(ids1) != len(businessIds) || len(ids2) != len(businessIds) {
		t.Errorf("Expected %d ids, but got %d and %d", len(businessIds), len(ids1), len(ids2))
	}

	var wg sync.WaitGroup
	n := 10
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := gen.GetNextIds(businessIds)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		}()
	}
	wg.Wait()
	ids1, err = gen.GetNextIds(businessIds)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	t.Log(ids1)
}

func BenchmarkUidGenerator_GetNextIds(b *testing.B) {
	// 初始化配置
	config.InitConfig("../../app.yaml")
	db.InitMySQL(config.GlobalConfig.MySQL.DNS)

	gen := NewGeneratorUid()
	// 构造测试数据
	businessIds := make([]string, 10)
	for i := 0; i < 10; i++ {
		businessIds[i] = Int64ToStr(int64(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gen.GetNextIds(businessIds)
		if err != nil {
			b.Errorf("Unexpected error: %v", err)
		}
	}
}
