	cpu 8008new             ; use "new" 8008 mnemonics
	radix 10                ; use base 10 for numbers
	org 0
	    
	MVI	A,1
	RLC
	MOV	B,A
	RRC
	MOV C,A
	MVI A,80h
	RLC
	MOV D,A
	RRC
	MOV E,A
	HLT
