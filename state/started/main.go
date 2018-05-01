package started

type State struct  {
	read chan *stateRead
	write chan *stateWrite
}

//Read operation on the state
type stateRead struct {
	resp chan bool
}

//Write operation on the state
type stateWrite struct {
	new bool
}


func StateFactory() *State {
	
	readChan := make(chan *stateRead)
	writeChan := make(chan *stateWrite)
	
	s := State{
		read: readChan,
		write: writeChan,
	}
	
	go func() {
		
		state := false
		
		for {
			select {
			case read := <- s.read:
				read.resp <- state
			case write := <- s.write:
				state = write.new
			}
		}
		
	}()
	
	return &s
	
}

//Check if bootstrap has started
func (s *State) HasStarted() bool {
	
	res := make(chan bool)
	
	s.read <- &stateRead{res}
	
	return <-res
}

//Start bootstrapping
func (s *State) Start()  {
	s.write <- &stateWrite{true}
}

//Stop bootstrapping
func (s *State) Stop()  {
	s.write <- &stateWrite{false}
}