package main

import (
	"encoding/hex"
	"fmt"
	"math"
)

const (
	nothingInstruction  = byte(0x00)
	loadRegInstruction  = byte(0x01)
	loadRamInstruction  = byte(0x02)
	writeRamInstruction = byte(0x03)
	moveInstruction     = byte(0x04)
	bitshiftInstruction = byte(0x05)
	invertInstruction   = byte(0x06)
	orInstruction       = byte(0x07)
	andInstruction      = byte(0x08)
	xorInstruction      = byte(0x09)
	addInstruction      = byte(0x0A)
	subInstruction      = byte(0x0B)
	compareInstruction  = byte(0x0C)
	skipInstruction     = byte(0x0D)
	jumpInstruction     = byte(0x0E)

	operation1Reg = 0
	operation2Reg = 1
	ramAddressReg = 2
	moveReg       = 3
	jumpReg       = 4
)

var (
	romSize int = int(math.Pow(2.0, 16.0))
)

type CPU struct {
	rom        []byte
	ram        []byte
	reg        []byte
	pointer    int16
	loadAsData bool
	selReg     byte
}

func new(hexROM string) (CPU, error) {
	data, err := hex.DecodeString(hexROM)
	if err != nil {
		return CPU{}, err
	}
	if len(data) > romSize {
		return CPU{}, fmt.Errorf("Provided rom with size of %d is bigger than max allowed %d", len(data), 2^16)
	}

	rom := make([]byte, romSize, romSize)
	copy(rom[:len(data)], data)

	fmt.Println(data)
	return CPU{
		rom:        rom,
		ram:        make([]byte, 256), // 2^8 Bytes of RAM
		reg:        make([]byte, 16),
		pointer:    0,
		loadAsData: false,
	}, nil
}

func (c *CPU) DoCycle() bool {
	// load instruction
	instruction := c.rom[c.pointer]
	c.pointer = c.pointer + 1

	// if previous instruction was load into register
	if c.loadAsData {
		c.reg[c.selReg] = instruction

		// cleanup
		c.loadAsData = false
		c.selReg = 0
		return true
	}

	// split up instruction and registers
	c.selReg = instruction & byte(0x0F)
	instruction = instruction >> 4

	switch instruction {
	case loadRegInstruction:
		c.loadAsData = true
	case loadRamInstruction:
		c.reg[c.selReg] = c.ram[c.reg[ramAddressReg]]
	case writeRamInstruction:
		c.ram[c.reg[ramAddressReg]] = c.reg[c.selReg]
	case moveInstruction:
		tmp := c.reg[moveReg]
		c.reg[moveReg] = c.reg[c.selReg]
		c.reg[c.selReg] = tmp
	case bitshiftInstruction:
		c.reg[c.selReg] = c.reg[c.selReg] << 1
	case invertInstruction:
		c.reg[c.selReg] = c.reg[c.selReg] ^ byte(0xFF)
	case orInstruction:
		c.reg[c.selReg] = c.reg[operation1Reg] | c.reg[operation2Reg]
	case andInstruction:
		c.reg[c.selReg] = c.reg[operation1Reg] & c.reg[operation2Reg]
	case xorInstruction:
		c.reg[c.selReg] = c.reg[operation1Reg] ^ c.reg[operation2Reg]
	case addInstruction:
		c.reg[c.selReg] = c.reg[operation1Reg] + c.reg[operation2Reg]
	case subInstruction:
		c.reg[c.selReg] = c.reg[operation1Reg] - c.reg[operation2Reg]
	case compareInstruction:
		if c.reg[operation1Reg] > c.reg[operation2Reg] {
			c.reg[c.selReg] = byte(0x01)
		}
	case skipInstruction:
		if c.reg[c.selReg]&byte(0x01) == byte(0x01) {
			c.pointer = c.pointer + 1
		}
	case jumpInstruction:
		jumpPoint := int16(c.reg[jumpReg]) << 8
		jumpPoint += int16(c.reg[c.selReg])
		c.pointer = jumpPoint
	default:
		return false
	}
	return true
}

func (c *CPU) getRegister(reg int) byte {
	if reg > len(c.reg) {
		return 0
	}
	return c.reg[reg]
}

func main() {
	fmt.Println("Start")
	rom := "10" + // load 0F in reg 0
		"FF" + // Number 1
		"11" + // load F0 in reg 1
		"F0" + // Number 2
		"C5" + // compare reg 0 with reg 1, store in reg 5
		"16" + // store end location of if statement in reg 6
		"0E" + // TODO
		"17" + // store location of if true in reg 7
		"0D" + // TODO
		"D5" + // skip if reg 0 bigger than reg 1
		"E7" + // jump if false
		"40" + // move reg 0 to move reg
		"E6" + // jump to end of if
		"41" + // move reg 1 to move reg
		"33" + // store in ram
		"00" // end

	fmt.Println(rom)

	cpu, err := new(rom)
	if err != nil {
		panic(err)
	}

	//run
	run := true
	for run && cpu.pointer < int16(100) {
		run = cpu.DoCycle()
	}

	fmt.Println("Done")
	fmt.Println(cpu.getRegister(moveReg))
}
