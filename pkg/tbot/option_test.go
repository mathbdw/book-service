package tbot

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWithToken(t *testing.T){
	token := "test_token"

	b := &Bot{}
	opt := WithToken(token)
	opt(b)

	require.Equal(t, token, b.token)
}

func TestWithDebug(t *testing.T){
	debug := true

	b := &Bot{}
	opt := WithDebug(debug)
	opt(b)

	require.Equal(t, debug, b.debug)
}

func TestWithReadTimeout(t *testing.T){
	readTimeout := 5

	b := &Bot{}
	opt := WithReadTimeout(readTimeout)
	opt(b)

	require.Equal(t, readTimeout, b.readTimeout)
}