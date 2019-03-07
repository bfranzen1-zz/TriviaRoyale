package indexes

import (
	"reflect"
	"strings"
	"testing"
)

//TODO: implement automated tests for your trie data structure

func TestAddandFind(t *testing.T) {
	cases := []struct {
		name     string
		key      []string
		value    []int64
		expected []int64
		find     string
	}{
		{
			"Add one value",
			[]string{"test"},
			[]int64{1},
			[]int64{1},
			"test",
		},
		{
			"No key",
			[]string{""},
			[]int64{1},
			[]int64{},
			"",
		},
		{
			"Add on same branch",
			[]string{"going", "go"},
			[]int64{1, 2},
			[]int64{2, 1},
			"go",
		},
		{
			"Add to different branches",
			[]string{"test", "go"},
			[]int64{2, 10},
			[]int64{2},
			"test",
		},
		{
			"Add same key",
			[]string{"go", "go", "test"},
			[]int64{10, 20, 30},
			[]int64{10, 20},
			"go",
		},
		{
			"Find doesn't exist",
			[]string{"go", "golang", "test"},
			[]int64{10, 20, 30},
			[]int64{},
			"golf",
		},
		{
			"Find short prefix",
			[]string{"go", "golang", "ghost", "gopher"},
			[]int64{1, 2, 4, 5},
			[]int64{4, 1, 2, 5},
			"g",
		},
	}
	for _, c := range cases {
		trie := NewTrie()
		for i, k := range c.key {
			trie.Add(k, c.value[i])
		}
		//log.Println(c.name + "\n")
		if res := trie.Find(c.find, len(c.expected)); !reflect.DeepEqual(res, c.expected) {
			//spew.Dump(trie)
			t.Errorf("%s: Got wrong values, expected %v\n got %v", c.name, c.expected, res)
		}
	}
}

func TestRemoveAndFind(t *testing.T) {
	cases := []struct {
		name      string
		key       []string
		value     []int64
		remove    []string
		removeVal []int64
	}{
		{
			"Add and remove one value",
			[]string{"test"},
			[]int64{1},
			[]string{"test"},
			[]int64{1},
		},
		{
			"Add and remove multiple",
			[]string{"go", "gog"},
			[]int64{1, 2},
			[]string{"go", "gog"},
			[]int64{1, 2},
		},
		{
			"Remove from different branch",
			[]string{"test", "go"},
			[]int64{1, 2},
			[]string{"test"},
			[]int64{1},
		},
		{
			"Remove non-existent node",
			[]string{"go"},
			[]int64{1},
			[]string{"test"},
			[]int64{10},
		},
	}

	for _, c := range cases {
		trie := NewTrie()
		for i, n := range c.key {
			trie.Add(n, c.value[i])
		}

		for i, n := range c.remove {
			trie.Remove(n, c.removeVal[i])
			if res := trie.Find(n, 5); len(res) != len(c.key)-len(c.remove) &&
				!checkString(n, c.key) {
				t.Errorf("\n%s: Expected no returned values when looking for %s, got values: %v", c.name, n, res)
			}
		}
	}
}

func TestLen(t *testing.T) {
	cases := []struct {
		name     string
		key      []string
		value    []int64
		remove   int
		expected int
	}{
		{
			"Remove one value",
			[]string{"test", "go"},
			[]int64{1, 2},
			1,
			1,
		},
		{
			"Remove all",
			[]string{"test", "go"},
			[]int64{1, 2},
			2,
			0,
		},
		{
			"Test after trim",
			[]string{"go", "gog"},
			[]int64{1, 2},
			1,
			1,
		},
		{
			"Test multiple removes",
			[]string{"test", "golang", "got", "out"},
			[]int64{1, 2, 3, 4},
			3,
			1,
		},
	}

	for _, c := range cases {
		trie := NewTrie()
		for i, n := range c.key {
			trie.Add(n, c.value[i])
		}
		var i = 0
		for i < c.remove {
			trie.Remove(c.key[i], c.value[i])
			i++
		}
		if trie.Len() != c.expected {
			t.Errorf("%s: Wrong size returned, got: %v, wanted: %v", c.name, trie.Len(), c.expected)
		}
	}
}

// checkString checks and returns whether the s1 string is a prefix for any of the others strings
// used for TestRemoveAndFind
func checkString(s1 string, others []string) bool {
	for _, s := range others {
		if strings.HasPrefix(s, s1) {
			return true
		}
	}
	return false
}
