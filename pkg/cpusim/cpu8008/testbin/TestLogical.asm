	cpu 8008new             ; use "new" 8008 mnemonics
	radix 10                ; use base 10 for numbers
	org 0
	    
	MVI A, 12H
	ANI 03H
	MOV B,A
	MVI A, 34H
	ORI 01H
	MOV C,A
	MVI A, 56H
	XRI 33H
	MOV D,A
	MVI A, 78H
	CPI 78H
	HLT
