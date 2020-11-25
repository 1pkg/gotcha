package main

import "sync/atomic"

// units definition coppied from https://github.com/alecthomas/units

//nolint
const (
	Kibibyte uint64 = 1024
	KiB             = Kibibyte
	Mebibyte        = Kibibyte * 1024
	MiB             = Mebibyte
	Gibibyte        = Mebibyte * 1024
	GiB             = Gibibyte
	Tebibyte        = Gibibyte * 1024
	TiB             = Tebibyte
	Pebibyte        = Tebibyte * 1024
	PiB             = Pebibyte
	Exbibyte        = Pebibyte * 1024
	EiB             = Exbibyte
)

//nolint
const (
	Kilobyte uint64 = 1000
	KB              = Kilobyte
	Megabyte        = Kilobyte * 1000
	MB              = Megabyte
	Gigabyte        = Megabyte * 1000
	GB              = Gigabyte
	Terabyte        = Gigabyte * 1000
	TB              = Terabyte
	Petabyte        = Terabyte * 1000
	PB              = Petabyte
	Exabyte         = Petabyte * 1000
	EB              = Exabyte
)

//nolint
const (
	Kilo uint64 = 1000
	Mega        = Kilo * 1000
	Giga        = Mega * 1000
	Tera        = Giga * 1000
	Peta        = Tera * 1000
	Exa         = Peta * 1000
)

type ContextOpt func(*gotchactx)

func ContextWithLimitBytes(lbytes uint64) ContextOpt {
	return func(ctx *gotchactx) {
		atomic.StoreUint64(&ctx.lbytes, lbytes)
	}
}

func ContextWithLimitObjects(lobjects uint64) ContextOpt {
	return func(ctx *gotchactx) {
		atomic.StoreUint64(&ctx.lobjects, lobjects)
	}
}

func ContextWithLimitCalls(lcalls uint64) ContextOpt {
	return func(ctx *gotchactx) {
		atomic.StoreUint64(&ctx.lcalls, lcalls)
	}
}
