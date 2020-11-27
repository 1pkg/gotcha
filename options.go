package gotcha

import "sync/atomic"

// units definition coppied from https://github.com/alecthomas/units

//nolint
const (
	Kibibyte int64 = 1024
	KiB            = Kibibyte
	Mebibyte       = Kibibyte * 1024
	MiB            = Mebibyte
	Gibibyte       = Mebibyte * 1024
	GiB            = Gibibyte
	Tebibyte       = Gibibyte * 1024
	TiB            = Tebibyte
	Pebibyte       = Tebibyte * 1024
	PiB            = Pebibyte
	Exbibyte       = Pebibyte * 1024
	EiB            = Exbibyte
)

//nolint
const (
	Kilobyte int64 = 1000
	KB             = Kilobyte
	Megabyte       = Kilobyte * 1000
	MB             = Megabyte
	Gigabyte       = Megabyte * 1000
	GB             = Gigabyte
	Terabyte       = Gigabyte * 1000
	TB             = Terabyte
	Petabyte       = Terabyte * 1000
	PB             = Petabyte
	Exabyte        = Petabyte * 1000
	EB             = Exabyte
)

//nolint
const (
	Kilo     int64 = 1000
	Mega           = Kilo * 1000
	Giga           = Mega * 1000
	Tera           = Giga * 1000
	Peta           = Tera * 1000
	Exa            = Peta * 1000
	Infinity       = -1
)

// ContextOpt defines gotcha context limit options
// that could be applied to gotchactx.
type ContextOpt func(*gotchactx)

// ContextWithLimitBytes defines allocation limit bytes gotcha context option.
func ContextWithLimitBytes(lbytes int64) ContextOpt {
	return func(ctx *gotchactx) {
		atomic.StoreInt64(&ctx.lbytes, lbytes)
	}
}

// ContextWithLimitObjects defines allocation limit objects gotcha context option.
func ContextWithLimitObjects(lobjects int64) ContextOpt {
	return func(ctx *gotchactx) {
		atomic.StoreInt64(&ctx.lobjects, lobjects)
	}
}

// ContextWithLimitCalls defines allocation limit calls gotcha context option.
func ContextWithLimitCalls(lcalls int64) ContextOpt {
	return func(ctx *gotchactx) {
		atomic.StoreInt64(&ctx.lcalls, lcalls)
	}
}
