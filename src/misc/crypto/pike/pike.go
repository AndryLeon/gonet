package pike

type ff_addikey struct {
	sd     uint32
	dis1   int32
	dis2   int32
	index  int32
	carry  int8
	buffer [64]uint32
}

type FF_CTX struct {
	sd      uint32
	index   int
	addikey [3]ff_addikey
	buffer  [4096]byte
}

const (
	GENIUS_NUMBER = 0x05027919
)

/*! 参见<<应用密码学>>中的线性反馈移位寄存器算法*/
func linearity(key *uint32) {
	*key = ((((*key >> 31) ^ (*key >> 6) ^ (*key >> 4) ^ (*key >> 2) ^ (*key >> 1) ^ *key) & 0x00000001) << 31) | (*key >> 1)
}

//----------------------------------------------- Create New Pike
func NewCtx(sd uint32) (ctx *FF_CTX) {
	ctx = &FF_CTX{}
	ctx.sd = sd ^ GENIUS_NUMBER

	ctx.addikey[0].sd = ctx.sd
	linearity(&ctx.addikey[0].sd)
	ctx.addikey[0].dis1 = 55
	ctx.addikey[0].dis2 = 24

	ctx.addikey[1].sd = ((ctx.sd & 0xAAAAAAAA) >> 1) | ((ctx.sd & 0x55555555) << 1)
	linearity(&ctx.addikey[1].sd)
	ctx.addikey[1].dis1 = 57
	ctx.addikey[1].dis2 = 7

	ctx.addikey[2].sd = ^(((ctx.sd & 0xF0F0F0F0) >> 4) | ((ctx.sd & 0x0F0F0F0F) << 4))
	linearity(&ctx.addikey[2].sd)
	ctx.addikey[2].dis1 = 58
	ctx.addikey[2].dis2 = 19

	for i := 0; i < 3; i++ {
		tmp := ctx.addikey[i].sd
		for j := 0; j < 64; j++ {
			for k := 0; k < 32; k++ {
				linearity(&tmp)
			}
			ctx.addikey[i].buffer[j] = tmp
		}
		ctx.addikey[i].carry = 0
		ctx.addikey[i].index = 63
	}

	ctx.index = 4096

	return
}

func addikey_next(addikey *ff_addikey) {
	tmp := addikey.index + 1
	addikey.index = tmp & 0x03F

	i1 := ((addikey.index | 0x40) - addikey.dis1) & 0x03F
	i2 := ((addikey.index | 0x40) - addikey.dis2) & 0x03F

	addikey.buffer[addikey.index] = addikey.buffer[i1] + addikey.buffer[i2]

	if (addikey.buffer[addikey.index] < addikey.buffer[i1]) || (addikey.buffer[addikey.index] < addikey.buffer[i2]) {
		addikey.carry = 1
	} else {
		addikey.carry = 0
	}
}

func _generate(ctx *FF_CTX) {
	for i := uint(0); i < 1024; i++ {
		carry := ctx.addikey[0].carry + ctx.addikey[1].carry + ctx.addikey[2].carry

		if carry == 0 || carry == 3 { /*!< 如果三个位相同(全0或全1),那么钟控所有的发生器*/
			addikey_next(&ctx.addikey[0])
			addikey_next(&ctx.addikey[1])
			addikey_next(&ctx.addikey[2])
		} else { /*!< 如果三个位不全相同,则钟控两个相同的发生器*/
			flag := int8(0)

			if carry == 2 {
				flag = 1
			}

			for j := 0; j < 3; j++ {
				if ctx.addikey[j].carry == flag {
					addikey_next(&ctx.addikey[j])
				}
			}
		}

		tmp := ctx.addikey[2].buffer[ctx.addikey[2].index] ^ ctx.addikey[1].buffer[ctx.addikey[1].index] ^ ctx.addikey[0].buffer[ctx.addikey[0].index]

		for j := uint(0); j < 4; j++ {
			mask := uint32(0xFF) << (j * 8)
			ctx.buffer[i*4+j] = byte(tmp & mask >> (j * 8))
		}
	}

	ctx.index = 0
}

//----------------------------------------------- Encode a given buffer
func Encode(ctx *FF_CTX, data []byte) {
	LEN := len(data)
	if LEN == 0 {
		return
	}

	idx := 0

	for {
		remnant := 4096 - ctx.index
		if remnant <= 0 {
			_generate(ctx)
			continue
		}

		if remnant > LEN {
			remnant = LEN
		}

		LEN -= remnant
		base := ctx.index

		for i := 0; i < remnant; i++ {
			data[idx] ^= ctx.buffer[base+idx]
			idx++
		}

		ctx.index += remnant

		if LEN <= 0 {
			break
		}
	}
}
