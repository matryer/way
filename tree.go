package way

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

const (
	static = iota
	wildcard
)

var errMultipleRegistrations = errors.New("multiple registrations")

type node struct {
	Key      string
	Value    map[string]http.Handler
	Children []*node
	Indices  []byte
	Wildcard *node
	Type     int
}

type methodBoundHandler struct {
	method  string
	handler http.Handler
}

func (mh *methodBoundHandler) asMap() map[string]http.Handler {
	if mh == nil {
		return nil
	}
	return map[string]http.Handler{mh.method: mh.handler}
}

func (n *node) Insert(key string, value *methodBoundHandler) error {
	idx := strings.IndexAny(key, ":*")
	if idx == -1 {
		_, err := n.insert(key, value)
		return err
	}

	pre := key[:idx]

	n, err := n.insert(pre, nil)
	if err != nil {
		return err
	}

	post := key[idx:]

	slashIdx := strings.IndexByte(post, '/')
	if slashIdx == -1 {
		_, err := n.insert(post, value)
		return err
	}

	n, err = n.insert(post[:slashIdx], nil)
	if err != nil {
		return err
	}

	return n.Insert(post[slashIdx:], value)
}

func (n *node) insert(key string, value *methodBoundHandler) (*node, error) {
	switch key[0] {
	case ':':
		if n.Wildcard != nil {
			if n.Wildcard.Key != key[1:] {
				return nil, fmt.Errorf("mismatched wild cards :%s and %s", n.Wildcard.Key, key)
			}
			if value != nil {
				if n.Wildcard.Value == nil {
					n.Wildcard.Value = make(map[string]http.Handler)
				}
				if n.Wildcard.Value[value.method] != nil {
					return nil, errMultipleRegistrations
				}
				n.Wildcard.Value[value.method] = value.handler

			}
			return n.Wildcard, nil
		}

		n.Wildcard = &node{
			Key:   key[1:],
			Value: value.asMap(),
			Type:  wildcard,
		}

		return n.Wildcard, nil
	}

	for i, childNode := range n.Children {
		if key == childNode.Key {
			if value != nil {
				if childNode.Value == nil {
					childNode.Value = make(map[string]http.Handler)
				}
				if childNode.Value[value.method] != nil {
					return nil, errMultipleRegistrations
				}
				childNode.Value[value.method] = value.handler
			}
			return childNode, nil
		}

		cp := commonPrefixLength(childNode.Key, key)
		if cp == 0 {
			continue
		}

		if cp == len(childNode.Key) {
			return childNode.insert(key[cp:], value)
		}

		childNode.Key = childNode.Key[cp:]

		if cp == len(key) {
			n.Children[i] = &node{
				Key:      key,
				Children: []*node{childNode},
				Indices:  []byte{childNode.Key[0]},
				Value:    value.asMap(),
			}
			return n.Children[i], nil
		}

		targetNode := &node{
			Key:      key[cp:],
			Children: []*node{},
			Value:    value.asMap(),
		}

		n.Children[i] = &node{
			Key:      key[:cp],
			Children: []*node{childNode, targetNode},
			Indices:  []byte{childNode.Key[0], targetNode.Key[0]},
		}

		return targetNode, nil
	}

	targetNode := &node{
		Key:      key,
		Children: []*node{},
		Value:    value.asMap(),
	}

	n.Children = append(n.Children, targetNode)
	n.Indices = append(n.Indices, targetNode.Key[0])

	return targetNode, nil
}

func (n *node) Lookup(path string, params *[]param) (result map[string]http.Handler) {
	var fallback map[string]http.Handler
	defer func() {
		if result == nil {
			result = fallback
		}
	}()

	var wildcardbackup *node

Walk:
	for {
		switch n.Type {
		case static:
			if !strings.HasPrefix(path, n.Key) {
				if wildcardbackup != nil {
					n = wildcardbackup
					continue Walk
				}
				if n.Value != nil && path+"/" == n.Key {
					return n.Value
				}
				return nil
			}
			path = path[len(n.Key):]
			if path == "" {
				return n.Value
			}
			if n.IsSubdirNode() {
				fallback = n.Value
			}
		case wildcard:
			if idx := strings.IndexByte(path, '/'); idx == -1 {
				*params = append(*params, param{
					key:   n.Key,
					value: path,
				})
				return n.Value
			} else {
				*params = append(*params, param{
					key:   n.Key,
					value: path[:idx],
				})
				path = path[idx:]
			}

		}

		if path == "/" && n.Value != nil {
			fallback = n.Value
		}

		wildcardbackup = n.Wildcard

		targetIndice := path[0]
		for i, c := range n.Indices {
			if c == targetIndice {
				n = n.Children[i]
				continue Walk
			}
		}

		if n.Wildcard != nil {
			n = n.Wildcard
			continue Walk
		}

		return nil
	}
}

func (node *node) IsSubdirNode() bool {
	return node != nil && node.Value != nil && strings.HasSuffix(node.Key, "/")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func commonPrefixLength(a, b string) (i int) {
	for ; i < min(len(a), len(b)); i++ {
		if a[i] != b[i] {
			break
		}
	}
	return
}
