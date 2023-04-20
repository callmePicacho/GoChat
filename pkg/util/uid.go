package util

import (
	"GoChat/pkg/db"
	"database/sql"
	"gorm.io/gorm"
	"sync"
)

const (
	UidStep = 1000
)

var (
	UidGen = NewGeneratorUid()
)

// uid 发号器
type uidGenerator struct {
	batchUidMap map[string]*uid // 存在发号器中的一批 uid，其中 k 为 businessId，v 为 cur_seq
	mu          sync.Mutex
}

func NewGeneratorUid() *uidGenerator {
	return &uidGenerator{
		batchUidMap: make(map[string]*uid),
	}
}

// GetNextId 获取下一个 id
func (u *uidGenerator) GetNextId(businessId string) (uint64, error) {
	u.mu.Lock()
	defer u.mu.Unlock()
	if uid, ok := u.batchUidMap[businessId]; ok {
		return uid.nextId()
	}
	uid := newUid(businessId)
	u.batchUidMap[businessId] = uid
	return uid.nextId()
}

// GetNextIds 获取一批 businessId
func (u *uidGenerator) GetNextIds(businessIds []string) ([]uint64, error) {
	result := make([]uint64, 0, len(businessIds))
	for _, businessId := range businessIds {
		id, err := u.GetNextId(businessId)
		if err != nil {
			return nil, err
		}
		result = append(result, id)
	}
	return result, nil
}

type uid struct {
	businessId string // 业务id
	curId      uint64 // 当前分配的 id
	maxId      uint64 // 当前号段最大 id
	step       int    // 每次分配出的号段步长
	mu         sync.Mutex
}

func newUid(businessId string) *uid {
	id := &uid{
		businessId: businessId,
		curId:      0,
		maxId:      0,
		step:       UidStep,
	}
	return id
}

// 假设 step = 1000 时，
// 首次获取，cur_id = 1, max_id = 1000，取出号段 [1, 1000]
// 再次获取，cur_id = 1001, max_id = 2000，取出号段 [1001, 2000]
func (u *uid) nextId() (uint64, error) {
	// 加锁保证并发安全
	u.mu.Lock()
	defer u.mu.Unlock()

	// 判断是否需要更新 ID 段
	if u.curId == u.maxId {
		err := u.getFromDB()
		if err != nil {
			return 0, err
		}
	}

	u.curId++
	return u.curId, nil
}

// 从数据库拉取id段
// 如果存在，cur_id 从 max_id 开始，max_id = max_id + step，分配出去 [step, max_id + step)
func (u *uid) getFromDB() error {
	var (
		maxId uint64
		step  int
	)
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		// 查询
		err := tx.Raw("select max_id, step from uid where business_id = ? for update", u.businessId).Row().Scan(&maxId, &step)
		if err != nil && err != sql.ErrNoRows {
			return err
		}
		// 不存在就插入
		if err == sql.ErrNoRows {
			err = tx.Exec("insert into uid(business_id, max_id, step) values(?,?,?)", u.businessId, u.maxId, u.step).Error
			if err != nil {
				return err
			}

		} else {
			// 存在就更新
			err = tx.Exec("update uid set max_id = max_id + step where business_id = ?", u.businessId).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	if maxId != 0 {
		// 如果已存在，cur_id = max_id
		u.curId = maxId
	}
	u.maxId = maxId + uint64(step)
	u.step = step
	return nil
}
