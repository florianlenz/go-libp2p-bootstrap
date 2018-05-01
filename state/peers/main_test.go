package started

import (
	"testing"
	"github.com/stretchr/testify/require"
)

func TestState(t *testing.T)  {
	
	s := StateFactory()
	
	//Initial state should be 0
	require.Equal(t, 0, s.Amount())
	
	//Update the amount
	s.SetAmountOfPeers(4)
	require.Equal(t, 4, s.Amount())
	
}
