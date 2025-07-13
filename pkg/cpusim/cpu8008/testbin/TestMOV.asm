	cpu 8008new             ; use "new" 8008 mnemonics
	radix 10                ; use base 10 for numbers
	org 0
	    
	MVI A, 12H
	MOV B,A
	INR B
	MOV C,B
	INR C
	MOV D,C
	INR D
	MOV E,D
	INR E
	MOV H,E
	INR H
	MOV L,H
	INR L
	HLT
