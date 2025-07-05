package chezmoigit

import (
	"regexp"
	"strconv"
	"strings"
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
	for line := range strings.Lines(string(output)) {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		switch line[0] {
		case '1':
			m := statusPorcelainV2ZOrdinaryRx.FindStringSubmatchIndex(line)
			if m == nil {
				return nil, ParseError(line)
			}
			mH, err := strconv.ParseInt(line[m[8]:m[9]], 8, 64)
			if err != nil {
				return nil, err
			}
			mI, err := strconv.ParseInt(line[m[10]:m[11]], 8, 64)
			if err != nil {
				return nil, err
			}
			mW, err := strconv.ParseInt(line[m[12]:m[13]], 8, 64)
			if err != nil {
				return nil, err
			}
			os := OrdinaryStatus{
				X:    line[m[2]],
				Y:    line[m[4]],
				Sub:  line[m[6]:m[7]],
				MH:   mH,
				MI:   mI,
				MW:   mW,
				HH:   line[m[14]:m[15]],
				HI:   line[m[16]:m[17]],
				Path: line[m[18]:m[19]],
			}
			status.Ordinary = append(status.Ordinary, os)
		case '2':
			m := statusPorcelainV2ZRenamedOrCopiedRx.FindStringSubmatchIndex(line)
			if m == nil {
				return nil, ParseError(line)
			}
			mH, err := strconv.ParseInt(line[m[8]:m[9]], 8, 64)
			if err != nil {
				return nil, err
			}
			mI, err := strconv.ParseInt(line[m[10]:m[11]], 8, 64)
			if err != nil {
				return nil, err
			}
			mW, err := strconv.ParseInt(line[m[12]:m[13]], 8, 64)
			if err != nil {
				return nil, err
			}
			score, err := strconv.ParseInt(line[m[20]:m[21]], 10, 64)
			if err != nil {
				return nil, err
			}
			rocs := RenamedOrCopiedStatus{
				X:        line[m[2]],
				Y:        line[m[4]],
				Sub:      line[m[6]:m[7]],
				MH:       mH,
				MI:       mI,
				MW:       mW,
				HH:       line[m[14]:m[15]],
				HI:       line[m[16]:m[17]],
				RC:       line[m[18]],
				Score:    score,
				Path:     line[m[22]:m[23]],
				OrigPath: line[m[24]:m[25]],
			}
			status.RenamedOrCopied = append(status.RenamedOrCopied, rocs)
		case 'u':
			m := statusPorcelainV2ZUnmergedRx.FindStringSubmatchIndex(line)
			if m == nil {
				return nil, ParseError(line)
			}
			m1, err := strconv.ParseInt(line[m[8]:m[9]], 8, 64)
			if err != nil {
				return nil, err
			}
			m2, err := strconv.ParseInt(line[m[10]:m[11]], 8, 64)
			if err != nil {
				return nil, err
			}
			m3, err := strconv.ParseInt(line[m[12]:m[13]], 8, 64)
			if err != nil {
				return nil, err
			}
			mW, err := strconv.ParseInt(line[m[14]:m[15]], 8, 64)
			if err != nil {
				return nil, err
			}
			us := UnmergedStatus{
				X:    line[m[2]],
				Y:    line[m[4]],
				Sub:  line[m[6]:m[7]],
				M1:   m1,
				M2:   m2,
				M3:   m3,
				MW:   mW,
				H1:   line[m[16]:m[17]],
				H2:   line[m[18]:m[19]],
				H3:   line[m[20]:m[21]],
				Path: line[m[22]:m[23]],
			}
			status.Unmerged = append(status.Unmerged, us)
		case '?':
			m := statusPorcelainV2ZUntrackedRx.FindStringSubmatchIndex(line)
			if m == nil {
				return nil, ParseError(line)
			}
			us := UntrackedStatus{
				Path: line[m[2]:m[3]],
			}
			status.Untracked = append(status.Untracked, us)
		case '!':
			m := statusPorcelainV2ZIgnoredRx.FindStringSubmatchIndex(line)
			if m == nil {
				return nil, ParseError(line)
			}
			us := IgnoredStatus{
				Path: line[m[2]:m[3]],
			}
			status.Ignored = append(status.Ignored, us)
		case '#':
			continue
		default:
			return nil, ParseError(line)
		}
	}
	return &status, nil
}

// IsEmpty returns true if s is empty.
func (s *Status) IsEmpty() bool {
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
