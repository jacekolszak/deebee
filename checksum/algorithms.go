// (c) 2021 Jacek Olszak
// This code is licensed under MIT license (see LICENSE for details)

package checksum

import (
	"crypto/md5"
	"crypto/sha512"
	"hash"
	"hash/crc32"
	"hash/crc64"
	"hash/fnv"
)

var CRC64 = &algorithm{
	newSum: func() Sum {
		table := crc64.MakeTable(crc64.ISO)
		return &hashSum{
			Hash: crc64.New(table),
		}
	},
	name: algorithmName("crc64"),
}

var CRC32 = &algorithm{
	newSum: func() Sum {
		return &hashSum{
			Hash: crc32.New(crc32.IEEETable),
		}
	},
	name: algorithmName("crc32"),
}

var SHA512 = &algorithm{
	newSum: func() Sum {
		return &hashSum{
			Hash: sha512.New(),
		}
	},
	name: algorithmName("sha512"),
}

var MD5 = &algorithm{
	newSum: func() Sum {
		return &hashSum{
			Hash: md5.New(),
		}
	},
	name: algorithmName("md5"),
}

var FNV32 = &algorithm{
	newSum: func() Sum {
		return &hashSum{
			Hash: fnv.New32(),
		}
	},
	name: algorithmName("fnv32"),
}

var FNV32a = &algorithm{
	newSum: func() Sum {
		return &hashSum{
			Hash: fnv.New32a(),
		}
	},
	name: algorithmName("fnv32a"),
}

var FNV64 = &algorithm{
	newSum: func() Sum {
		return &hashSum{
			Hash: fnv.New64(),
		}
	},
	name: algorithmName("fnv64"),
}

var FNV64a = &algorithm{
	newSum: func() Sum {
		return &hashSum{
			Hash: fnv.New64a(),
		}
	},
	name: algorithmName("fnv64a"),
}

var FNV128 = &algorithm{
	newSum: func() Sum {
		return &hashSum{
			Hash: fnv.New128(),
		}
	},
	name: algorithmName("fnv128"),
}

var FNV128a = &algorithm{
	newSum: func() Sum {
		return &hashSum{
			Hash: fnv.New128a(),
		}
	},
	name: algorithmName("fnv128a"),
}

type algorithm struct {
	newSum func() Sum
	name   AlgorithmName
}

func (h *algorithm) Name() AlgorithmName {
	return h.name
}

func (h *algorithm) NewSum() Sum {
	return h.newSum()
}

type hashSum struct {
	hash.Hash
}

func (f *hashSum) Marshal() []byte {
	return f.Hash.Sum([]byte{})
}

func algorithmName(s string) AlgorithmName {
	name := AlgorithmName{}
	copy(name[:], s)
	return name
}

var algorithmsByName = map[AlgorithmName]*algorithm{
	MD5.Name():     MD5,
	CRC64.Name():   CRC64,
	CRC32.Name():   CRC32,
	SHA512.Name():  SHA512,
	FNV64a.Name():  FNV64a,
	FNV64.Name():   FNV64,
	FNV32.Name():   FNV32,
	FNV32a.Name():  FNV32a,
	FNV128.Name():  FNV128,
	FNV128a.Name(): FNV128a,
}
