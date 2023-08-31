package bdd

import "context"

type RequireFunc func(ctx context.Context) (context.Context, error)

type RequireFuncs map[string]RequireFunc

const TagRequire = "@require="
