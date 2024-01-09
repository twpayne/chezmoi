package git

import (
	"bufio"
	"bytes"
	"regexp"
	"strconv"
)

// A ParseError is a parse error.
type ParseError string

// An OrdinaryStatus is a status of a modified file.
type OrdinaryStatus struct {
	X    byte
	Y    byte
	Sub  string
	MH   int64
	MI   int64
	MW   int64
	HH   string
	HI   string
	Path string
}

// A RenamedOrCopiedStatus is a status of a renamed or copied file.
type RenamedOrCopiedStatus struct {
	X        byte
	Y        byte
	Sub      string
	MH       int64
	MI       int64
	MW       int64
	HH       string
	HI       string
	RC       byte
	Score    int64
	Path     string
	OrigPath string
}

// An UnmergedStatus is the status of an unmerged file.
type UnmergedStatus struct {
	X    byte
	Y    byte
	Sub  string
	M1   int64
	M2   int64
	M3   int64
	MW   int64
	H1   string
	H2   string
	H3   string
	Path string
}

// An UntrackedStatus is a status of an untracked file.
type UntrackedStatus struct {
	Path string
}

// An IgnoredStatus is a status of an ignored file.
type IgnoredStatus struct {
	Path string
}

// A Status is a status.
type Status struct {
	Ordinary        []OrdinaryStatus
	RenamedOrCopied []RenamedOrCopiedStatus
	Unmerged        []UnmergedStatus
	Untracked       []UntrackedStatus
	Ignored         []IgnoredStatus
}

var (
	statusPorcelainV2ZOrdinaryRx = regexp.MustCompile(`` +
		`^1 ` +
		`([!\.\?ACDMRU])([!\.\?ACDMRU]) ` +
		`(N\.\.\.|S[\.C][\.M][\.U]) ` +
		`([0-7]+) ` +
		`([0-7]+) ` +
		`([0-7]+) ` +
		`([0-9a-f]+) ` +
		`([0-9a-f]+) ` +
		`(.*)` +
		`$`,
	)
	statusPorcelainV2ZRenamedOrCopiedRx = regexp.MustCompile(`` +
		`^2 ` +
		`([!\.\?ACDMRU])([!\.\?ACDMRU]) ` +
		`(N\.\.\.|S[\.C][\.M][\.U]) ` +
		`([0-7]+) ` +
		`([0-7]+) ` +
		`([0-7]+) ` +
		`([0-9a-f]+) ` +
		`([0-9a-f]+) ` +
		`([CR])([0-9]+) ` +
		`(.*?)\t(.*)` +
		`$`,
	)
	statusPorcelainV2ZUnmergedRx = regexp.MustCompile(`` +
		`^u ` +
		`([!\.\?ACDMRU])([!\.\?ACDMRU]) ` +
		`(N\.\.\.|S[\.C][\.M][\.U]) ` +
		`([0-7]+) ` +
		`([0-7]+) ` +
		`([0-7]+) ` +
		`([0-7]+) ` +
		`([0-9a-f]+) ` +
		`([0-9a-f]+) ` +
		`([0-9a-f]+) ` +
		`(.*)` +
		`$`,
	)
	statusPorcelainV2ZUntrackedRx = regexp.MustCompile(`` +
		`^\? ` +
		`(.*)` +
		`$`,
	)
	statusPorcelainV2ZIgnoredRx = regexp.MustCompile(`` +
		`^! ` +
		`(.*)` +
		`$`,
	)
)

func (e ParseError) Error() string {
	return string(e) + ": parse error"
}

// ParseStatusPorcelainV2 parses the output of
//
//	git status --ignored --porcelain=v2
//
// See https://git-scm.com/docs/git-status.
func ParseStatusPorcelainV2(output []byte) (*Status, error) {
	var status Status
	s := bufio.NewScanner(bytes.NewReader(output))
	for s.Scan() {
		text := s.Text()
		switch text[0] {
		case '1':
			m := statusPorcelainV2ZOrdinaryRx.FindStringSubmatchIndex(text)
			if m == nil {
				return nil, ParseError(text)
			}
			mH, err := strconv.ParseInt(text[m[8]:m[9]], 8, 64)
			if err != nil {
				return nil, err
			}
			mI, err := strconv.ParseInt(text[m[10]:m[11]], 8, 64)
			if err != nil {
				return nil, err
			}
			mW, err := strconv.ParseInt(text[m[12]:m[13]], 8, 64)
			if err != nil {
				return nil, err
			}
			os := OrdinaryStatus{
				X:    text[m[2]],
				Y:    text[m[4]],
				Sub:  text[m[6]:m[7]],
				MH:   mH,
				MI:   mI,
				MW:   mW,
				HH:   text[m[14]:m[15]],
				HI:   text[m[16]:m[17]],
				Path: text[m[18]:m[19]],
			}
			status.Ordinary = append(status.Ordinary, os)
		case '2':
			m := statusPorcelainV2ZRenamedOrCopiedRx.FindStringSubmatchIndex(text)
			if m == nil {
				return nil, ParseError(text)
			}
			mH, err := strconv.ParseInt(text[m[8]:m[9]], 8, 64)
			if err != nil {
				return nil, err
			}
			mI, err := strconv.ParseInt(text[m[10]:m[11]], 8, 64)
			if err != nil {
				return nil, err
			}
			mW, err := strconv.ParseInt(text[m[12]:m[13]], 8, 64)
			if err != nil {
				return nil, err
			}
			score, err := strconv.ParseInt(text[m[20]:m[21]], 10, 64)
			if err != nil {
				return nil, err
			}
			rocs := RenamedOrCopiedStatus{
				X:        text[m[2]],
				Y:        text[m[4]],
				Sub:      text[m[6]:m[7]],
				MH:       mH,
				MI:       mI,
				MW:       mW,
				HH:       text[m[14]:m[15]],
				HI:       text[m[16]:m[17]],
				RC:       text[m[18]],
				Score:    score,
				Path:     text[m[22]:m[23]],
				OrigPath: text[m[24]:m[25]],
			}
			status.RenamedOrCopied = append(status.RenamedOrCopied, rocs)
		case 'u':
			m := statusPorcelainV2ZUnmergedRx.FindStringSubmatchIndex(text)
			if m == nil {
				return nil, ParseError(text)
			}
			m1, err := strconv.ParseInt(text[m[8]:m[9]], 8, 64)
			if err != nil {
				return nil, err
			}
			m2, err := strconv.ParseInt(text[m[10]:m[11]], 8, 64)
			if err != nil {
				return nil, err
			}
			m3, err := strconv.ParseInt(text[m[12]:m[13]], 8, 64)
			if err != nil {
				return nil, err
			}
			mW, err := strconv.ParseInt(text[m[14]:m[15]], 8, 64)
			if err != nil {
				return nil, err
			}
			us := UnmergedStatus{
				X:    text[m[2]],
				Y:    text[m[4]],
				Sub:  text[m[6]:m[7]],
				M1:   m1,
				M2:   m2,
				M3:   m3,
				MW:   mW,
				H1:   text[m[16]:m[17]],
				H2:   text[m[18]:m[19]],
				H3:   text[m[20]:m[21]],
				Path: text[m[22]:m[23]],
			}
			status.Unmerged = append(status.Unmerged, us)
		case '?':
			m := statusPorcelainV2ZUntrackedRx.FindStringSubmatchIndex(text)
			if m == nil {
				return nil, ParseError(text)
			}
			us := UntrackedStatus{
				Path: text[m[2]:m[3]],
			}
			status.Untracked = append(status.Untracked, us)
		case '!':
			m := statusPorcelainV2ZIgnoredRx.FindStringSubmatchIndex(text)
			if m == nil {
				return nil, ParseError(text)
			}
			us := IgnoredStatus{
				Path: text[m[2]:m[3]],
			}
			status.Ignored = append(status.Ignored, us)
		case '#':
			continue
		default:
			return nil, ParseError(text)
		}
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	return &status, nil
}

// Empty returns true if s is empty.
func (s *Status) Empty() bool {
	switch {
	case s == nil:
		return true
	case len(s.Ignored) != 0:
		return false
	case len(s.Ordinary) != 0:
		return false
	case len(s.RenamedOrCopied) != 0:
		return false
	case len(s.Unmerged) != 0:
		return false
	case len(s.Untracked) != 0:
		return false
	default:
		return true
	}
}
