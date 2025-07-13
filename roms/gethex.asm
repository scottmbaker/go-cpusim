        cpu 8008new             ; use "new" 8008 mnemonics
        radix 10                ; use base 10 for numbers

        org 2000h               ; beginning of EPROM
        rst 1

        org 2008h               ; rst 1 jumps here
        jmp start

start:  call get_hex
        HLT

get_hex:    call getch          
            ani 01111111B           ; mask out most significant bit
            cpi 0DH
            jz get_hex3             ; jump if enter key
            cpi 1BH
            jz get_hex3             ; jump if escape key
            cpi 20H
            jz get_hex3             ; jump if space
            cpi '0'
            jc get_hex              ; try again if less than '0'
            cpi 'a'
            jc get_hex1             ; jump if already upper case...
            sui 20H                 ; else convert to upper case
get_hex1:   cpi 'G'
            jnc get_hex             ; try again if greater than 'F'
            cpi ':'
            jc get_hex2             ; continue if '0'-'9'
            cpi 'A'
            jc get_hex              ; try again if less than 'A'
            
get_hex2:   mov b,a                 ; save the character in B
            call putch              ; echo the character
            sub a                   ; clear the carry flag
            mov a,b                 ; restore the character
            ret                     ; return with carry cleared and character in a

get_hex3:   mov b,a
            mvi a,1
            rrc                     ; set carry flag
            mov a,b
            ret                     ; return with carry set and character in a  

;----------------------------------------------

getch:      mov e,b            ; save B

            mvi A, '1'          ; simulate key press

            ani 07FH           ; strip high bit, for H9, otherwise we'll echo it
            mov b,e
            ret

;-----------------------------------------------------

CPRINT:
putch:  mov     e,c
        mov     c,a
        call    sout
        mov     c,e
        ret


sout:   in      3
        ani     01h
        jz      sout
        mov     a,c
        out     12h
        ret


