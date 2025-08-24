package fs

import (
	"HelaList/internal/model"
	"context"
)

func list(ctx context.Context, path string, args *ListArgs) ([]model.Obj, error) {
	files := make([]model.Obj, 0)

}
