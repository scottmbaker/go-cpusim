	cpu 8008new             ; use "new" 8008 mnemonics
	radix 10                ; use base 10 for numbers
	org 0
	    
	JMP L1
	ORG 8h
	MVI	B,2
	RET
	ORG 10h
	MVI C,3
	RET
	ORG 18h
	MVI D,4
	RET
	ORG 20h
	MVI E,5
	RET
	ORG 28h
	MVI H,6
	RET
	ORG 30h
	MVI L,7
	RET
	ORG 38h
	MVI A,8
	RET
L1:
	RST 1
	RST 2
	RST 3
	RST 4
	RST 5
	RST 6
	RST 7
	HLT
