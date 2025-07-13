        cpu 8008new             ; use "new" 8008 mnemonics
        radix 10                ; use base 10 for numbers

        org 2000h               ; beginning of EPROM
        rst 1

        org 2008h               ; rst 1 jumps here
        jmp start

start:  mvi     a,12h
        mov     a,b
        mov     a,c
        mov     a,d
        mov     a,e
        mov     a,h
        mov     a,l
        mov     a,m

        mvi     b,34h
        mov     b,a
        mov     b,c
        mov     b,d
        mov     b,e
        mov     b,h
        mov     b,l

        mvi     m,56h
        mov     m,a
        mov     m,b
        mov     m,c
        mov     m,d
        mov     m,e
        mov     m,h
        mov     m,l

        hlt

	
