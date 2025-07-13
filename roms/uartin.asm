        cpu 8008new             ; use "new" 8008 mnemonics
        radix 10                ; use base 10 for numbers

        org 2000h               ; beginning of EPROM
        rst 1

        org 2008h               ; rst 1 jumps here
        jmp start

start:  mvi     H, 22h
        mvi     L, 0h
        call    printstr

        call    sin
        mov     c,a
        call    sout

        call    crlf
        
        hlt

sout:   in      3
        ani     01h
        jz      sout
        mov     a,c
        out     12h
        ret

sin:    in      3
        ani     02h
        jz      sin
        in      2
        ret

crlf:   mvi     c,0dh
        call    sout
        mvi     c,0ah
        jmp     sout

printstr:   mov     A,M
            ora     A
            rz
            mov     C,A

            call    sout

            inr     L
            jmp     printstr

        org     2200h
msg     DB      "press any key and we will echo it",0Dh,0Ah,0
	
