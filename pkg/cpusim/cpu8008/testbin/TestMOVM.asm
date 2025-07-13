	cpu 8008new             ; use "new" 8008 mnemonics
	radix 10                ; use base 10 for numbers
	org 0
	    
	MVI	H,12H
	MVI L,34H
	MOV A,M
	ADI 7
	MVI H,12H
	MVI L,35H
	MOV M,A
	HLT
