        cpu 8008new             ; use "new" 8008 mnemonics
        radix 10                ; use base 10 for numbers

        org 2000h               ; beginning of EPROM
        rst 1

        org 2008h               ; rst 1 jumps here
        jmp start

start:  mvi     c, 'S'
        call    sout
        mvi     c, 'C'
        call    sout
        mvi     c, 'O'
        call    sout
        mvi     c, 'T'
        call    sout
        mvi     c, 'T'
        call    sout
        mvi     c, 0Dh
        call    sout
        mvi     c, 0Ah
        call    sout

        ; print the alphabet
        mvi     b, 01Ah
        mvi     c, 'A'
loop:   call    sout
        inr     c
        dcr     b
        jnz     loop

        call    crlf

        hlt

sout:   in      3
        ani     01h
        jz      sout
        mov     a,c
        out     12h
        ret

crlf:   mvi     c,0dh
        call    sout
        mvi     c,0ah
        jmp     sout
	
