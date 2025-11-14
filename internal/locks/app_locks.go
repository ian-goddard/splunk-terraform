package locks

import (
	"context"
	"fmt"
	"sync"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	v2 "github.com/splunk/terraform-provider-scp/acs/v2"
)

// AppLockManager manages per-stack mutexes for app operations
type AppLockManager struct {
	mu     sync.RWMutex
	locks  map[string]*sync.Mutex
	counts map[string]int // track how many operations are using each lock
}

var (
	// Global instance of the lock manager
	globalLockManager = &AppLockManager{
		locks:  make(map[string]*sync.Mutex),
		counts: make(map[string]int),
	}
)

// GetAppLockManager returns the global app lock manager instance
func GetAppLockManager() *AppLockManager {
	return globalLockManager
}

// getLock gets or creates a mutex for the given stack
func (alm *AppLockManager) getLock(stack string) *sync.Mutex {
	alm.mu.Lock()
	defer alm.mu.Unlock()

	if _, exists := alm.locks[stack]; !exists {
		alm.locks[stack] = &sync.Mutex{}
		alm.counts[stack] = 0
	}
	alm.counts[stack]++
	return alm.locks[stack]
}

// releaseLock decrements the usage count and cleans up unused locks
func (alm *AppLockManager) releaseLock(stack string) {
	alm.mu.Lock()
	defer alm.mu.Unlock()

	if count, exists := alm.counts[stack]; exists {
		alm.counts[stack] = count - 1
		if alm.counts[stack] <= 0 {
			delete(alm.locks, stack)
			delete(alm.counts, stack)
		}
	}
}

// LockAppOperation acquires a lock for app operations on the specified stack
// Returns an unlock function that should be called when the operation is complete
func (alm *AppLockManager) LockAppOperation(ctx context.Context, stack v2.Stack, operation string) func() {
	stackStr := string(stack)

	tflog.Info(ctx, fmt.Sprintf("Acquiring app operation lock for stack '%s', operation: %s", stackStr, operation))

	lock := alm.getLock(stackStr)
	lock.Lock()

	tflog.Info(ctx, fmt.Sprintf("Acquired app operation lock for stack '%s', operation: %s", stackStr, operation))

	return func() {
		tflog.Info(ctx, fmt.Sprintf("Releasing app operation lock for stack '%s', operation: %s", stackStr, operation))
		lock.Unlock()
		alm.releaseLock(stackStr)
		tflog.Info(ctx, fmt.Sprintf("Released app operation lock for stack '%s', operation: %s", stackStr, operation))
	}
}

// WithAppLock is a convenience function that executes a function while holding an app lock
func (alm *AppLockManager) WithAppLock(ctx context.Context, stack v2.Stack, operation string, fn func() error) error {
	unlock := alm.LockAppOperation(ctx, stack, operation)
	defer unlock()
	return fn()
}
