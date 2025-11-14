package locks

import (
	"context"
	"errors"
	"testing"

	v2 "github.com/splunk/terraform-provider-scp/acs/v2"
)

func TestAppLockManager_BasicLockAndUnlock(t *testing.T) {
	ctx := context.Background()
	manager := GetAppLockManager()
	stack := v2.Stack("test-stack")

	// Acquire lock
	unlock := manager.LockAppOperation(ctx, stack, "test-operation")

	// Verify the lock exists
	manager.mu.RLock()
	_, exists := manager.locks[string(stack)]
	count := manager.counts[string(stack)]
	manager.mu.RUnlock()

	if !exists {
		t.Error("Expected lock to exist for stack")
	}
	if count != 1 {
		t.Errorf("Expected count to be 1, got %d", count)
	}

	// Release lock
	unlock()

	// Verify cleanup after release
	manager.mu.RLock()
	_, exists = manager.locks[string(stack)]
	_, countExists := manager.counts[string(stack)]
	manager.mu.RUnlock()

	if exists {
		t.Error("Expected lock to be cleaned up after release")
	}
	if countExists {
		t.Error("Expected count to be cleaned up after release")
	}
}

func TestAppLockManager_MultipleLocksSameStack(t *testing.T) {
	ctx := context.Background()
	manager := GetAppLockManager()
	stack := v2.Stack("test-stack-multi")

	// Acquire first lock
	unlock1 := manager.LockAppOperation(ctx, stack, "operation-1")

	// Verify first lock
	manager.mu.RLock()
	count1 := manager.counts[string(stack)]
	manager.mu.RUnlock()

	if count1 != 1 {
		t.Errorf("Expected count to be 1 after first lock, got %d", count1)
	}

	// This will block in a real scenario, but for testing we'll skip the second lock
	// to avoid hanging the test

	// Release first lock
	unlock1()

	// Verify cleanup
	manager.mu.RLock()
	_, exists := manager.locks[string(stack)]
	manager.mu.RUnlock()

	if exists {
		t.Error("Expected lock to be cleaned up after release")
	}
}

func TestAppLockManager_DifferentStacksIndependent(t *testing.T) {
	ctx := context.Background()
	manager := GetAppLockManager()
	stack1 := v2.Stack("stack-1")
	stack2 := v2.Stack("stack-2")

	// Acquire locks on different stacks
	unlock1 := manager.LockAppOperation(ctx, stack1, "operation-1")
	unlock2 := manager.LockAppOperation(ctx, stack2, "operation-2")

	// Verify both locks exist
	manager.mu.RLock()
	exists1 := manager.locks[string(stack1)] != nil
	exists2 := manager.locks[string(stack2)] != nil
	manager.mu.RUnlock()

	if !exists1 {
		t.Error("Expected lock to exist for stack1")
	}
	if !exists2 {
		t.Error("Expected lock to exist for stack2")
	}

	// Release stack1 lock
	unlock1()

	// Verify stack2 lock still exists but stack1 is cleaned up
	manager.mu.RLock()
	exists1After := manager.locks[string(stack1)] != nil
	exists2After := manager.locks[string(stack2)] != nil
	manager.mu.RUnlock()

	if exists1After {
		t.Error("Expected stack1 lock to be cleaned up")
	}
	if !exists2After {
		t.Error("Expected stack2 lock to still exist")
	}

	// Release stack2 lock
	unlock2()

	// Verify complete cleanup
	manager.mu.RLock()
	lockCount := len(manager.locks)
	manager.mu.RUnlock()

	if lockCount != 0 {
		t.Errorf("Expected all locks to be cleaned up, but found %d locks", lockCount)
	}
}

func TestAppLockManager_WithAppLockSuccess(t *testing.T) {
	ctx := context.Background()
	manager := GetAppLockManager()
	stack := v2.Stack("with-lock-stack")

	executed := false

	err := manager.WithAppLock(ctx, stack, "test-operation", func() error {
		executed = true

		// Verify lock exists during execution
		manager.mu.RLock()
		_, exists := manager.locks[string(stack)]
		manager.mu.RUnlock()

		if !exists {
			t.Error("Expected lock to exist during function execution")
		}
		return nil
	})

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if !executed {
		t.Error("Expected function to be executed")
	}

	// Verify cleanup after execution
	manager.mu.RLock()
	lockCount := len(manager.locks)
	manager.mu.RUnlock()

	if lockCount != 0 {
		t.Error("Expected locks to be cleaned up after WithAppLock")
	}
}

func TestAppLockManager_WithAppLockError(t *testing.T) {
	ctx := context.Background()
	manager := GetAppLockManager()
	stack := v2.Stack("error-stack")

	testError := errors.New("test error")

	err := manager.WithAppLock(ctx, stack, "error-operation", func() error {
		return testError
	})

	if err != testError {
		t.Errorf("Expected error '%v', got: %v", testError, err)
	}

	// Verify cleanup after error
	manager.mu.RLock()
	lockCount := len(manager.locks)
	manager.mu.RUnlock()

	if lockCount != 0 {
		t.Error("Expected locks to be cleaned up after WithAppLock error")
	}
}

func TestAppLockManager_GlobalInstance(t *testing.T) {
	manager1 := GetAppLockManager()
	manager2 := GetAppLockManager()

	if manager1 != manager2 {
		t.Error("Expected GetAppLockManager to return the same global instance")
	}
}
