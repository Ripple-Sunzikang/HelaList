package fs

import (
	"HelaList/internal/model"
	"HelaList/internal/op"
	"context"
	"fmt"

	"github.com/pkg/errors"
)

func putDirectly(ctx context.Context, dstDirPath string, file model.FileStreamer, lazyCache ...bool) error {
	storage, dstDirActualPath, err := op.GetStorageAndActualPath(dstDirPath)
	if err != nil {
		_ = file.Close()
		return errors.WithMessage(err, "failed get storage")
	}
	if storage.Config().NoUpload {
		_ = file.Close()
		return errors.WithStack(fmt.Errorf("UploadNotSupported"))
	}
	return op.Put(ctx, storage, dstDirActualPath, file, nil, lazyCache...)
}
