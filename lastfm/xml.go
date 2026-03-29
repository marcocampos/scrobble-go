package lastfm

import (
	"encoding/xml"
	"strings"
)

// xmlNode is a generic XML element that captures all attributes, character
// data, and child elements, enabling DOM-like traversal of API responses.
type xmlNode struct {
	XMLName xml.Name
	Attrs   []xml.Attr `xml:",any,attr"`
	Content string     `xml:",chardata"`
	Nodes   []xmlNode  `xml:",any"`
}

// attr returns the value of the named attribute, or "" if not found.
func (n *xmlNode) attr(name string) string {
	for _, a := range n.Attrs {
		if a.Name.Local == name {
			return a.Value
		}
	}
	return ""
}

// text returns the trimmed character data of the node.
func (n *xmlNode) text() string {
	return strings.TrimSpace(n.Content)
}

// find returns the first descendant element whose local tag name matches,
// or nil if none is found. Search is depth-first.
func (n *xmlNode) find(tag string) *xmlNode {
	for i := range n.Nodes {
		if n.Nodes[i].XMLName.Local == tag {
			return &n.Nodes[i]
		}
		if found := n.Nodes[i].find(tag); found != nil {
			return found
		}
	}
	return nil
}

// findAll returns all descendant elements whose local tag name matches.
// Search is depth-first; order matches document order.
func (n *xmlNode) findAll(tag string) []*xmlNode {
	var result []*xmlNode
	for i := range n.Nodes {
		if n.Nodes[i].XMLName.Local == tag {
			result = append(result, &n.Nodes[i])
		}
		result = append(result, n.Nodes[i].findAll(tag)...)
	}
	return result
}

// extract returns the trimmed text of the nth descendant element with the
// given tag name. index defaults to 0 (the first match).
// Returns "" if no match at that index exists.
func extract(n *xmlNode, tag string, index ...int) string {
	idx := 0
	if len(index) > 0 {
		idx = index[0]
	}
	nodes := n.findAll(tag)
	if idx >= len(nodes) {
		return ""
	}
	return nodes[idx].text()
}

// extractAll returns the trimmed text of every descendant element with the
// given tag name, in document order.
func extractAll(n *xmlNode, tag string) []string {
	nodes := n.findAll(tag)
	out := make([]string, len(nodes))
	for i, node := range nodes {
		out[i] = node.text()
	}
	return out
}

// parseXMLResponse parses a raw XML string into an xmlNode tree.
func parseXMLResponse(data string) (*xmlNode, error) {
	var root xmlNode
	if err := xml.Unmarshal([]byte(data), &root); err != nil {
		return nil, err
	}
	return &root, nil
}
