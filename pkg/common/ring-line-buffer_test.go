package common

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRingLineBuffer_AddLine(t *testing.T) {
	instance := NewRingLineBuffer(5, 15)

	l1 := line(t, 1)
	l2 := line(t, 2)
	l3 := line(t, 3)
	l4 := line(t, 4)
	l5 := line(t, 5)
	l6 := line(t, 6)
	l7 := line(t, 7)
	l8 := line(t, 8)
	l9 := line(t, 9)
	l10 := line(t, 10)
	l11 := line(t, 11)
	l14 := line(t, 14)
	l15 := line(t, 15)
	l16 := line(t, 16)

	assert.NoError(t, instance.AddLine(l1))
	assert.NoError(t, instance.AddLine(l2))
	assert.NoError(t, instance.AddLine(l3))
	assert.NoError(t, instance.AddLine(l4))
	assert.NoError(t, instance.AddLine(l5))

	assert.Equal(t, [][]byte{l1, l2, l3, l4, l5}, instance.lines)

	assert.NoError(t, instance.AddLine(l6))
	assert.Equal(t, [][]byte{l6, l2, l3, l4, l5}, instance.lines)

	assert.NoError(t, instance.AddLine(l7))
	assert.Equal(t, [][]byte{l6, l7, l3, l4, l5}, instance.lines)

	assert.NoError(t, instance.AddLine(l8))
	assert.Equal(t, [][]byte{l6, l7, l8, l4, l5}, instance.lines)

	assert.NoError(t, instance.AddLine(l9))
	assert.Equal(t, [][]byte{l6, l7, l8, l9, l5}, instance.lines)

	assert.NoError(t, instance.AddLine(l10))
	assert.Equal(t, [][]byte{l6, l7, l8, l9, l10}, instance.lines)

	assert.NoError(t, instance.AddLine(l11))
	assert.Equal(t, [][]byte{l11, l7, l8, l9, l10}, instance.lines)

	assert.NoError(t, instance.AddLine(l14))
	assert.Equal(t, [][]byte{l11, l14, l8, l9, l10}, instance.lines)

	assert.NoError(t, instance.AddLine(l15))
	assert.Equal(t, [][]byte{l11, l14, l15, l9, l10}, instance.lines)

	assert.Equal(t, ErrLineTooLong, instance.AddLine(l16))
	assert.Equal(t, [][]byte{l11, l14, l15, l9, l10}, instance.lines)
}

func TestRingLineBuffer_Write(t *testing.T) {
	instance := NewRingLineBuffer(5, 15)
	write := func(v ...string) {
		t.Helper()
		toWrite := b(v...)
		n, err := instance.Write(toWrite)
		require.NoError(t, err)
		require.Equal(t, len(toWrite), n)
	}

	assert.Equal(t, [][]byte{nil, nil, nil, nil, nil}, instance.lines)

	write("abc")
	assert.Equal(t, [][]byte{nil, nil, nil, nil, nil}, instance.lines)
	assert.Equal(t, 3, instance.currentLength)
	assert.Equal(t, b("abc"), instance.current[:3])

	write("def")
	assert.Equal(t, [][]byte{nil, nil, nil, nil, nil}, instance.lines)
	assert.Equal(t, 6, instance.currentLength)
	assert.Equal(t, b("abcdef"), instance.current[:6])

	write("foo\n")
	assert.Equal(t, [][]byte{b("abcdeffoo"), nil, nil, nil, nil}, instance.lines)
	assert.Equal(t, 0, instance.currentLength)

	write("bar\nxyz")
	assert.Equal(t, [][]byte{b("abcdeffoo"), b("bar"), nil, nil, nil}, instance.lines)
	assert.Equal(t, 3, instance.currentLength)
	assert.Equal(t, b("xyz"), instance.current[:3])

	write("\n")
	assert.Equal(t, [][]byte{b("abcdeffoo"), b("bar"), b("xyz"), nil, nil}, instance.lines)
	assert.Equal(t, 0, instance.currentLength)

	write("\n")
	assert.Equal(t, [][]byte{b("abcdeffoo"), b("bar"), b("xyz"), b(""), nil}, instance.lines)
	assert.Equal(t, 0, instance.currentLength)

	_, actualErr := instance.Write(b("0123456789abcdefg\n"))
	require.Equal(t, ErrLineTooLong, actualErr)
	assert.Equal(t, [][]byte{b("abcdeffoo"), b("bar"), b("xyz"), b(""), nil}, instance.lines)
	assert.Equal(t, 0, instance.currentLength)

	instance.TruncateTooLongLines = true
	actualN, actualErr := instance.Write(b("0123456789abcdefg\n"))
	require.NoError(t, actualErr)
	require.Equal(t, 18, actualN)
	assert.Equal(t, [][]byte{b("fg"), b("bar"), b("xyz"), b(""), b("0123456789abcde")}, instance.lines)
	assert.Equal(t, 0, instance.currentLength)
}

func line(t testing.TB, len uint32) []byte {
	result := make([]byte, len)
	n, err := rand.New(rand.NewSource(666)).Read(result)
	require.NoError(t, err)
	require.Equal(t, int(len), n)
	return result
}

func b(v ...string) []byte {
	return []byte(strings.Join(v, ""))
}
