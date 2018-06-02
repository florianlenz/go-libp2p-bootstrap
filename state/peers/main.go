package started

type State struct {
	read  chan *stateRead
	write chan *stateWrite
}

//Read operation on the state
type stateRead struct {
	resp chan int
}

//Write operation on the state
type stateWrite struct {
	amount int
}

func StateFactory() *State {

	readChan := make(chan *stateRead)
	writeChan := make(chan *stateWrite)

	s := State{
		read:  readChan,
		write: writeChan,
	}

	go func() {

		state := 0

		for {
			select {
			case read := <-s.read:
				read.resp <- state
			case write := <-s.write:
				state = write.amount
			}
		}

	}()

	return &s

}

//Get the amount of connected peer's
func (s *State) Amount() int {

	c := make(chan int)

	s.read <- &stateRead{c}

	return <-c
}

//set the amount of connected peer's
func (s *State) SetAmountOfPeers(amount int) {

	s.write <- &stateWrite{amount: amount}

}
