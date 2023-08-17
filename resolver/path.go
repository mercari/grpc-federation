package resolver

import (
	"fmt"
	"strconv"
)

// Path represents path used to refer to the value.
type Path interface {
	Selectors() []string
	Type(from *Type) (*Type, error)
	addPath(Path)
}

// PathType the type of path.
type PathType string

const (
	// UnknownPathType unknown path type.
	UnknownPathType PathType = "unknown"

	// MessageArgumentPathType a path to refer to message arguments starting with $.
	MessageArgumentPathType PathType = "message_argument"

	// NameReferencePathType this is the path to refer to the value defined in the message option starting with the name.
	NameReferencePathType PathType = "name_reference"
)

type PathBuilder struct {
	buf  []byte
	root *PathRoot
	path Path
}

type PathBase struct {
	Child Path
}

func (b *PathBase) addPath(path Path) {
	b.Child = path
}

type PathRoot struct {
	*PathBase
}

func (r *PathRoot) Type(from *Type) (*Type, error) {
	if r.Child == nil {
		return from, nil
	}
	return r.Child.Type(from)
}

func (r *PathRoot) Selectors() []string {
	base := []string{"$"}
	if r.Child != nil {
		return append(base, r.Child.Selectors()...)
	}
	return base
}

type PathSelector struct {
	*PathBase
	Selector string
}

func (s *PathSelector) Selectors() []string {
	base := []string{s.Selector}
	if s.Child != nil {
		return append(base, s.Child.Selectors()...)
	}
	return base
}

func (s *PathSelector) Type(from *Type) (*Type, error) {
	if from.Ref == nil {
		return nil, fmt.Errorf(
			`"%s" path selector must refer to a message type. but it is not a message type`,
			s.Selector,
		)
	}
	var foundType *Type
	for _, field := range from.Ref.Fields {
		if field.Name == s.Selector {
			foundType = field.Type
			break
		}
	}
	if foundType == nil {
		return nil, fmt.Errorf(`"%s" field does not exist in "%s" message`, s.Selector, from.Ref.Name)
	}
	if s.Child == nil {
		return foundType, nil
	}
	return s.Child.Type(foundType)
}

type PathIndex struct {
	*PathBase
	Selector int
}

func (i *PathIndex) Selectors() []string {
	base := []string{fmt.Sprintf("[%d]", i.Selector)}
	if i.Child != nil {
		return append(base, i.Child.Selectors()...)
	}
	return base
}

func (i *PathIndex) Type(from *Type) (*Type, error) {
	if !from.Repeated {
		return nil, fmt.Errorf(`path index selector cannot be used in this context. the specified type is not a repeated type`)
	}
	typ := &Type{
		Type:     from.Type,
		Repeated: from.Repeated,
		Ref:      from.Ref,
	}
	if i.Child == nil {
		return typ, nil
	}
	return i.Child.Type(typ)
}

type PathIndexAll struct {
	*PathBase
}

func (i *PathIndexAll) Selectors() []string {
	base := []string{"[*]"}
	if i.Child != nil {
		return append(base, i.Child.Selectors()...)
	}
	return base
}

func (i *PathIndexAll) Type(from *Type) (*Type, error) {
	if !from.Repeated {
		return nil, fmt.Errorf(`path index all selector cannot be used in this context. the specified type is not a repeated type`)
	}
	typ := &Type{
		Type:     from.Type,
		Repeated: from.Repeated,
		Ref:      from.Ref,
	}
	if i.Child == nil {
		return typ, nil
	}
	return i.Child.Type(typ)
}

func NewPathBuilder(path string) *PathBuilder {
	return &PathBuilder{buf: []byte(path)}
}

func (b *PathBuilder) Build() (Path, error) {
	if len(b.buf) == 0 {
		return nil, nil
	}

	b.root = &PathRoot{PathBase: &PathBase{}}
	b.path = b.root

	if b.buf[0] != '$' {
		buf := b.buf
		cursor, err := b.buildSelector(buf)
		if err != nil {
			return nil, err
		}
		if len(buf) > cursor {
			return nil, fmt.Errorf("invalid path format. remain unparsed characters %q", buf[cursor:])
		}
		return b.root.Child, nil
	}

	if len(b.buf) == 1 {
		return b.root, nil
	}
	buf := b.buf[1:]
	cursor, err := b.buildPath(buf)
	if err != nil {
		return nil, err
	}
	if len(buf) > cursor {
		return nil, fmt.Errorf("invalid path format. remain unparsed characters %q", buf[cursor:])
	}
	return b.root, nil
}

func (b *PathBuilder) buildPath(buf []byte) (int, error) {
	switch buf[0] {
	case '.':
		if len(buf) == 1 {
			return 0, fmt.Errorf("invalid path format. the path ends with dot character")
		}
		cursor, err := b.buildSelector(buf[1:])
		if err != nil {
			return 0, err
		}
		return cursor + 1, nil
	case '[':
		if len(buf) == 1 {
			return 0, fmt.Errorf("invalid path format: the path ends with left bracket character")
		}
		cursor, err := b.buildIndex(buf[1:])
		if err != nil {
			return 0, err
		}
		return cursor + 1, nil
	default:
		return 0, fmt.Errorf("invalid path format. expect dot or left bracket character. but found %c character", buf[0])
	}
}

func (b *PathBuilder) buildSelector(buf []byte) (int, error) {
	switch buf[0] {
	case '.', '[', ']', '$', '*':
		return 0, fmt.Errorf(`invalid path format. "%c" cannot be used after dot character`, buf[0])
	}
	for cursor := 0; cursor < len(buf); cursor++ {
		switch buf[cursor] {
		case '$', '*', ']':
			return 0, fmt.Errorf(`invalid path format. "%c" cannot be used in field selecting context`, buf[cursor])
		case '.':
			if cursor+1 >= len(buf) {
				return 0, fmt.Errorf("invalid path format. the path ends with dot character")
			}
			selector := buf[:cursor]
			b.addSelectorNode(string(selector))
			offset, err := b.buildSelector(buf[cursor+1:])
			if err != nil {
				return 0, err
			}
			return cursor + 1 + offset, nil
		case '[':
			if cursor+1 >= len(buf) {
				return 0, fmt.Errorf("invalid path format. the path ends with left bracket character")
			}
			selector := buf[:cursor]
			b.addSelectorNode(string(selector))
			offset, err := b.buildIndex(buf[cursor+1:])
			if err != nil {
				return 0, err
			}
			return cursor + 1 + offset, nil
		case ' ':
			return 0, fmt.Errorf(`invalid path format. cannot include empty characters in the path`)
		}
	}
	b.addSelectorNode(string(buf))
	return len(buf), nil
}

func (b *PathBuilder) buildIndex(buf []byte) (int, error) {
	switch buf[0] {
	case '.', '[', ']', '$':
		return 0, fmt.Errorf(`invalid path format. "%c" cannot be used after left bracket`, buf[0])
	case '*':
		if len(buf) == 1 {
			return 0, fmt.Errorf("invalid path format. the path ends with star character")
		}
		if buf[1] != ']' {
			return 0, fmt.Errorf(`invalid path format. expect right bracket character for index all selector. but specified "%c"`, buf[1])
		}
		b.addIndexAllNode()
		offset := len("*]")
		if len(buf) > 2 {
			cursor, err := b.buildPath(buf[2:])
			if err != nil {
				return 0, err
			}
			return cursor + offset, nil
		}
		return offset, nil
	case ' ':
		return 0, fmt.Errorf(`invalid path format. cannot include empty characters in the path`)
	}

	for cursor := 0; cursor < len(buf); cursor++ {
		switch buf[cursor] {
		case ']':
			index, err := strconv.ParseInt(string(buf[:cursor]), 10, 64)
			if err != nil {
				return 0, fmt.Errorf(`invalid path format. "%q" cannot be used as index path selector`, buf[:cursor])
			}
			b.addIndexNode(int(index))
			return b.buildNextCharIfExists(buf, cursor+1)
		}
	}
	return 0, fmt.Errorf("invalid path format. cannot find right bracket character in index path context")
}

func (b *PathBuilder) buildNextCharIfExists(buf []byte, cursor int) (int, error) {
	if len(buf) > cursor {
		offset, err := b.buildPath(buf[cursor:])
		if err != nil {
			return 0, err
		}
		return cursor + 1 + offset, nil
	}
	return cursor, nil
}

func (b *PathBuilder) addIndexAllNode() {
	p := &PathIndexAll{
		PathBase: &PathBase{},
	}
	b.path.addPath(p)
	b.path = p
}

func (b *PathBuilder) addSelectorNode(name string) {
	p := &PathSelector{
		PathBase: &PathBase{},
		Selector: name,
	}
	b.path.addPath(p)
	b.path = p
}

func (b *PathBuilder) addIndexNode(idx int) {
	p := &PathIndex{
		PathBase: &PathBase{},
		Selector: idx,
	}
	b.path.addPath(p)
	b.path = p
}
