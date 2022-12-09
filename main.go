/*
	roman numerial terminal app
	- take an arabic or roman numeral string
		> determine which, and if valid
	- convert to the other numeral type R -> A or A -> R
	- take flag for subtractive or additive output
	- valid range of arabic numbers 1 - 4000 (MMMM)
*/

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var usage = `Roman numeral converter (Arabic to Roman & Roman to Arabic)
Sample Usage:
	$ roman 1965
	$ roman MCMLXV
	$ roman -a=true 1965
	$ roman MDCCCCLXV
	$ roman -sim=true -r=true -s 100 -e 250 -o=true -p my_values.txt -app=true

Options:
	a	<bool>	  default=false		do not use subtractive notation
	sim	<bool>	  default=false		simple output, only the converted value
	r	<bool>	  default=false		produce a range of output from start to end (inclusive)
	s	<int>	  default=1		the start value of the range
	e	<int>	  default=1		the end value of the range
	o	<bool>	  defalt=false		write output to local file
	p	<string>  default="out.txt"	the filename to produce
	app	<bool>	  default=false		write in append mode
`

var (
	// global command line flags
	addF        = flag.Bool("a", false, "use addative roman numerals")
	simpleOutF  = flag.Bool("sim", false, "only output the converted value")
	rangeF      = flag.Bool("r", false, "produce a range of values")
	startF      = flag.Int("s", 1, "starting arabic value in range")
	endF        = flag.Int("e", 1, "starting arabic value in range")
	writeFileF  = flag.Bool("o", false, "produce an output file with output")
	outpathF    = flag.String("p", "out.txt", "relative path of the output file")
	appendFileF = flag.Bool("app", false, "append file write versus truncate")
)

type NumType uint8

const (
	Arabic NumType = iota
	Roman
	UnDef
)

var aTorMap = map[uint16]string{
	1000: "M",
	900:  "CM",
	500:  "D",
	400:  "CD",
	100:  "C",
	90:   "XC",
	50:   "L",
	40:   "XL",
	10:   "X",
	9:    "IX",
	5:    "V",
	4:    "IV",
	1:    "I",
}

var rToaMap = map[string]uint16{
	"M":  1000,
	"CM": 900,
	"D":  500,
	"CD": 400,
	"C":  100,
	"XC": 90,
	"L":  50,
	"XL": 40,
	"X":  10,
	"IX": 9,
	"V":  5,
	"IV": 4,
	"I":  1,
}

func main() {
	flag.Parse()

	val := strings.ToUpper(flag.Arg(0))
	if val == "" && !*rangeF {
		fmt.Print(usage)
		return
	}

	// range case
	if *rangeF {
		outVals := genRange()
		if *writeFileF {
			log.Printf("writing Arabic range %d to %d to file", *startF, *endF)
			writeToFile(outVals)
			return
		} else {
			// write to terminal
			for _, v := range outVals {
				fmt.Println(v)
			}
			return
		}
	}

	// single value case
	theNumType := whichNumeralType(val)

	switch theNumType {
	case Arabic:
		convVal := makeInt64(val)
		if *writeFileF {
			writeToFile([]string{formatValue(convVal, arabicToRoman(convVal), Roman)})
		} else {
			fmt.Println(formatValue(convVal, arabicToRoman(convVal), Roman))
		}

	case Roman:
		if *writeFileF {
			writeToFile([]string{formatValue(romanToArabic(val), val, Arabic)})
		} else {
			fmt.Println(formatValue(romanToArabic(val), val, Arabic))
		}

	default:
		log.Fatalf("%s is not defined and is neither roman or arabic\n", val)
	}
}

func whichNumeralType(str string) NumType {
	// is the value Roman, Arabic, or neither
	var arabicPattern = regexp.MustCompile(`^[1-9]\d*$`)
	var romanPattern = regexp.MustCompile(`^[IVXLCDM]+$`)

	if arabicPattern.MatchString(str) {
		return Arabic
	}

	if romanPattern.MatchString(str) {
		return Roman
	}

	return UnDef
}

func isValArabic(num int64) error {
	// number must be less than 4000
	if num > 4000 {
		return fmt.Errorf("%d is greater than 4000", num)
	}

	if num < 1 {
		return fmt.Errorf("%d is less than 1", num)
	}

	return nil
}

func makeInt64(str string) int64 {
	convVal, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		log.Fatalf("%s cannot be converted to an int!\n", str)
	}
	if err := isValArabic(convVal); err != nil {
		log.Fatalf("%v", err)
	}
	return convVal
}

func romanToArabic(str string) int64 {
	// take a valid roman numeral and return an arabic numeral
	//	run through string left to right, check two chars if avaialble against map
	//	and grab value to add to running total until chars are exhausted

	var val int64

	for i := 0; i < len(str); {
		c := str[i]

		// grab next char if possible
		var xc []byte
		if i+1 < len(str) {
			xc = append([]byte{c}, str[i+1])
		}

		// check two char sequence first
		if v, fnd := rToaMap[string(xc)]; fnd {
			val += int64(v)
			i += 2
			continue
		}

		if v, fnd := rToaMap[string(c)]; fnd {
			val += int64(v)
			i++
			continue
		} else {
			log.Fatalf("%s was not found in roman to arabic map\n invalid character", str)
		}

	}

	return val
}

func arabicToRoman(val int64) string {
	// take an arabic numeral and return a roman numeral
	// loop over map to find greatst match for current value
	//	apend the value and decrease current by key
	// coninue until current is zero

	var (
		outStr  string
		current int64 = val
	)

	useMap := aTorMap
	if *addF {
		useMap = makeAddMap(aTorMap)
	}

	for current > 0 {
		a, r := findLargest(current, useMap)
		outStr += r
		current -= int64(a)
	}

	return outStr
}

func findLargest(n int64, m map[uint16]string) (a uint16, r string) {
	// find largest key in map <= n
	var (
		lAr uint16
		lRm string
	)
	for k, v := range m {
		if uint16(n) >= k && k > lAr { //assume n is in uint16 space
			lAr = k
			lRm = v
		}
	}

	return lAr, lRm
}

func makeAddMap(inmap map[uint16]string) map[uint16]string {
	// take exisitng A -> R map and
	// return new map sans subtractive elements
	outmap := map[uint16]string{}
	for k, v := range inmap {
		if len(v) == 1 { // assume all subtractive keys are two bytes
			outmap[k] = v
		}
	}
	return outmap
}

func formatValue(arVal int64, romVal string, outType NumType) string {
	var outStr string
	switch outType {
	case Roman:
		if *simpleOutF {
			outStr = romVal
		} else {
			outStr = fmt.Sprintf("%d = %s", arVal, romVal)
			if *addF {
				outStr += "\t (add)"
			}
		}
	case Arabic:
		if *simpleOutF {
			outStr = fmt.Sprintf("%d", arVal)
		} else {
			outStr = fmt.Sprintf("%s = %d", romVal, arVal)
		}
	default:
		outStr = "NA"
	}

	return outStr
}

func genRange() []string {
	outVals := []string{}
	for i := *startF; i <= *endF; i++ {
		appendStr := formatValue(int64(i), arabicToRoman(int64(i)), Roman)
		outVals = append(outVals, appendStr)
	}

	return outVals
}

func writeToFile(strs []string) {

	writeType := os.O_WRONLY | os.O_CREATE
	writeMode := "truncate"
	if *appendFileF {
		writeType = os.O_WRONLY | os.O_APPEND | os.O_CREATE
		writeMode = "append"
	}
	log.Printf("opening file %s in mode %v\n", *outpathF, writeMode)

	f, err := os.OpenFile(*outpathF, writeType, 0664)
	if err != nil {
		log.Fatalf("failed to open file %s for writing\n%v", *outpathF, err)
	}
	defer f.Close()

	b, err := f.WriteString(strings.Join(strs, "\n") + "\n")
	if err != nil {
		log.Fatalf("error writing to file %s\n%v", *outpathF, err)
	}
	log.Printf("wrote %d bytes to %s\n", b, *outpathF)
}
