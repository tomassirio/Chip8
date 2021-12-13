package chip8

import (
	"log"
	"math/rand"
	"time"
)

const FONTSET_SIZE = 80
const TWO_BYTES_SIZE = 16
const MEMORY_SIZE = 4096

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
	drawFlag	 bool
}

const GFX_SIZE = 64 * 32

//35 OPCODES
type Opcode uint16
type Memory [MEMORY_SIZE]byte
type Gfx [GFX_SIZE]byte //Video System
type DelayedTimer byte
type SoundTimer byte
type Stack [TWO_BYTES_SIZE]byte
type Key [TWO_BYTES_SIZE]byte
type Registers [TWO_BYTES_SIZE]byte

/*
0x000-0x1FF - Chip 8 interpreter (contains font set in emu)
0x050-0x0A0 - Used for the built in 4x5 pixel font set (0-F)
0x200-0xFFF - Program ROM and work RAM
*/

func (c *Chip8) EmulateCycle() {
	// Fetch Opcode
	c.fetch()
	// Decode Opcode
	c.decode()
	// Execute Opcode
	c.execute()

	// Update timers
}

func (c *Chip8) initialize() {
	// Initialize registers and memory once
	c.pc = 0x200 //Program counter starts at byte 512 (0x200)
	c.opcode = 0
	c.ir = 0
	c.sp = 0 // Reset current Opcode, Index Register & Stack Pointer

	// Clear display
	for i:= 0; i < GFX_SIZE; i++ {
		c.gfx[i] = 0
	}
	// Clear stack
	for i:= 0; i < TWO_BYTES_SIZE; i++ {
		c.stack[i] = 0
	}
	// Clear registers V0-VF
	for i:= 0; i < TWO_BYTES_SIZE; i++ {
		c.key[i] = 0
		c.V[i] = 0
	}
	// Clear memory
	for i:= 0; i < MEMORY_SIZE; i++ {
		c.memory[i] = 0
	}

	//Load Fontset
	for i := 0; i < FONTSET_SIZE; i++ {
		c.memory[i] = getChip8Fontset()[i]
	}

	//Reset Timers
	c.delayedTimer = 0
	c.soundTimer = 0

	c.drawFlag = true

	rand.Seed(time.Now().Unix())
}

func (c *Chip8) fetch() {
	opcode := c.memory[c.pc]<<8 | c.memory[c.pc+1]
	c.opcode = Opcode(opcode)
}

func (c *Chip8) decode() {
	vXPos := c.opcode&0xF00>>8
	vYPos := c.opcode&0x0F0>>4
	switch c.opcode & 0xF000 {
	case 0x0000:
		switch c.opcode & 0x000F {
		case 0x0000: // 0x00E0: Clears the screen
			for i := 0; i < GFX_SIZE; i++ {
				c.gfx[i] = 0x0
			}
			//c.drawFlag = true
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
		Vx := c.V[vXPos]
		kk := byte(c.opcode & 0x00FF)
		if Vx == kk {
			c.incrementProgramCounter()
		}
		c.incrementProgramCounter()
		break
	case 0x4000: // SNE Vx, byte. Skip next instruction if Vx != byte
		Vx := c.V[vXPos]
		kk := byte(c.opcode & 0x00FF)
		if Vx != kk {
			c.incrementProgramCounter()
		}
		c.incrementProgramCounter()
		break
	case 0x5000: // SE Vx, Vy. Skip next instruction if Vx = Vy
		Vx := c.V[vXPos]
		Vy := c.V[vYPos]
		if Vx == Vy {
			c.incrementProgramCounter()
		}
		c.incrementProgramCounter()
		break
	case 0x6000: // LD Vx, byte. Set Vx = kk
		c.V[vXPos] = byte(c.opcode & 0x00FF)
		c.incrementProgramCounter()
		break
	case 0x7000: // Add Vx, Byte. Set Vx = Vx + kk
		c.V[vXPos] = c.V[vXPos] + byte(c.opcode&0x00FF)
		c.incrementProgramCounter()
		break
	case 0x8000:
		switch c.opcode & 0x000F {
		case 0x0000: // LD Vx, Vy. Set Vx = Vy
			c.V[vXPos] = c.V[vYPos]
			c.incrementProgramCounter()
			break
		case 0x0001: // Or Vx, Vy. Set Vx = Vx OR Vy
			c.V[vXPos] |= c.V[vYPos]
			c.incrementProgramCounter()
			break
		case 0x0002: // AND Vx, Vy. Set Vx = Vx AND Vy
			c.V[vXPos] &= c.V[vYPos]
			c.incrementProgramCounter()
			break
		case 0x0003: // XOR Vx, Vy. Set Vx = VX XOR Vy
			c.V[vXPos] ^= c.V[vYPos]
			c.incrementProgramCounter()
			break
		case 0x0004: // ADD Vx, Vy. Set Vx = Vx + Vy, set VF = carry
			c.V[0xF] = 0
			if c.V[vXPos] > (0xFF - c.V[vYPos]) {
				c.V[0xF] = 1
			}
			c.V[vXPos] += c.V[(vYPos) >> 4]
			c.incrementProgramCounter()
			break
		case 0x0005: // SUB Vx, Vy. Set Vx = Vx - Vy, set VF = NOT borrow.
			c.V[0xF] = 0
			if c.V[vXPos] > (0xFF - c.V[vYPos]) {
				c.V[0xF] = 1
			}
			c.V[vXPos] -= c.V[vYPos]
			c.incrementProgramCounter()
			break
		case 0x0006: // SHR Vx {, Vy}. Set Vx = Vx SHR 1
			c.V[0xF] = 0
			if c.V[vXPos] & 0x1 == 1 {
				c.V[0xF] = 1
			}
			c.V[vXPos] >>= 1
			c.incrementProgramCounter()
			break
		case 0x0007: // SUBN Vx, Vy. Set Vx = Vy - Vx, set VF = NOT borrow.
			c.V[0xF] = 0
			if c.V[vYPos] > c.V[vXPos] {
				c.V[0xF] = 1
			}
			c.V[vXPos] = c.V[vYPos] - c.V[vXPos]
			c.incrementProgramCounter()
			break
		case 0x000E: // SHL Vx {, Vy}. Set Vx = Vx SHL 1.
			c.V[0xF] = 0
			if c.V[vYPos] >> 7 == 1 {
				c.V[0xF] = 1
			}
			c.V[vXPos] <<= 1
			c.incrementProgramCounter()
			break
		default:
			log.Panicf("Unknown code [0x8000]: 0x%X\n", c.opcode)
		}
	case 0x9000: // SNE Vx, Vy. Skip next instruction if Vx != Vy
		if c.V[vXPos] != c.V[vYPos] {
			c.incrementProgramCounter()
		}
		c.incrementProgramCounter()
		break
	case 0xA000: // LD I, addr. Set I = NNN
		c.ir = uint16(c.opcode & 0x0FFF)
		c.incrementProgramCounter()
		break
	case 0xB000: // JP V0, addr. Jump to location NNN + V0
		c.pc = uint16(c.opcode & 0x0FFF) + uint16(c.V[0])
		break
	case 0xC000: // RND Vx, byte. Set vx = random byte AND kk
		c.V[vXPos] = byte(rand.Intn(255)%0xFF & int(c.opcode&0x00FF))
		c.incrementProgramCounter()
		break
	case 0xD000: // TODO: Implement
	case 0xE000: // TODO: Implement
		switch c.opcode&0x00FF {
		case 0x09E: // TODO: Implement
		case 0x0A1: // TODO: Implement
		}
	case 0xF000: // TODO: Implement
		switch c.opcode&0x00FF {
		case 0x007: // TODO: Implement
		case 0x00A: // TODO: Implement
		case 0x018: // TODO: Implement
		case 0x01E: // TODO: Implement
		case 0x029: // TODO: Implement
		case 0x033: // TODO: Implement
		case 0x055: // TODO: Implement
		case 0x065: // TODO: Implement
		}
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

func getChip8Fontset() [FONTSET_SIZE]byte {
	return [FONTSET_SIZE]byte{
		0xF0, 0x90, 0x90, 0x90, 0xF0, //0
		0x20, 0x60, 0x20, 0x20, 0x70, //1
		0xF0, 0x10, 0xF0, 0x80, 0xF0, //2
		0xF0, 0x10, 0xF0, 0x10, 0xF0, //3
		0x90, 0x90, 0xF0, 0x10, 0x10, //4
		0xF0, 0x80, 0xF0, 0x10, 0xF0, //5
		0xF0, 0x80, 0xF0, 0x90, 0xF0, //6
		0xF0, 0x10, 0x20, 0x40, 0x40, //7
		0xF0, 0x90, 0xF0, 0x90, 0xF0, //8
		0xF0, 0x90, 0xF0, 0x10, 0xF0, //9
		0xF0, 0x90, 0xF0, 0x90, 0x90, //A
		0xE0, 0x90, 0xE0, 0x90, 0xE0, //B
		0xF0, 0x80, 0x80, 0x80, 0xF0, //C
		0xE0, 0x90, 0x90, 0x90, 0xE0, //D
		0xF0, 0x80, 0xF0, 0x80, 0xF0, //E
		0xF0, 0x80, 0xF0, 0x80, 0x80,  //F
	}
}