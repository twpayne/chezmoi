package chezmoi

// A contentsFunc is a function that returns the contents of a file or an error.
// It is typically used for lazy evaluation of a file's contents.
type contentsFunc func() ([]byte, error)

// A lazyContents evaluates its contents lazily.
type lazyContents struct {
	contentsFunc   contentsFunc
	contents       []byte
	contentsErr    error
	contentsSHA256 []byte
}

// A linknameFunc is a function that returns the target of a symlink or an
// error. It is typically used for lazy evaluation of a symlink's target.
type linknameFunc func() (string, error)

// A lazyLinkname evaluates its linkname lazily.
type lazyLinkname struct {
	linknameFunc   linknameFunc
	linkname       string
	linknameErr    error
	linknameSHA256 []byte
}

// newLazyContents returns a new lazyContents with contents.
func newLazyContents(contents []byte) *lazyContents {
	return &lazyContents{
		contents: contents,
	}
}

// newLazyContentsFunc returns a new lazyContents with contentsFunc.
func newLazyContentsFunc(contentsFunc func() ([]byte, error)) *lazyContents {
	return &lazyContents{
		contentsFunc: contentsFunc,
	}
}

// Contents returns lc's contents.
func (lc *lazyContents) Contents() ([]byte, error) {
	if lc == nil {
		return nil, nil
	}
	if lc.contentsFunc != nil {
		lc.contents, lc.contentsErr = lc.contentsFunc()
		lc.contentsFunc = nil
		if lc.contentsErr == nil {
			lc.contentsSHA256 = SHA256Sum(lc.contents)
		}
	}
	return lc.contents, lc.contentsErr
}

// ContentsSHA256 returns the SHA256 sum of lc's contents.
func (lc *lazyContents) ContentsSHA256() ([]byte, error) {
	if lc == nil {
		return SHA256Sum(nil), nil
	}
	if lc.contentsSHA256 == nil {
		contents, err := lc.Contents()
		if err != nil {
			return nil, err
		}
		lc.contentsSHA256 = SHA256Sum(contents)
	}
	return lc.contentsSHA256, nil
}

// newLazyLinkname returns a new lazyLinkname with linkname.
func newLazyLinkname(linkname string) *lazyLinkname {
	return &lazyLinkname{
		linkname: linkname,
	}
}

// newLazyLinknameFunc returns a new lazyLinkname with linknameFunc.
func newLazyLinknameFunc(linknameFunc func() (string, error)) *lazyLinkname {
	return &lazyLinkname{
		linknameFunc: linknameFunc,
	}
}

// Linkname returns s's linkname.
func (ll *lazyLinkname) Linkname() (string, error) {
	if ll == nil {
		return "", nil
	}
	if ll.linknameFunc != nil {
		ll.linkname, ll.linknameErr = ll.linknameFunc()
		ll.linknameFunc = nil
	}
	return ll.linkname, ll.linknameErr
}

// LinknameSHA256 returns the SHA256 sum of ll's linkname.
func (ll *lazyLinkname) LinknameSHA256() ([]byte, error) {
	if ll == nil {
		return SHA256Sum(nil), nil
	}
	if ll.linknameSHA256 == nil {
		linkname, err := ll.Linkname()
		if err != nil {
			return nil, err
		}
		ll.linknameSHA256 = SHA256Sum([]byte(linkname))
	}
	return ll.linknameSHA256, nil
}
