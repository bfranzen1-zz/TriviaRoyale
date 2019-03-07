package handlers

// TODO: Need queue logic here

/*
- when lobby owner starts game we get game-start message
- start looping over questions
			- publish question to queue
			- wait x seconds
			- read off queue and remove people with wrong answers
- after loop
	- save results
	- type game-end
		- results
	- remove lobby from context
*/
