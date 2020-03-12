package lflag

import (
	"sync"
	. "testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// these are used in the cli and env tests
var testParams = []Param{
	Param{ParamType: ParamTypeString, Name: "foo"},
	Param{ParamType: ParamTypeString, Name: "bar", Usage: "wut"},
	Param{ParamType: ParamTypeString, Name: "baz", Usage: "wut", Default: "wat"},
	Param{ParamType: ParamTypeBool, Name: "flag1"},
	Param{ParamType: ParamTypeBool, Name: "flag2"},
	Param{ParamType: ParamTypeString, Name: "foo-bar"},
}

type jstruct struct {
	Foo string `json:"foo"`
	Bar int    `json:"bar"`
}

func TestParse(t *T) {
	s := String("str", "default", "Some string")
	i := Int("int", 5, "Some int")
	b := Bool("bool", false, "Some bool")
	bf := Bool("bool-false", true, "Some bool")
	d := Duration("dur", 10*time.Second, "Some duration")
	var j jstruct
	JSON(&j, "json", jstruct{Foo: "foo", Bar: 5}, "Some json")

	Parse(SourceStub{
		"str":        "hello",
		"int":        "8",
		"bool":       "true",
		"bool-false": "false",
		"dur":        "5m",
		"json":       `{"foo":"FOO","bar":10}`,
	})

	assert.Equal(t, "hello", *s)
	assert.Equal(t, 8, *i)
	assert.True(t, *b)
	assert.False(t, *bf)
	assert.Equal(t, 5*time.Minute, *d)
	assert.Equal(t, jstruct{Foo: "FOO", Bar: 10}, j)
}

func TestParseDefault(t *T) {
	s := String("str", "default", "Some string")
	i := Int("int", 5, "Some int")
	b := Bool("bool", true, "Some bool")
	d := Duration("dur", 10*time.Second, "Some duration")
	var j jstruct
	JSON(&j, "json", jstruct{Foo: "foo", Bar: 5}, "Some json")

	Parse(SourceStub{})

	assert.Equal(t, "default", *s)
	assert.Equal(t, 5, *i)
	assert.True(t, *b)
	assert.Equal(t, 10*time.Second, *d)
	assert.Equal(t, jstruct{Foo: "foo", Bar: 5}, j)
}

func TestParseRequired(t *T) {
	s := RequiredString("str", "Some string")
	i := RequiredInt("int", "Some int")
	b := RequiredBool("bool", "Some bool")
	d := RequiredDuration("dur", "Some duration")
	var j jstruct
	RequiredJSON(&j, "json", "Some json")

	Parse(SourceStub{
		"str":  "hello",
		"int":  "8",
		"bool": "true",
		"dur":  "5h",
		"json": `{"foo":"FOO","bar":10}`,
	})

	assert.Equal(t, "hello", *s)
	assert.Equal(t, 8, *i)
	assert.True(t, *b)
	assert.Equal(t, 5*time.Hour, *d)
	assert.Equal(t, jstruct{Foo: "FOO", Bar: 10}, j)
}

func TestDo(t *T) {
	Reset()

	// we shouldn't need a lock but we do because the race detector is sensitive
	var i int
	var il sync.Mutex
	Do(func() {
		il.Lock()
		assert.Equal(t, 0, i)
		i++
		il.Unlock()
		Do(func() {
			il.Lock()
			assert.Equal(t, 2, i)
			i++
			il.Unlock()
		})
	})

	Do(func() {
		il.Lock()
		assert.Equal(t, 1, i)
		i++
		il.Unlock()
		Do(func() {
			il.Lock()
			assert.Equal(t, 3, i)
			i++
			il.Unlock()
			Do(func() {
				il.Lock()
				assert.Equal(t, 4, i)
				i++
				il.Unlock()
			})
		})
	})

	Parse(SourceStub{})

	ch := make(chan bool)
	Do(func() {
		il.Lock()
		assert.Equal(t, 5, i)
		i++
		il.Unlock()
		Do(func() {
			il.Lock()
			assert.Equal(t, 6, i)
			i++
			il.Unlock()
			close(ch)
		})
	})
	<-ch
	assert.Equal(t, 7, i)
}
