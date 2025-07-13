	cpu 8008new             ; use "new" 8008 mnemonics
	radix 10                ; use base 10 for numbers
	org 0
	    
	MVI A, 12H
	MVI B, 03H
	ANA B
	MOV B,A
	MVI A, 34H
	MVI C, 01H
	ORA C
	MOV C,A
	MVI A, 56H
	MVI D, 33H
	XRA D
	MOV D,A
	MVI A, 78H
	CPI 78H
	HLT
