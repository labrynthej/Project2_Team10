package Project2_Team10

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
)

type Instruction struct {
	typeOfInstruction string // instruction type "R", "I", etc..
	rawInstruction    string // raw data (we need to run this through a function that figures out the OPcode)
	lineValue         uint64 // linevalue = rawinstruction converted to uint64 so I could use mask and shift on it.
	programCnt        int    // program counter
	opcode            uint64 // once we know this we can figure out everything else
	op                string // what is it? ADD, SUB, LSL, etc...
	rd                uint8
	rn                uint8
	rm                uint8
	im                string
	rt                uint8
	address           uint8
	offset            string
	conditional       uint8
	field             uint64
	shamt             uint8
	op2               string
}

type Simulation struct {
	cycle         int
	lineValue     uint64 // linevalue = rawinstruction converted to uint64 so I could use mask and shift on it.
	programCnt    int    // program counter
	opcode        uint64 // once we know this we can figure out everything else
	op            string // what is it? ADD, SUB, LSL, etc...
	registerArray [24]int
	rd            uint8
	rn            uint8
	rm            uint8
	im            string
	rt            uint8
	address       uint8
	offset        string
	conditional   uint8
	field         uint64
	shamt         uint8
	op2           string
}

// global data slice
var dataSlice [][]int

func main() {
	//flag.String gets pointers to command line arguments
	cmdInFile := flag.String("i", "addtest1_bin.txt", "-i [input file path/name]")
	cmdOutFile := flag.String("o", "team10_out.txt", "-o [output file path/name]")
	flag.Parse() //flag.parse just makes things work

	inFile, _ := os.Open(*cmdInFile) //*cmdInFile because we need to dereference it

	//create a new array of instructions based on the data read from the inFile
	var instructionsArray []Instruction = readFile(inFile)
	initializeInstructions(instructionsArray) //initialize the instructions

	printResults(instructionsArray, *cmdOutFile+"_dis.txt")

	// begin simulation
	var simulationArray []Simulation
	simResults(instructionsArray, simulationArray, *cmdOutFile+"_sim.txt")

	fmt.Println("infile:", *cmdInFile)
	fmt.Println("outfile: ", *cmdOutFile)
}

// reads the file and loads each line into the rawInstruction part of the Instruction
func readFile(fileBeingRead io.Reader) (inputParsed []Instruction) {
	index := 0

	// ^^ begins reading from the fileBeingRead (io.Reader is like an iostream) ...
	// ... and prepares an array of Instructions that will be returned later
	scanner := bufio.NewScanner(fileBeingRead) // "scanner" is a bufio.NewScanner of the file being read. ...

	// ...By default, a "Scan" terminates at new line.
	for scanner.Scan() { // for each Scan (each line to read) in "scanner"
		newInstruction := Instruction{rawInstruction: scanner.Text()} // creates a new Instruction and assigns...
		// ...scanner.Text() to rawInstruction (scanner.Text() is the text of one line of the input file)
		inputParsed = append(inputParsed, newInstruction) // add the newInstruction (containing the raw data) to the...
		// ...array of instructions

		// set the current memory value then increment by 4
		inputParsed[index].programCnt = 96 + (4 * index)
		index++

	}
	fmt.Println(scanner.Text())
	if err := scanner.Err(); err != nil {
		fmt.Println(err)
	}

	return
}

// intialize all values of the struct array based on the line value
// mask bits using with bitwise-AND value to hex value of mask
// ie. "lineValue & 0xFF00" with a 32-bit mask 0xFF00 == 00000000000000001111111100000000
// shift bits using >> operation
func initializeInstructions(instArray []Instruction) {
	for i := 0; i < len(instArray); i++ {
		//the below converts 32 characters from a base 2 string to base 10 uint64

		lineValue, _ := strconv.ParseUint(instArray[i].rawInstruction, 2, 32)

		if lineValue > 335544320 || lineValue == 0 {
			// assign lineValue and 11 bit opcode for setting the instruction
			instArray[i].lineValue = lineValue
			instArray[i].opcode = lineValue >> 21

			setInstructionType(instArray, i)

			// set values for instruction type "R" | opcode | Rm | Shamt | Rn | Rd |
			if instArray[i].typeOfInstruction == "R" {
				instArray[i].rn = uint8((lineValue & 0x3E0) >> 5)
				instArray[i].rm = uint8((lineValue & 0x1F0000) >> 16)
				instArray[i].rd = uint8(lineValue & 0x1F)
				instArray[i].shamt = uint8((lineValue & 0xFC00) >> 11)
			}

			// set values for instruction type "D" | opcode | address | op2 | Rn | Rt |
			if instArray[i].typeOfInstruction == "D" {
				instArray[i].rn = uint8((lineValue & 0x3E0) >> 5)
				instArray[i].address = uint8((lineValue & 0x1FF000) >> 12)
				instArray[i].op2 = "#" + strconv.Itoa(int((lineValue&0xC00)>>10))
				instArray[i].rt = uint8(lineValue & 0x1F)
			}

			// set values for instruction type "I" | opcode | immediate | Rn | Rd |
			if instArray[i].typeOfInstruction == "I" {
				instArray[i].opcode = lineValue >> 22
				instArray[i].rn = uint8((lineValue & 0x3E0) >> 5)
				instArray[i].im = "#" + strconv.Itoa(int(signedVariable(lineValue&0x3FFC00>>10, 12)))
				instArray[i].rd = uint8(lineValue & 0x1F)
			}

			// set values for instruction type "B" | opcode | offset |
			if instArray[i].typeOfInstruction == "B" {
				instArray[i].opcode = lineValue >> 26
				instArray[i].offset = "#" + strconv.Itoa(int(signedVariable(lineValue&0x3FFFFFF, 26)))
			}

			// set values for instruction type "CB" (conditional B) | opcode | offset |
			if instArray[i].typeOfInstruction == "CB" {
				instArray[i].opcode = lineValue >> 24
				instArray[i].offset = "#" + strconv.Itoa(int(signedVariable(lineValue&0xFFFFE0>>5, 19)))
				instArray[i].conditional = uint8(lineValue & 0x1F)
			}

			// set values for instruction type "IM" | opcode | shift | field | Rd |
			if instArray[i].typeOfInstruction == "IM" {
				instArray[i].opcode = lineValue >> 23
				instArray[i].shamt = uint8(lineValue)
				instArray[i].field = lineValue & 0x300
				instArray[i].rd = uint8(lineValue & 0x1F)
			}

			if instArray[i].op == "BREAK" {
				break
			}
		}

	}
}

// check for signed variable to convert to negation using two's complement
func signedVariable(value uint64, length int) int32 {
	//var mask uint64 = 1 << (length - 1) // set mask to get sign
	var temp = value >> (length - 1)

	if temp == 1 { //
		value = value | (0xFFFFFFFF << length)
		//value = ^value + 1
	}
	return int32(value)
}

// function that determines the type of instruction and what it is
func setInstructionType(instrArray []Instruction, i int) {
	var decimalOPC uint64 = instrArray[i].opcode
	switch true { //switch defines a base case to test against.
	case ((decimalOPC >= 160) && (decimalOPC <= 191)): //if case == switch, do stuff in that one and ignore other cases
		instrArray[i].op = "B"
		instrArray[i].typeOfInstruction = "B"
	case (decimalOPC == 1104):
		instrArray[i].op = "AND"
		instrArray[i].typeOfInstruction = "R"
	case (decimalOPC == 1112):
		instrArray[i].op = "ADD"
		instrArray[i].typeOfInstruction = "R"
	case (decimalOPC >= 1160 && decimalOPC <= 1161):
		instrArray[i].op = "ADDI"
		instrArray[i].typeOfInstruction = "I"
	case (decimalOPC == 1360):
		instrArray[i].op = "ORR"
		instrArray[i].typeOfInstruction = "R"
	case (decimalOPC >= 1440 && decimalOPC <= 1447):
		instrArray[i].op = "CBZ"
		instrArray[i].typeOfInstruction = "CB"
	case (decimalOPC >= 1448 && decimalOPC <= 1455):
		instrArray[i].op = "CBNZ"
		instrArray[i].typeOfInstruction = "CB"
	case (decimalOPC == 1624):
		instrArray[i].op = "SUB"
		instrArray[i].typeOfInstruction = "R"
	case (decimalOPC >= 1672 && decimalOPC <= 1673):
		instrArray[i].op = "SUBI"
		instrArray[i].typeOfInstruction = "I"
	case (decimalOPC >= 1684 && decimalOPC <= 1687):
		instrArray[i].op = "MOVZ"
		instrArray[i].typeOfInstruction = "IM"
	case (decimalOPC >= 1940 && decimalOPC <= 1943):
		instrArray[i].op = "MOVK"
		instrArray[i].typeOfInstruction = "IM"
	case (decimalOPC == 1690):
		instrArray[i].op = "LSR"
		instrArray[i].typeOfInstruction = "R"
	case (decimalOPC == 1691):
		instrArray[i].op = "LSL"
		instrArray[i].typeOfInstruction = "R"
	case (decimalOPC == 1984):
		instrArray[i].op = "STUR"
		instrArray[i].typeOfInstruction = "D"
	case (decimalOPC == 1986):
		instrArray[i].op = "LDUR"
		instrArray[i].typeOfInstruction = "D"
	case (decimalOPC == 1692):
		instrArray[i].op = "ASR"
		instrArray[i].typeOfInstruction = "R"
	case (decimalOPC == 1872):
		instrArray[i].op = "EOR"
		instrArray[i].typeOfInstruction = "R"
	case decimalOPC == 0:
		instrArray[i].op = "NOP"
		instrArray[i].typeOfInstruction = "N/A"
	case decimalOPC == 2038: //check for break
		instrArray[i].op = "BREAK"
		instrArray[i].typeOfInstruction = "BREAK"
	default:
		break
	}
}

func printResults(instrArray []Instruction, fileName string) {

	file, fileErr := os.Create(fileName)
	if fileErr != nil {
		fmt.Println(fileErr)
	}
	i := 0
	count := 0
	for instrArray[i].typeOfInstruction != "BREAK" { // loop through each array of structs
		switch instrArray[i].typeOfInstruction {
		// print results for R Type instructions == opcode (11 bits), Rm (5 bits), Shamt (6 bits), Rn (5 bits), Rd (5 bits)
		case "R":
			// print separated binary opcode
			for j := 0; j < 32; j++ {
				if j == 11 || j == 16 || j == 22 || j == 27 { // print spaces to separate
					_, _ = file.WriteString(" ")
				}
				if j <= 10 { // print binary for opcode
					_, _ = file.WriteString(string(instrArray[i].rawInstruction[j]))
				} else if j >= 11 && j <= 15 { // print binary for Rm/R2
					_, _ = file.WriteString(string(instrArray[i].rawInstruction[j]))
				} else if j >= 16 && j <= 21 { // print binary for Shamt
					_, _ = file.WriteString(string(instrArray[i].rawInstruction[j]))
				} else if j >= 22 && j <= 26 { // print binary for Rn/R1
					_, _ = file.WriteString(string(instrArray[i].rawInstruction[j]))
				} else if j >= 27 && j <= 31 { // print binary for Rd/R3
					_, _ = file.WriteString(string(instrArray[i].rawInstruction[j]))
				}
			}
			_, _ = file.WriteString(" " + strconv.Itoa(instrArray[i].programCnt) + " " + instrArray[i].op + " R" +
				strconv.Itoa(int(instrArray[i].rd)) + ", R" + strconv.Itoa(int(instrArray[i].rn)) +
				", R" + strconv.Itoa(int(instrArray[i].rm))) // print pc, type, Rm, Shamt, Rn, Rd
		// print results for D type instruction == opcode (11 bits), address (9 bits), op2 (2 bits), Rn (5 bits), Rt (5 bits)
		case "D":
			// print seperated binary opcode
			for j := 0; j < 32; j++ {
				if j == 11 || j == 20 || j == 22 || j == 27 {
					_, _ = file.WriteString(" ")
				}
				if j <= 10 { // print binary for opcode
					_, _ = file.WriteString(string(instrArray[i].rawInstruction[j]))
				} else if j >= 11 && j <= 20 { // print binary for address
					_, _ = file.WriteString(string(instrArray[i].rawInstruction[j]))
				} else if j >= 21 && j <= 22 { // print binary for op2
					_, _ = file.WriteString(string(instrArray[i].rawInstruction[j]))
				} else if j >= 23 && j <= 27 { // print binary for Rn
					_, _ = file.WriteString(string(instrArray[i].rawInstruction[j]))
				} else if j >= 28 && j <= 31 { // print binary for Rt
					_, _ = file.WriteString(string(instrArray[i].rawInstruction[j]))
				}
			}
			_, _ = file.WriteString(" " + strconv.Itoa(instrArray[i].programCnt) + " " + instrArray[i].op + " R" +
				strconv.Itoa(int(instrArray[i].rt)) + ", [R" + strconv.Itoa(int(instrArray[i].rn)) + ", #" +
				strconv.Itoa(int(instrArray[i].address)) + "]") // print pc, type, Rt, Rn, address
		// print results for I type instruction == opcode (10 bits), immediate (12 bits), Rn (5 bits), Rd (5 bits)
		case "I":
			// print separated binary opcode
			for j := 0; j < 32; j++ {
				if j == 10 || j == 22 || j == 27 {
					_, _ = file.WriteString(" ")
				}
				if j <= 9 { // print binary for opcode
					_, _ = file.WriteString(string(instrArray[i].rawInstruction[j]))
				} else if j >= 10 && j <= 21 { // print binary for immediate
					_, _ = file.WriteString(string(instrArray[i].rawInstruction[j]))
				} else if j >= 22 && j <= 26 { // print binary for Rn
					_, _ = file.WriteString(string(instrArray[i].rawInstruction[j]))
				} else if j >= 27 && j <= 31 { // print binary for Rd
					_, _ = file.WriteString(string(instrArray[i].rawInstruction[j]))
				}
			}
			_, _ = file.WriteString(" " + strconv.Itoa(instrArray[i].programCnt) + " " + instrArray[i].op + " R" +
				strconv.Itoa(int(instrArray[i].rd)) + ", R" + strconv.Itoa(int(instrArray[i].rn)) +
				", " + instrArray[i].im) // print pc, type, Rd, rn, im
		// print results for B type instruction == opcode (6 bits), offset (26 bits)
		case "B":
			// print separated binary code
			for j := 0; j < 32; j++ {
				if j == 6 {
					_, _ = file.WriteString(" ")
				}
				if j <= 6 { // print binary for opcode
					_, _ = file.WriteString(string(instrArray[i].rawInstruction[j]))
				} else if j >= 7 && j <= 31 { // print binary for offset
					_, _ = file.WriteString(string(instrArray[i].rawInstruction[j]))
				}
			}
			_, _ = file.WriteString(" " + strconv.Itoa(instrArray[i].programCnt) + " " + instrArray[i].op + " " +
				instrArray[i].offset) // print pc, type, offset
		// print results for CB type instructions == opcode (8 bits), offset (19 bits), conditional (5 bits)
		case "CB":
			// print separated binary code
			for j := 0; j < 32; j++ {
				if j == 8 || j == 27 {
					_, _ = file.WriteString(" ")
				}
				if j <= 8 { // print binary for opcode
					_, _ = file.WriteString(string(instrArray[i].rawInstruction[j]))
				} else if j >= 8 && j <= 26 { // print binary for offset
					_, _ = file.WriteString(string(instrArray[i].rawInstruction[j]))
				} else if j >= 27 && j <= 31 { // print binary for conditional
					_, _ = file.WriteString(string(instrArray[i].rawInstruction[j]))
				}
			}
			_, _ = file.WriteString(" " + strconv.Itoa(instrArray[i].programCnt) + " " + instrArray[i].op + " R" +
				strconv.Itoa(int(instrArray[i].conditional)) + ", " + instrArray[i].offset)
		// print results for IM type instructions == opcode (9 bits), shift code (2 bits), field (16 bits), Rd (5 bits)
		case "IM":
			// print seperated binary code
			for j := 0; j < 32; j++ {
				if j == 9 || j == 11 || j == 27 {
					_, _ = file.WriteString(" ")
				}
				if j <= 9 { // print binary for opcode
					_, _ = file.WriteString(string(instrArray[i].rawInstruction[j]))
				} else if j >= 10 && j <= 11 { // print binary for shift code
					_, _ = file.WriteString(string(instrArray[i].rawInstruction[j]))
				} else if j >= 12 && j <= 27 { // print binary for field
					_, _ = file.WriteString(string(instrArray[i].rawInstruction[j]))
				} else if j >= 28 && j <= 31 {
					_, _ = file.WriteString(string(instrArray[i].rawInstruction[j]))
				}
			}
			_, _ = file.WriteString(" " + strconv.Itoa(instrArray[i].programCnt) + " " + instrArray[i].op + " R" +
				strconv.Itoa(int(instrArray[i].rd)) + ", " + strconv.Itoa(int(instrArray[i].field)) + ", LSL " +
				strconv.Itoa(int(instrArray[i].shamt)))
		case "N/A":
			_, _ = file.WriteString(instrArray[i].rawInstruction + " " +
				strconv.Itoa(instrArray[i].programCnt) + " " + "NOP")
		default:
			println("Invalid value on line " + strconv.Itoa(i))
		}

		_, _ = file.WriteString("\n")
		i++
	}
	_, _ = file.WriteString(instrArray[i].rawInstruction + " " + strconv.Itoa(instrArray[i].programCnt) + " BREAK\n")
	for i = i + 1; i < len(instrArray); i++ {
		lineValue, _ := strconv.ParseUint(instrArray[i].rawInstruction, 2, 32)
		count--
		_, _ = file.WriteString(instrArray[i].rawInstruction + " " + strconv.Itoa(instrArray[i].programCnt) +
			" " + strconv.Itoa(int(signedVariable(lineValue, 32))) + "\n")
	}
}

func simResults(instrArray []Instruction, simArray []Simulation, fileName string) {

	return
}
