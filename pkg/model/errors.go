package model

import "errors"

var ErrDataCorrupted = errors.New("data corrupted")
var ErrFailedNotify = errors.New("failed notify")
var ErrServerDown = errors.New("server down")
