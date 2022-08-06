package main

import (
	"bufio"
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"golang.org/x/term"
)

const (
	MINE       byte = 0b0010_0000
	OPEN       byte = 0b0100_0000
	FLAG       byte = 0b1000_0000
	VALUE_MASK byte = 0b0000_1111
)

var (
	reader    = bufio.NewReader(os.Stdin)
	output    = bufio.NewWriter(os.Stdout)
	screen    = new(bytes.Buffer)
	selected  = 0
	x, y      = 8, 8
	minecount = 10
	r         = rand.New(rand.NewSource(time.Now().UnixNano()))

	correctflags = 0
)

func main() {
	setup()
	defer cleanup()
	originalTermState, err := term.MakeRaw(int(os.Stdin.Fd()))
	defer term.Restore(int(os.Stdin.Fd()), originalTermState)
	if err != nil {
		panic(err)
	}
	grid := newGrid()
loop:
	for {
		moveCursor(1, 1)
		clear()
		printGrid(grid)
		flush()
		key, _, err := reader.ReadRune()
		if err != nil {
			panic(err)
		}
		switch key {
		case 3:
			break loop
		case 'q':
			break loop
		case 'j':
			selected = clamp(selected + x)
		case 'k':
			selected = clamp(selected - x)
		case 'l':
			selected = clamp(selected + 1)
		case 'h':
			selected = clamp(selected - 1)
		case 'f':
			placeFlag(grid)
		case 'd':
			open(grid)
		}
	}
}

func setup() {
	clear()
	moveCursor(1, 1)
	hideCursor()
	flush()
}

func cleanup() {
	clear()
	moveCursor(1, 1)
	showCursor()
	flush()
}

func hideCursor() {
	fmt.Fprintf(screen, "\033[?25l")
}

func showCursor() {
	fmt.Fprintf(screen, "\033[?25h")
}

func clear() {
	fmt.Fprintf(screen, "\033[2J")
}

func flush() {
	_, height := getSize()
	for idx, str := range strings.SplitAfter(screen.String(), "\n") {
		if idx > height {
			return
		}
		output.WriteString(str)
	}
	output.Flush()
	screen.Reset()
}

func getSize() (width int, height int) {
	width, height, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	return width, height
}

func printGrid(grid []byte) {
	for idx, val := range grid {
		display := ""
		if isOpen(val) {
			if isMine(val) {
				display = "*"
			} else {
				switch getValue(val) {
				case 0:
					display = " "
				case 1:
					display = fmt.Sprintf("\033[34m%d\033[0m", getValue(val))
				case 2:
					display = fmt.Sprintf("\033[32m%d\033[0m", getValue(val))
				case 3:
					display = fmt.Sprintf("\033[33m%d\033[0m", getValue(val))
				case 4:
					display = fmt.Sprintf("\033[35m%d\033[0m", getValue(val))
				case 5:
					display = fmt.Sprintf("\033[36m%d\033[0m", getValue(val))
				case 6:
					display = fmt.Sprintf("\033[37m%d\033[0m", getValue(val))
				case 7:
					display = fmt.Sprintf("\x1b[2m\033[35m%d\033[0m\x1b[22m", getValue(val))
				default:
					display = fmt.Sprintf("%d", getValue(val))
				}

			}
		} else if isFlag(val) {
			display = "\x1b[2m\033[33m?\033[0m\x1b[22m"
		} else {
			display = "#"
		}
		if idx == selected {
			display = "[" + display + "]"
		} else {
			display = " " + display + " "
		}
		if (idx+1)%x == 0 {
			display += "\n\r"
		}
		fmt.Fprintf(screen, display)
	}
	if correctflags == minecount {
		fmt.Fprintf(screen, "-- YOU WIN --")
	}
}

func getValue(field byte) byte {
	return VALUE_MASK & field
}

func isFlag(field byte) bool {
	return FLAG&field == FLAG
}

func isMine(field byte) bool {
	return MINE&field == MINE
}

func isOpen(field byte) bool {
	return OPEN&field == OPEN
}

func open(grid []byte) {
	grid[selected] |= OPEN
	if isFlag(grid[selected]) {
		grid[selected] ^= FLAG
	}
	if isMine(grid[selected]) {
		for idx, val := range grid {
			if isMine(val) {
				grid[idx] |= OPEN
			}
		}
		return
	}
	fieldsToCheck := make([]int, 0)
	idx := selected
	if getValue(grid[idx]) != 0 {
		return
	}
	for {
		isTopRow := idx/y == 0
		isBottomRow := idx/y == y-1
		isLeftCol := idx%x == 0
		isRightCol := idx%x == x-1
		if !isTopRow && !isLeftCol {
			if grid[(idx-x)-1] == 0 {
				fieldsToCheck = append(fieldsToCheck, (idx-x)-1)
			}
			grid[(idx-x)-1] |= OPEN
		}
		if !isTopRow {
			if grid[idx-x] == 0 {
				fieldsToCheck = append(fieldsToCheck, idx-x)
			}
			grid[(idx - x)] |= OPEN
		}
		if !isTopRow && !isRightCol {
			if grid[(idx-x)+1] == 0 {
				fieldsToCheck = append(fieldsToCheck, (idx-x)+1)
			}
			grid[(idx-x)+1] |= OPEN
		}
		if !isLeftCol {
			if grid[idx-1] == 0 {
				fieldsToCheck = append(fieldsToCheck, (idx - 1))
			}
			grid[idx-1] |= OPEN
		}
		if !isRightCol {
			if grid[idx+1] == 0 {
				fieldsToCheck = append(fieldsToCheck, idx+1)
			}
			grid[idx+1] |= OPEN
		}
		if !isBottomRow && !isLeftCol {
			if grid[(idx+x)-1] == 0 {
				fieldsToCheck = append(fieldsToCheck, (idx+x)-1)
			}
			grid[(idx+x)-1] |= OPEN
		}
		if !isBottomRow {
			if grid[idx+x] == 0 {
				fieldsToCheck = append(fieldsToCheck, (idx + x))
			}
			grid[idx+x] |= OPEN
		}
		if !isBottomRow && !isRightCol {
			if grid[(idx+x)+1] == 0 {
				fieldsToCheck = append(fieldsToCheck, (idx+x)+1)
			}
			grid[(idx+x)+1] |= OPEN
		}
		if len(fieldsToCheck) > 0 {
			idx, fieldsToCheck = fieldsToCheck[0], fieldsToCheck[1:]
		} else {
			break
		}
	}
}

func placeFlag(grid []byte) {
	if !isOpen(grid[selected]) {
		if isFlag(grid[selected]) {
			grid[selected] &^= FLAG
			if isMine(grid[selected]) {
				correctflags -= 1
			}
		} else {
			grid[selected] |= FLAG
			if isMine(grid[selected]) {
				correctflags += 1
			}
		}
	}
}

func moveCursor(x int, y int) {
	fmt.Fprintf(screen, "\033[%d;%dH", x, y)
}

func clamp(value int) int {
	if value >= x*y {
		value = (x * y) - 1
	} else if value < 0 {
		value = 0
	}
	return value
}

func newGrid() (grid []byte) {
	grid = make([]byte, x*y)
	placedMines := 0
	for placedMines < minecount {
		pos := r.Intn(x * y)
		if !isMine(grid[pos]) {
			grid[pos] |= MINE
			placedMines += 1
		}
	}
	for idx, val := range grid {
		if isMine(val) {
			continue
		}
		isTopRow := idx/y == 0
		isBottomRow := idx/y == y-1
		isLeftCol := idx%x == 0
		isRightCol := idx%x == x-1
		if !isTopRow && !isLeftCol {
			if isMine(grid[(idx-x)-1]) {
				grid[idx] += 1
			}
		}
		if !isTopRow {
			if isMine(grid[idx-x]) {
				grid[idx] += 1
			}
		}
		if !isTopRow && !isRightCol {
			if isMine(grid[(idx-x)+1]) {
				grid[idx] += 1
			}
		}
		if !isLeftCol {
			if isMine(grid[idx-1]) {
				grid[idx] += 1
			}
		}
		if !isRightCol {
			if isMine(grid[idx+1]) {
				grid[idx] += 1
			}
		}
		if !isBottomRow && !isLeftCol {
			if isMine(grid[(idx+x)-1]) {
				grid[idx] += 1
			}
		}
		if !isBottomRow {
			if isMine(grid[idx+x]) {
				grid[idx] += 1
			}
		}
		if !isBottomRow && !isRightCol {
			if isMine(grid[(idx+x)+1]) {
				grid[idx] += 1
			}
		}
	}
	return grid
}
