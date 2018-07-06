package main

import (
	"fmt"
	"io"
	"os"
	"time"
)

// p checks the error e and panics if it's not null.
func p(e error) {
	if e != nil {
		panic(e)
	}
}

func runTest() {
	f, err := os.Open("example_data/test.log")
	p(err)
	defer f.Close()

	var n int
	// var totalBytes int
	var buff = make([]byte, 10)
	lf := []byte("\n")
	line := make([]byte, 0, 200)
	// remander := make([]byte, 0, 0)
	// eol := false

	// Read the file until we hit an EOF.
	for err != io.EOF {

		// Read up to our buffer.
		n, err = f.Read(buff)
		fmt.Printf("Read in %d lines form buffer as: %s\n", n, string(buff))
		// Probaly should check for non io.EOF errors...

		// Check for an end to the line. If so save off the line so far.
		var i int
		for i = 0; i < n; i++ {

			if buff[i] == lf[0] {
				fmt.Println("Foudn new line!")
				// Line feed found so save off the line.
				line = append(line, buff[:i]...)
				fmt.Printf("Added everythign but new line to string: [%s]\n", string(line))
				fmt.Printf("LINE WAS [%s]\n", string(line))
				line = line[:0] // Reset the line buffer.
				fmt.Printf("Reset line to: [%s]\n", string(line))
				line = append(line, buff[i+1:n]...)
				fmt.Printf("Adding remaining items [%s]\n", string(buff[i:n]))
				fmt.Printf("Current line is: [%s]\n", string(line))
				continue
			}
		}

		// Add any remaining to the line.
		fmt.Printf("Adding full buffer to line: [%s]\n", string(buff[:n]))
		line = append(line, buff[:n]...)

		// if len(remander) > 0 {
		// 	line = append(line, remander...)
		// }

		// n, err = f.Read(buff)

		// eol = false
		// var i int
		// for i = 0; i < n; i++ {
		// 	if buff[i] == lf[0] {
		// 		line = append(line, buff[:i-1]...)
		// 		eol = true
		// 		break
		// 	}
		// }
		// totalBytes += n
		// if !eol {
		// 	line = append(line, buff...)
		// }

		// // save leftovers
		// if n > i {
		// 	remander = append(remander, buff[i:n]...)
		// }

		// fmt.Printf("LINE WAS [%s]", string(line))
	}

}

func runTest2() {

	fmt.Println("running test")

	f, err := os.Open("example_data/test.log")
	p(err)
	defer f.Close()

	var n int
	var buffSize = 10
	var buff = make([]byte, buffSize)
	var lf = []byte("\n")
	var cr = []byte("\r")
	var line = make([]byte, 0, 200)
	var offset int64

	for {
		_, err = f.Seek(offset, 0)
		p(err)

	LINELOOP:
		for err != io.EOF {
			// Reading up to the buffer length.
			n, err = f.Read(buff)

			// Check each character in the buffer for line feed \n or carraige return \r.
			// Finding it means the line has ended and we should save it off.
			for i := 0; i < n; i++ {

				// [v][f][f][cr][  lf ][x][x][x] n = 8
				// [    0:3    ][ 3:4 ][  i+1: ]  0:9
				// [    0:i    ][i:i+1]
				switch {
				case buff[i] == cr[0]:
					// Add anything before the \r to the line and saving it.
					line = append(line, buff[:i]...)
					if len(line) > 0 {
						fmt.Printf("FOUND LINE: [%s]\n", string(line))
						offset += int64(len(line)) + 2 // Adding 2 for \r\n
					}
					line = line[:0]

					// Checking to see if the returned data is larger enough for another character
					// and if so contains a line feed. If so kick it out by pushing i up one.
					if (i+1 < n) && buff[i+1] == lf[0] {
						i++
					}

					// If there is data left, add it to the line.
					if i+1 < n {
						line = append(line, buff[i+1:]...)
					}
					continue LINELOOP

				case buff[i] == lf[0]:
					// Should only reach here if no \r was found. So simply add everything before to the line,
					// save it. Then add anythign left to he new line.
					line = append(line, buff[:i]...)
					// Saving the line if not empty.
					if len(line) > 0 {
						fmt.Printf("FOUND LINE: [%s]\n", string(line))
						offset += int64(len(line)) + 1 // Adding 1 for \n
					}

					line = line[:0]

					if i+1 < n {
						line = append(line, buff[i+1:]...)
					}

					continue LINELOOP
				}
			}

			// Add all data from the buffer
			line = append(line, buff[:n]...)
		}

		time.Sleep(time.Duration(30) * time.Second)
	}
}

func runTest3() {

	fmt.Println("running test")

	f, err := os.Open("example_data/test.log")
	p(err)
	defer f.Close()

	var n int
	var buffSize = 10
	var buff = make([]byte, buffSize)
	var lf = []byte("\n")
	var cr = []byte("\r")
	var line = make([]byte, 0, 200)
	var offset int64

	for {
		_, err = f.Seek(offset, 0)
		p(err)

		for err != io.EOF {
			// Reading up to the buffer length.
			n, err = f.Read(buff)

			// Check each character in the buffer for line feed \n or carraige return \r.
			// Finding it means the line has ended and we should save it off. Then continue on.
			var startIndex = 0
			for i := 0; i < n; i++ {

				// [v][f][f][cr][  lf ][x][x][x] n = 8
				// [    0:3    ][ 3:4 ][  i+1: ]  0:9
				// [    0:i    ][i:i+1]
				switch {
				case buff[i] == cr[0]:
					// Add anything before the \r to the line and saving it.
					line = append(line, buff[startIndex:i]...)
					if len(line) > 0 {
						fmt.Printf("FOUND LINE: [%s]\n", string(line))
						offset += int64(len(line)) + 2 // Adding 2 for \r\n
					}
					line = line[:0]

					// Checking to see if the returned data is larger enough for another character
					// and if so contains a line feed. If so kick it out by pushing i up one.
					if (i+1 < n) && buff[i+1] == lf[0] {
						i++
					}

					// Update the start index to be the next value.
					startIndex = i + 1

					// If there is data left, add it to the line.
					// if i+1 < n {
					// 	line = append(line, buff[i+1:]...)
					// }
					continue

				case buff[i] == lf[0]:
					// Should only reach here if no \r was found. So simply add everything before to the line,
					// save it. Then add anythign left to he new line.
					line = append(line, buff[startIndex:i]...)
					// Saving the line if not empty.
					if len(line) > 0 {
						fmt.Printf("FOUND LINE: [%s]\n", string(line))
						offset += int64(len(line)) + 1 // Adding 1 for \n
					}

					line = line[:0]
					// Update the start index to be the next value.
					startIndex = i + 1

					// if i+1 < n {
					// 	line = append(line, buff[i+1:]...)
					// }

					continue
				}
			}

			// Add all data from the buffer
			line = append(line, buff[startIndex:n]...)
		}

		time.Sleep(time.Duration(30) * time.Second)
	}
}
