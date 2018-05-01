package started

import (
	"testing"
	"github.com/stretchr/testify/require"
)

func TestState(t *testing.T) {

	s := StateFactory()
	
	//Initial should be false
	require.False(t, s.HasStarted())
	
	//Start
	s.Start()
	require.True(t, s.HasStarted())
	
	//Stop again
	s.Stop()
	require.False(t, s.HasStarted())
	
}