package chip8

import (
	"log"
)

const FONTSET_SIZE = 80

type Chip8 struct {
	v0,v1,v2,v3,
	v4,v5,v6,v7,
	v8,v9,vA,vB,
	vC,vD,vE,vF byte // Registers
	ir, pc, sp uint16 //Index Register, Program Counter, Stack Pointer
	opcode Opcode
	memory Memory
	gfx Gfx
	delayedTimer DelayedTimer
	soundTimer SoundTimer
	stack Stack
	key Key
}

//35 OPCODES
type Opcode uint16
type Memory [4096]byte
type Gfx [64 * 32]byte //Video System
type DelayedTimer byte
type SoundTimer byte
type Stack [16]byte
type Key [16]byte

/*
0x000-0x1FF - Chip 8 interpreter (contains font set in emu)
0x050-0x0A0 - Used for the built in 4x5 pixel font set (0-F)
0x200-0xFFF - Program ROM and work RAM
*/

func (c *Chip8) Initialize() {
	// Initialize registers and memory once
	c.pc = 0x200 //Program counter starts at byte 512 (0x200)
	c.opcode = 0
	c.ir = 0
	c.sp = 0  // Reset current Opcode, Index Register & Stack Pointer

	// Clear display
	// Clear stack
	// Clear registers V0-VF
	// Clear memory

	//Load Fontset
	for i:= 0; i < FONTSET_SIZE; i++ {
		c.memory[i] = chip8Fontset[i]
	}

	//Reset Timers
}

func (c *Chip8) EmulateCycle() {
	// Fetch Opcode
	c.fetch()
	// Decode Opcode
	c.decode()
	// Execute Opcode
	c.execute()

	// Update timers
}

func (c *Chip8) fetch() {
	opcode := c.memory[c.pc] << 8 | c.memory[c.pc + 1]
	c.opcode = Opcode(opcode)
}

func (c *Chip8) decode() {
	switch c.opcode & 0xF000 {
		case 0x0000:
			switch c.opcode & 0x000F {
				case 0x0000: // 0x00E0: Clears the screen
					break
				case 0x000E: // 0x00EE: Returns from subroutine
					break
			default:
				log.Panicf("Unknown Opcode [0x0000]: 0x%X\n", c.opcode)
			}
		case 0x2000:
			c.stack[c.sp] = byte(c.pc)
			c.sp ++
			c.pc = uint16(c.opcode & 0x0FFF)
			break
		case 0xA000:
			c.ir = uint16(c.opcode & 0x0FFF)
			c.pc +=2
			break
		default:
			log.Panicf("Unknown Opcode: 0x%X\n", c.opcode)
	}

	if c.delayedTimer > 0 {
		c.delayedTimer --
	}

	if c.soundTimer > 0 {
		if c.soundTimer == 1 {
			log.Println("BEEP!")
			c.soundTimer --
		}
	}
}

func (c *Chip8) execute() {
	c.pc += 2
}