	cpu 8008new             ; use "new" 8008 mnemonics
	radix 10                ; use base 10 for numbers
	org 0
	    
	MVI D, 45H
	MVI E, 54H
	CALL L1
	HLT
L1:
	INR D
	CALL L2
	INR E
	RET
L2:
	INR D
	CALL L3
	INR E
	RET
L3:
	INR D
	CALL L4
	INR E
	RET
L4:
	INR D
	CALL L5
	INR E
	RET
L5:
	INR D
	CALL L6
	INR E
	RET
L6:
	INR D
	CALL L7
	INR E
	RET
L7:
	MOV B,D
	MOV C,E
	RET
