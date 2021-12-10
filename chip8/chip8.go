package chip8

import (
	"log"
)

const FONTSET_SIZE = 80

type Chip8 struct {
	V            Registers
	ir, pc, sp   uint16 //Index Register, Program Counter, Stack Pointer
	opcode       Opcode
	memory       Memory
	gfx          Gfx
	delayedTimer DelayedTimer
	soundTimer   SoundTimer
	stack        Stack
	key          Key
}

const GFX_SIZE = 64 * 32

//35 OPCODES
type Opcode uint16
type Memory [4096]byte
type Gfx [GFX_SIZE]byte //Video System
type DelayedTimer byte
type SoundTimer byte
type Stack [16]byte
type Key [16]byte
type Registers [16]byte

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
	c.sp = 0 // Reset current Opcode, Index Register & Stack Pointer

	// Clear display
	// Clear stack
	// Clear registers V0-VF
	// Clear memory

	//Load Fontset
	for i := 0; i < FONTSET_SIZE; i++ {
		//c.memory[i] = chip8Fontset[i]
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
	opcode := c.memory[c.pc]<<8 | c.memory[c.pc+1]
	c.opcode = Opcode(opcode)
}

func (c *Chip8) decode() {
	switch c.opcode & 0xF000 {
	case 0x0000:
		switch c.opcode & 0x000F {
		case 0x0000: // 0x00E0: Clears the screen
			for i := 0; i < GFX_SIZE; i++ {
				c.gfx[i] = 0x0
			}
			c.drawFlag = true
			c.incrementProgramCounter()
			break
		case 0x000E: // 0x00EE: Returns from subroutine
			c.sp--
			c.pc = uint16(c.stack[c.sp])
			c.incrementProgramCounter()
			break
		default:
			log.Panicf("Unknown Opcode [0x0000]: 0x%X\n", c.opcode)
		}
	case 0x1000: // JP addr. Jump to location nnn
		c.pc = uint16(c.opcode & 0x0FFF)
	case 0x2000: // CALL addr. Call Subroutine at nnn
		c.stack[c.sp] = byte(c.pc)
		c.sp++
		c.pc = uint16(c.opcode & 0x0FFF)
		break
	case 0x3000: // SE Vx, byte. Skip next instruction if Vx = kk
		Vx := c.V[c.opcode&0xF00>>8]
		kk := byte(c.opcode & 0x00FF)
		if Vx == kk {
			c.incrementProgramCounter()
		}
		c.incrementProgramCounter()
		break
	case 0x4000: // SNE Vx, byte. Skip next instruction if Vx != byte
		Vx := c.V[c.opcode&0xF00>>8]
		kk := byte(c.opcode & 0x00FF)
		if Vx != kk {
			c.incrementProgramCounter()
		}
		c.incrementProgramCounter()
		break
	case 0x5000: // SE Vx, Vy. Skip next instruction if Vx = Vy
		Vx := c.V[c.opcode&0xF00>>8]
		Vy := c.V[c.opcode&0x0F0>>8]
		if Vx == Vy {
			c.incrementProgramCounter()
		}
		c.incrementProgramCounter()
		break
	case 0x6000: // LD Vx, byte. Set Vx = kk
		c.V[c.opcode&0xF00>>8] = byte(c.opcode & 0x00FF)
		c.incrementProgramCounter()
		break
	case 0x7000: // Add Vx, Byte. Set Vx = Vx + kk
		c.V[c.opcode&0xF00>>8] = c.V[c.opcode&0xF00>>8] + byte(c.opcode&0x00FF)
		c.incrementProgramCounter()
		break
	case 0x8000:
		switch c.opcode & 0x000F {
		case 0x0000: // LD Vx, Vy. Set Vx = Vy
			c.V[c.opcode&0xF00>>8] = c.V[c.opcode&0x0F0>>4]
			c.incrementProgramCounter()
			break
		case 0x0001: // Or Vx, Vy. Set Vx = Vx OR Vy
			c.V[c.opcode&0xF00>>8] |= c.V[c.opcode&0x0F0>>4]
			c.incrementProgramCounter()
			break
		case 0x0002: // AND Vx, Vy. Set Vx = Vx AND Vy
			c.V[c.opcode&0xF00>>8] &= c.V[c.opcode&0x0F0>>4]
			c.incrementProgramCounter()
			break
		case 0x0003: // XOR Vx, Vy. Set Vx = VX XOR Vy
			c.V[c.opcode&0xF00>>8] ^= c.V[c.opcode&0x0F0>>4]
			c.incrementProgramCounter()
			break
		case 0x0004: // ADD Vx, Vy. Set Vx = Vx + Vy, set VF = carry
			if c.V[c.opcode&0xF00>>8] > (0xFF - c.V[c.opcode&0xF00>>8]) {
				c.V[0xF] = 1
			} else {
				c.V[0xF] = 0
			}
			c.V[c.opcode&0xF00>>8] += c.V[(c.opcode & 0x00F0) >> 4]
			c.incrementProgramCounter()
			break

		}
	case 0xA000:
		c.ir = uint16(c.opcode & 0x0FFF)
		c.incrementProgramCounter()
		break
	default:
		log.Panicf("Unknown Opcode: 0x%X\n", c.opcode)
	}

	if c.delayedTimer > 0 {
		c.delayedTimer--
	}

	if c.soundTimer > 0 {
		if c.soundTimer == 1 {
			log.Println("BEEP!")
			c.soundTimer--
		}
	}
}

func (c *Chip8) execute() {
	c.pc += 2
}

func (c *Chip8) incrementProgramCounter() {
	c.pc += 2
}
