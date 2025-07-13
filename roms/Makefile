ASL=/usr/local/bin/asl
P2BIN=/usr/local/bin/p2bin
P2HEX=/usr/local/bin/p2hex

all: ops.rom uart.rom uartin.rom map.rom

ops.rom: ops.asm
	$(ASL) -cpu 8008 -L ops.asm -o ops.p
	$(P2BIN) ops.p ops.rom
	rm -f ops.p 

uart.rom: uart.asm
	$(ASL) -cpu 8008 -L uart.asm -o uart.p
	$(P2BIN) uart.p uart.rom
	rm -f uart.p

uartin.rom: uartin.asm
	$(ASL) -cpu 8008 -L uartin.asm -o uartin.p
	$(P2BIN) uartin.p uartin.rom
	rm -f uartin.p

map.rom: map.asm
	$(ASL) -cpu 8008 -L map.asm -o map.p
	$(P2BIN) map.p map.rom
	rm -f map.p

gethex.rom: gethex.asm
	$(ASL) -cpu 8008 -L gethex.asm -o gethex.p
	$(P2BIN) gethex.p gethex.rom
	rm -f gethex.p
