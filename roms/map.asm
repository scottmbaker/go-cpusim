            cpu 8008new             ; use "new" 8008 mnemonics
            radix 10                ; use base 10 for numbers

            org 2000H                   ; start of EPROM
            rst 1

            org 2008H                   ; rst 1 jumps here
            jmp go_rom0

;----------------------------------------

MMAP0       equ 0CH
MMAP1       equ 0DH
MMAP2       equ 0EH
MMAP3       equ 0FH

PAGE0       equ 000H
PAGE1       equ 010H
PAGE2       equ 020H
PAGE3       equ 030H

RAM0        equ 80H             ; 0x80 enables RAM
RAM1        equ 81H
RAM2        equ 82H
RAM3        equ 83H
ROMOR       equ 00H

            org 2040H
go_rom0:    mvi a,00H
            jmp go_rom
go_rom1:    mvi a,01H
            jmp go_rom
go_rom2:    mvi a,02H
            jmp go_rom
go_rom3:    mvi a,03H
            jmp go_rom

            ;; go_rom
            ;; input
            ;;    A = rom number. Assumes each ROM consumes 2 pages (8KB total)
            ;; destroys
            ;;    B
go_rom:     mov b,a
            ral                 ; A = A * 2
            ani 0FEH
            ori ROMOR
            out MMAP2           ; page2 = (rom*2)
            adi 1H              
            out MMAP3           ; page3 = (rom*2)+1
            mvi a, RAM0
            out MMAP0
            mvi a, RAM1
            out MMAP1
            mov a,b
            in 1                ; enable mapper
            jmp rom_start

;---------------------------------

rom_start:  MVI     H, 21h
            MVI     L, 00h
            CALL    printstr

            MVI     B, 22h
            MVI     L, 00h
            MVI     C, 12h
            CALL    copystr

            MVI     H, 12h
            MVI     L, 00h
            CALL    printstr

            HLT

;--------------------------------
            
            ; print null-terminated string at HL
printstr:   mov     A,M
            ora     A
            rz
            mov     C,A

sout:       in      3
            ani     01h
            jz      sout
            mov     a,c
            out     12h

            inr     L
            jmp     printstr

;---------------------------------

            ; copy string from BL to CL
copystr:    MOV     H,B
            MOV     A,M
            ORA     A
            RZ
            MOV     H,C
            MOV     M,A
            INR     L
            JMP     copystr

;--------------------------------
            
            ORG     2100h
romstr:     DB      "string in ROM",0Dh,0Ah,0
        
            ORG     2200h
ramstr:     DB      "string in RAM",0Dh,0Ah,0



	
