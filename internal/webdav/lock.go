package webdav

/*
涉及webdav的锁系统的实现。东西比较多，但主要是注释比较多。

*/

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// 说实话Condition写得有点儿烂。
type Condition struct {
	Not   bool      // 必须符合条件，或必须不符合条件
	Token uuid.UUID // 有时候会给一个文件上锁，要求只有指定token的人可以使用。Condition的Token便是那个被指定的Token
}

type LockDetails struct {
	RootPath  string        //被锁定文件的路径，如/doc/test.md
	Duration  time.Duration // 锁的有效持续时间
	Owner     string        // 锁的所有者信息
	ZeroDepth bool          //涉及锁的深度，即只是锁文件本身，还是该文件及所有子文件。在实现锁文件夹时很有必要。
}

type lockNode struct {
	details LockDetails // 锁的元信息
	token   uuid.UUID   // 锁的唯一标识符，也是发给客户端的钥匙。选用UUID类型纯粹是觉得自增token太蠢了。
	expiry  time.Time   // 通过duration计算得出的一个时间点，意为“x时x刻，这个lockNode正式过期”
	held    bool        // 该锁是否正被Confirm验证中，防止出现并发冲突
}

type LockSystem interface {
	// Confirm确定用户是否有权操作。
	// 两个name是什么意思？因为诸如删除这样的操作，确实是只涉及一个文件。但倘若我要把文件A复制到文件夹B当中，那我就需要同时考虑A和B两者的上锁情况了
	Confirm(now time.Time, name0 string, name1 string, conditions ...Condition) (release func(), err error)
	// 创建一个新锁
	Create(now time.Time, details LockDetails) (token uuid.UUID, err error)
	// 刷新已有的锁 (正如前面所说，一个锁是有时间的)
	Refresh(now time.Time, token uuid.UUID, duration time.Duration) (LockDetails, error)
	// 移除一个锁
	Unlock(now time.Time, token uuid.UUID) error
}

type MemoryLockSystem struct {
	mu           sync.Mutex              // 互斥锁
	locksByPath  map[string]*lockNode    // 通过文件路径，检测该文件处是否有锁
	locksByToken map[uuid.UUID]*lockNode // 通过token,检索该文件处是否有锁
}

func NewMemoryLockSystem() LockSystem {
	return &MemoryLockSystem{
		locksByPath:  make(map[string]*lockNode),
		locksByToken: make(map[uuid.UUID]*lockNode),
	}
}

func (m *MemoryLockSystem) Confirm(now time.Time, name0 string, name1 string, conditions ...Condition) (func(), error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 先清理过期锁，再继续其他操作
	m.CleanExpiredLocks(now)

	var heldLocks []*lockNode // 存储1或2个要上锁的文件/文件夹的锁节点

	if name0 != "" {
		if lock, exists := m.isPathLocked(name0); exists {

			// 这里本应有一个Condition验证的环节的，但是我在写demo,所以先搁置
			if !m.ValidateConditions(lock, conditions) {
				return nil, fmt.Errorf("路径 %s 的锁条件验证失败", name0)
			}
			if lock.held {
				return nil, fmt.Errorf("路径 %s 的锁正被占用", name0)
			}

			lock.held = true
			heldLocks = append(heldLocks, lock)

		}
	}

	if name1 != "" && name1 != name0 {
		if lock, exists := m.isPathLocked(name1); exists {

			// 这里本应有一个Condition验证的环节的，但是我在写demo,所以先搁置
			if !m.ValidateConditions(lock, conditions) {
				// 回滚，因为你在name1处失败了，所以你得把name0处的held给置回原位
				for _, heldLock := range heldLocks {
					heldLock.held = false
				}
				return nil, fmt.Errorf("路径 %s 的锁条件验证失败", name1)
			}
			if lock.held {
				return nil, fmt.Errorf("路径 %s 的锁正被占用", name1)
			}

			lock.held = true
			heldLocks = append(heldLocks, lock)
		}
	}

	release := func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		for _, lock := range heldLocks {
			lock.held = false
		}
	}
	return release, nil
}

func (m *MemoryLockSystem) Create(now time.Time, lockDetails LockDetails) (uuid.UUID, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.CleanExpiredLocks(now)

	// 检查路径有没有被锁
	if lock, exists := m.isPathLocked(lockDetails.RootPath); exists {
		if now.Before(lock.expiry) {
			return uuid.Nil, fmt.Errorf("路径 %s 或其父节点已经被锁定了", lockDetails.RootPath)
		}
	}

	// 假如你要创建的是深度锁，那你就需要考虑该深度锁下面的子路径，是不是已经被锁定了
	if !lockDetails.ZeroDepth {
		for path, locks := range m.locksByPath {
			if isSubPath(path, lockDetails.RootPath) && now.Before(locks.expiry) {
				return uuid.Nil, fmt.Errorf("无法创建深度锁，子路径 %s 已被锁定", path)
			}
		}
	}

	// 创建新锁
	token, err := uuid.NewV7()
	if err != nil {
		return uuid.Nil, fmt.Errorf("新建token失败: %w", err)
	}
	lock := &lockNode{
		details: lockDetails,
		token:   token,
		expiry:  now.Add(lockDetails.Duration), // 我说什么来着，expiry就是now+duration
		held:    false,
	}
	m.locksByPath[lockDetails.RootPath] = lock
	m.locksByToken[token] = lock

	return token, nil
}

func (m *MemoryLockSystem) Refresh(now time.Time, token uuid.UUID, duration time.Duration) (LockDetails, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	lock, exists := m.locksByToken[token]
	if !exists {
		return LockDetails{}, fmt.Errorf("未找到制定锁")
	}

	// 锁过期了
	m.CleanExpiredLocks(now)

	// 执行刷新
	lock.expiry = now.Add(duration)
	lock.details.Duration = duration

	return lock.details, nil
}

func (m *MemoryLockSystem) Unlock(now time.Time, token uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	lock, exists := m.locksByToken[token]
	if !exists {
		return fmt.Errorf("未找到制定锁")
	}

	delete(m.locksByPath, lock.details.RootPath)
	delete(m.locksByToken, token)

	return nil
}

// 用于清理过期锁
func (m *MemoryLockSystem) CleanExpiredLocks(now time.Time) {
	for path, lock := range m.locksByPath {
		if now.After(lock.expiry) {
			delete(m.locksByPath, path)
			delete(m.locksByToken, lock.token)
		}
	}
}

func (m *MemoryLockSystem) ValidateConditions(lock *lockNode, conditions []Condition) bool {
	for _, condition := range conditions {
		hasToken := (lock.token == condition.Token)
		if condition.Not {
			if hasToken {
				return false
			}
		} else {
			if !hasToken {
				return false
			}
		}
	}
	return true
}

// 考虑深度锁，检查路径是否被锁定
func (m *MemoryLockSystem) isPathLocked(targetPath string) (*lockNode, bool) {
	// 精确匹配
	if lock, exists := m.locksByPath[targetPath]; exists {
		return lock, true
	}

	//
	for lockedPath, lock := range m.locksByPath {
		if !lock.details.ZeroDepth && isSubPath(targetPath, lockedPath) {
			return lock, true
		}
	}

	return nil, false
}

// 用于判断是否是子路径
func isSubPath(subPath string, parentPath string) bool {
	if parentPath == "/" {
		return true
	}
	return strings.HasPrefix(subPath, parentPath+"/")
}
