	cpu 8008new             ; use "new" 8008 mnemonics
	radix 10                ; use base 10 for numbers
	org 0
	    
	ORA A
	JNZ L1
	MVI B,2
	HLT
L1:
	MVI B,3
	HLT
