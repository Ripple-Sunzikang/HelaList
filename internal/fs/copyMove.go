package fs

import (
	"HelaList/internal/op"
	"context"
	"fmt"
)

// 在同一存储空间内复制对象。
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

	// todo: 跨网盘复制文件
	return fmt.Errorf("cross-storage copy is not supported by this function")
}

// 在同一存储空间内移动对象。
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

	// todo: 跨网盘移动文件
	return fmt.Errorf("cross-storage move is not supported by this function")
}
