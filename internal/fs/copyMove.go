package fs

import (
	"HelaList/internal/op"
	"context"
	"fmt"
)

// Copy 仿照Rename函数的参数形式，在同一存储空间内复制对象。
func copy(ctx context.Context, srcPath, dstPath string, lazyCache ...bool) error {
	srcStorage, srcActualPath, err := op.GetStorageAndActualPath(srcPath)
	if err != nil {
		return fmt.Errorf("failed to get source storage for %s: %w", srcPath, err)
	}
	dstStorage, dstActualPath, err := op.GetStorageAndActualPath(dstPath)
	if err != nil {
		return fmt.Errorf("failed to get destination storage for %s: %w", dstPath, err)
	}

	// 仅处理同一存储的情况
	if srcStorage.GetStorage() == dstStorage.GetStorage() {
		err = op.Copy(ctx, srcStorage, srcActualPath, dstActualPath, lazyCache...)
		if err != nil {
			return fmt.Errorf("failed to copy %s to %s: %w", srcPath, dstPath, err)
		}
		return nil
	}

	// 如果是跨存储，则返回错误，因为此函数不处理该情况
	return fmt.Errorf("cross-storage copy is not supported by this function")
}

// Move 仿照Rename函数的参数形式，在同一存储空间内移动对象。
func move(ctx context.Context, srcPath, dstPath string, lazyCache ...bool) error {
	srcStorage, srcActualPath, err := op.GetStorageAndActualPath(srcPath)
	if err != nil {
		return fmt.Errorf("failed to get source storage for %s: %w", srcPath, err)
	}
	dstStorage, dstActualPath, err := op.GetStorageAndActualPath(dstPath)
	if err != nil {
		return fmt.Errorf("failed to get destination storage for %s: %w", dstPath, err)
	}

	// 仅处理同一存储的情况
	if srcStorage.GetStorage() == dstStorage.GetStorage() {
		err = op.Move(ctx, srcStorage, srcActualPath, dstActualPath, lazyCache...)
		if err != nil {
			return fmt.Errorf("failed to move %s to %s: %w", srcPath, dstPath, err)
		}
		return nil
	}

	// 如果是跨存储，则返回错误，因为此函数不处理该情况
	return fmt.Errorf("cross-storage move is not supported by this function")
}
