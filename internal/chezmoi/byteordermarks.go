package chezmoi

type byteOrderMark struct {
	name   string
	prefix []byte
}

// Byte order marks. See https://en.wikipedia.org/wiki/Byte_order_mark.
var byteOrderMarks = []byteOrderMark{
	{
		name:   "UTF-8",
		prefix: []byte{0xef, 0xbb, 0xbf},
	},
	{
		name:   "UTF-16 (BE)",
		prefix: []byte{0xfe, 0xff},
	},
	{
		name:   "UTF-16 (LE)",
		prefix: []byte{0xff, 0xfe},
	},
	{
		name:   "UTF-32 (BE)",
		prefix: []byte{0x00, 0x00, 0xfe, 0xff},
	},
	{
		name:   "UTF-32 (LE)",
		prefix: []byte{0xff, 0xfe, 0x00, 0x00},
	},
	{
		name:   "UTF-7",
		prefix: []byte{0x2b, 0x2f, 0x76},
	},
	{
		name:   "UTF-1",
		prefix: []byte{0xf7, 0x64, 0x4c},
	},
	{
		name:   "UTF-EBCDIC",
		prefix: []byte{0xdd, 0x73, 0x66, 0x73},
	},
	{
		name:   "SCSU",
		prefix: []byte{0x0e, 0xfe, 0xff},
	},
	{
		name:   "BOCU-1",
		prefix: []byte{0xfb, 0xee, 0x28},
	},
	{
		name:   "GB18030",
		prefix: []byte{0x84, 0x31, 0x95, 0x33},
	},
}
