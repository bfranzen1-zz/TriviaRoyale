package indexes

import (
	"sort"
	"sync"
)

//PRO TIP: if you are having troubles and want to see
//what your trie structure looks like at various points,
//either use the debugger, or try this package:
//https://github.com/davecgh/go-spew

//Trie implements a trie data structure mapping strings to int64s
//that is safe for concurrent use.

type Node struct {
	Name     rune
	Vals     []int64
	Children map[rune]*Node // O(1) indexing
	Parent   *Node
	mx       sync.RWMutex // TODO: Move to NODE
}

type Trie struct {
	root *Node
	size int // constant time for len method
}

//NewTrie constructs a new Trie.
func NewTrie() *Trie {
	return &Trie{
		root: &Node{Vals: []int64{}, Children: map[rune]*Node{}},
		size: 0,
	}
}

//Len returns the number of entries in the trie.
func (t *Trie) Len() int {
	return t.size
}

//Add adds a key and value to the trie.
func (t *Trie) Add(key string, value int64) {
	if key == "" || value == 0 { // invalid parameters
		return
	}
	t.root.mx.Lock()
	t.size++
	curr := t.root
	for _, letter := range key {
		if curr.Children[letter] == nil {
			curr.Children[letter] = &Node{Name: letter, Vals: []int64{}, Children: map[rune]*Node{}, Parent: curr}
		}
		curr = curr.Children[letter]
	}
	curr.Vals = append(curr.Vals, value)
	t.root.mx.Unlock()
}

//Find finds `max` values matching `prefix`. If the trie
//is entirely empty, or the prefix is empty, or max == 0,
//or the prefix is not found, this returns a nil slice.
func (t *Trie) Find(prefix string, max int) []int64 {
	t.root.mx.RLock()
	defer t.root.mx.RUnlock()

	if (&Trie{}) == t || prefix == "" || max == 0 {
		return []int64{}
	}
	ret := []int64{}
	curr := getBranch(t.root, prefix)
	if curr != nil {
		ret = dfsNode(curr, max, make(map[*Node]bool))
	}
	return ret
}

//Remove removes a key/value pair from the trie
//and trims branches with no values.
func (t *Trie) Remove(key string, value int64) {
	t.root.mx.Lock()
	curr := getBranch(t.root, key) // get to key/value
	if curr != nil {               // make sure the node exists
		curr.Vals = removeVal(curr.Vals, value) // remove value
		t.size--
		if len(curr.Children) < 1 && len(curr.Vals) < 1 {
			par := curr.Parent
			delete(par.Children, curr.Name)
			for par.Parent != nil && len(par.Children) < 1 && len(par.Vals) < 1 { // trim branch
				curr = par
				par = par.Parent
				delete(par.Children, curr.Name)
			}
			if par.Name == 0 && len(par.Children) < 1 && len(par.Vals) < 1 {
				par = nil
			}
		}
	}
	t.root.mx.Unlock()
}

// removeVal takes in an array of int64 and a val to delete
// returns the slice of int64 without the val
func removeVal(arr []int64, val int64) []int64 {
	for i, v := range arr {
		if v == val { // if value found
			// swap last value in slice with value to delete
			arr[len(arr)-1], arr[i] = arr[i], arr[len(arr)-1]
			break // break from loop so we can return updated values
		}
	}
	// return array with last value (the value we want gone) deleted
	return arr[:len(arr)-1]
}

// dfsNodes takes in a starting node, max integer, and slice of found IDs and searches
// the children of the node for values adding them to the found slice and returning what was found
// will return x amount of values less than or equal to the max
func dfsNode(node *Node, max int, visited map[*Node]bool) []int64 {
	visited[node] = true
	found := []int64{}
	if node.Vals != nil {
		if len(node.Vals) < max-len(found) {
			found = node.Vals
		} else {
			found = node.Vals[:(max - len(found))]
			return found // end early because we found all we need
		}
	}
	keys := sortKeys(node.Children)
	for _, k := range keys {
		if !visited[node.Children[k]] {
			found = append(found, dfsNode(node.Children[k], max, visited)...)
		}
	}
	return found
}

// sortKeys takes in a map of runes to Nodes and sorts the map keys
// returning the sorted keys as a rune slice
func sortKeys(m map[rune]*Node) []rune {
	var keys []rune
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	return keys
}

// getBranch takes in a node to start from and a key string to search for
// and returns the branch of the trie that includes the key
func getBranch(node *Node, key string) *Node {
	for _, letter := range key {
		node = node.Children[letter]
		if node == nil { // visited null node
			break
		}
	}
	return node
}
