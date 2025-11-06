package qrcode

import (
	"fmt"
	"strings"

	qr "github.com/skip2/go-qrcode"
)

type QRCode struct {
	modules  [][]bool
	reserved [][]bool
	size     int
	version  int
}

const (
	ErrorCorrectionLow = iota
	ErrorCorrectionMedium
	ErrorCorrectionQuartile
	ErrorCorrectionHigh
)

type versionInfo struct {
	capacity      int
	ecCodewords   int
	dataCodewords int
}

var versionTable = []versionInfo{
	{capacity: 26, ecCodewords: 7, dataCodewords: 19},
	{capacity: 44, ecCodewords: 10, dataCodewords: 34},
	{capacity: 70, ecCodewords: 15, dataCodewords: 55},
	{capacity: 100, ecCodewords: 20, dataCodewords: 80},
	{capacity: 134, ecCodewords: 26, dataCodewords: 108},
}

var galoisExp [256]int
var galoisLog [256]int

func init() {
	x := 1
	for i := 0; i < 255; i++ {
		galoisExp[i] = x
		galoisLog[x] = i
		x *= 2
		if x > 255 {
			x ^= 0x11D
		}
	}
	galoisExp[255] = galoisExp[0]
}

func galoisMultiply(a, b int) int {
	if a == 0 || b == 0 {
		return 0
	}
	return galoisExp[(galoisLog[a]+galoisLog[b])%255]
}

func galoisDivide(a, b int) int {
	if a == 0 {
		return 0
	}
	if b == 0 {
		panic("division by zero")
	}
	return galoisExp[(galoisLog[a]-galoisLog[b]+255)%255]
}

func Encode(data string, level int) (*QRCode, error) {
	version := calculateVersion(len(data))
	size := version*4 + 17

	qr := &QRCode{
		size:     size,
		version:  version,
		modules:  make([][]bool, size),
		reserved: make([][]bool, size),
	}

	for i := range qr.modules {
		qr.modules[i] = make([]bool, size)
		qr.reserved[i] = make([]bool, size)
	}

	qr.addFinderPatterns()
	qr.addSeparators()
	qr.addTimingPatterns()
	qr.addDarkModule()
	qr.addFormatInfo(level, 0)

	dataBytes := encodeData(data, version, level)
	dataWithEC := addErrorCorrection(dataBytes, version, level)
	qr.placeData(dataWithEC)

	bestMask := qr.selectBestMask(level)
	qr.applyMask(bestMask)
	qr.addFormatInfo(level, bestMask)

	return qr, nil
}

func encodeData(data string, version, level int) []byte {
	modeIndicator := []byte{0x40}

	dataBytes := []byte(data)
	charCount := len(dataBytes)

	charCountBits := encodeNumber(charCount, 8)

	var bits []byte
	bits = append(bits, modeIndicator...)
	bits = append(bits, charCountBits...)

	for _, b := range dataBytes {
		bits = append(bits, encodeByte(b)...)
	}

	terminator := []byte{0x00, 0x00, 0x00, 0x00}
	bits = append(bits, terminator...)

	for len(bits)%8 != 0 {
		bits = append(bits, 0x00)
	}

	result := []byte{}
	for i := 0; i < len(bits); i += 8 {
		b := byte(0)
		for j := 0; j < 8 && i+j < len(bits); j++ {
			if bits[i+j] != 0 {
				b |= 1 << (7 - j)
			}
		}
		result = append(result, b)
	}

	capacity := getDataCapacity(version, level)
	pad := []byte{0xEC, 0x11}
	for len(result) < capacity {
		result = append(result, pad[len(result)%2])
	}

	return result[:capacity]
}

func encodeNumber(n, bits int) []byte {
	result := make([]byte, bits)
	for i := 0; i < bits; i++ {
		if (n & (1 << (bits - 1 - i))) != 0 {
			result[i] = 0x01
		}
	}
	return result
}

func encodeByte(b byte) []byte {
	result := make([]byte, 8)
	for i := 0; i < 8; i++ {
		if (b & (1 << (7 - i))) != 0 {
			result[i] = 0x01
		}
	}
	return result
}

func getDataCapacity(version, level int) int {
	capacities := map[int]int{
		1: 19, 2: 34, 3: 55, 4: 80, 5: 108,
	}
	return capacities[version]
}

func (qr *QRCode) addFinderPatterns() {
	positions := [][2]int{
		{0, 0},
		{qr.size - 7, 0},
		{0, qr.size - 7},
	}

	for _, pos := range positions {
		qr.addFinderPattern(pos[0], pos[1])
	}
}

func (qr *QRCode) addFinderPattern(row, col int) {
	for r := -1; r <= 7; r++ {
		for c := -1; c <= 7; c++ {
			nr, nc := row+r, col+c
			if nr >= 0 && nr < qr.size && nc >= 0 && nc < qr.size {
				isBlack := false
				if r >= 0 && r <= 6 && c >= 0 && c <= 6 {
					if (r == 0 || r == 6) || (c == 0 || c == 6) || (r >= 2 && r <= 4 && c >= 2 && c <= 4) {
						isBlack = true
					}
				}
				qr.modules[nr][nc] = isBlack
				qr.reserved[nr][nc] = true
			}
		}
	}
}

func (qr *QRCode) addTimingPatterns() {
	for i := 8; i < qr.size-8; i++ {
		qr.modules[6][i] = i%2 == 0
		qr.modules[i][6] = i%2 == 0
		qr.reserved[6][i] = true
		qr.reserved[i][6] = true
	}
}

func (qr *QRCode) addSeparators() {
	positions := [][2]int{
		{7, 0}, {0, 7},
		{qr.size - 8, 0}, {qr.size - 7, 7},
		{7, qr.size - 8}, {0, qr.size - 7},
	}

	for _, pos := range positions {
		row, col := pos[0], pos[1]
		for i := 0; i < 8; i++ {
			if row < qr.size && col+i < qr.size {
				qr.reserved[row][col+i] = true
			}
			if row+i < qr.size && col < qr.size {
				qr.reserved[row+i][col] = true
			}
		}
	}
}

func (qr *QRCode) addDarkModule() {
	qr.modules[4*qr.version+9][8] = true
	qr.reserved[4*qr.version+9][8] = true
}

func (qr *QRCode) addFormatInfo(ecLevel, maskPattern int) {
	formatBits := calculateFormatBits(ecLevel, maskPattern)

	for i := 0; i < 15; i++ {
		bit := (formatBits >> i) & 1
		module := bit == 1

		if i < 6 {
			qr.modules[8][i] = module
			qr.reserved[8][i] = true
		} else if i < 8 {
			qr.modules[8][i+1] = module
			qr.reserved[8][i+1] = true
		} else if i < 9 {
			qr.modules[7][8] = module
			qr.reserved[7][8] = true
		} else {
			qr.modules[14-i][8] = module
			qr.reserved[14-i][8] = true
		}

		if i < 8 {
			qr.modules[qr.size-1-i][8] = module
			qr.reserved[qr.size-1-i][8] = true
		} else {
			qr.modules[8][qr.size-15+i] = module
			qr.reserved[8][qr.size-15+i] = true
		}
	}
}

func calculateFormatBits(ecLevel, maskPattern int) int {
	data := (ecLevel << 3) | maskPattern

	remainder := data
	for i := 0; i < 10; i++ {
		if (remainder & (1 << (9 + i))) != 0 {
			remainder ^= 0x537 << i
		}
	}

	formatInfo := (data << 10) | remainder
	formatInfo ^= 0x5412

	return formatInfo
}

func addErrorCorrection(data []byte, version, level int) []byte {
	info := versionTable[version-1]
	ecCount := info.ecCodewords

	generator := generateECPolynomial(ecCount)

	dataExtended := make([]int, len(data)+ecCount)
	for i, b := range data {
		dataExtended[i] = int(b)
	}

	for i := 0; i < len(data); i++ {
		coef := dataExtended[i]
		if coef != 0 {
			for j := 0; j < len(generator); j++ {
				dataExtended[i+j] ^= galoisMultiply(generator[j], coef)
			}
		}
	}

	result := make([]byte, len(data)+ecCount)
	for i, b := range data {
		result[i] = b
	}
	for i := 0; i < ecCount; i++ {
		result[len(data)+i] = byte(dataExtended[len(data)+i])
	}

	return result
}

func generateECPolynomial(degree int) []int {
	poly := make([]int, degree+1)
	poly[0] = 1

	for i := 0; i < degree; i++ {
		for j := len(poly) - 1; j > 0; j-- {
			poly[j] = galoisMultiply(poly[j], galoisExp[i])
			if j > 0 {
				poly[j] ^= poly[j-1]
			}
		}
		poly[0] = galoisMultiply(poly[0], galoisExp[i])
	}

	return poly
}

func (qr *QRCode) selectBestMask(ecLevel int) int {
	bestMask := 0
	bestScore := 999999

	for mask := 0; mask < 8; mask++ {
		qrCopy := qr.copy()
		qrCopy.applyMask(mask)
		score := qrCopy.calculatePenalty()

		if score < bestScore {
			bestScore = score
			bestMask = mask
		}
	}

	return bestMask
}

func (qr *QRCode) copy() *QRCode {
	newQR := &QRCode{
		size:     qr.size,
		version:  qr.version,
		modules:  make([][]bool, qr.size),
		reserved: make([][]bool, qr.size),
	}

	for i := range qr.modules {
		newQR.modules[i] = make([]bool, qr.size)
		newQR.reserved[i] = make([]bool, qr.size)
		copy(newQR.modules[i], qr.modules[i])
		copy(newQR.reserved[i], qr.reserved[i])
	}

	return newQR
}

func (qr *QRCode) calculatePenalty() int {
	penalty := 0

	for row := 0; row < qr.size; row++ {
		runColor := qr.modules[row][0]
		runLength := 1
		for col := 1; col < qr.size; col++ {
			if qr.modules[row][col] == runColor {
				runLength++
			} else {
				if runLength >= 5 {
					penalty += runLength - 2
				}
				runColor = qr.modules[row][col]
				runLength = 1
			}
		}
		if runLength >= 5 {
			penalty += runLength - 2
		}
	}

	for col := 0; col < qr.size; col++ {
		runColor := qr.modules[0][col]
		runLength := 1
		for row := 1; row < qr.size; row++ {
			if qr.modules[row][col] == runColor {
				runLength++
			} else {
				if runLength >= 5 {
					penalty += runLength - 2
				}
				runColor = qr.modules[row][col]
				runLength = 1
			}
		}
		if runLength >= 5 {
			penalty += runLength - 2
		}
	}

	return penalty
}

func (qr *QRCode) placeData(data []byte) {
	col := qr.size - 1
	row := qr.size - 1
	direction := -1
	byteIndex := 0
	bitIndex := 7

	for col > 0 {
		if col == 6 {
			col--
		}

		for {
			for dc := 0; dc < 2; dc++ {
				c := col - dc

				if !qr.reserved[row][c] {
					bit := false
					if byteIndex < len(data) {
						bit = (data[byteIndex] & (1 << bitIndex)) != 0
					}
					qr.modules[row][c] = bit

					bitIndex--
					if bitIndex < 0 {
						byteIndex++
						bitIndex = 7
					}
				}
			}

			row += direction
			if row < 0 || row >= qr.size {
				row -= direction
				direction = -direction
				break
			}
		}

		col -= 2
	}
}

func (qr *QRCode) applyMask(pattern int) {
	for row := 0; row < qr.size; row++ {
		for col := 0; col < qr.size; col++ {
			if !qr.reserved[row][col] {
				invert := false
				switch pattern {
				case 0:
					invert = (row+col)%2 == 0
				case 1:
					invert = row%2 == 0
				case 2:
					invert = col%3 == 0
				case 3:
					invert = (row+col)%3 == 0
				case 4:
					invert = (row/2+col/3)%2 == 0
				case 5:
					invert = (row*col)%2+(row*col)%3 == 0
				case 6:
					invert = ((row*col)%2+(row*col)%3)%2 == 0
				case 7:
					invert = ((row+col)%2+(row*col)%3)%2 == 0
				}

				if invert {
					qr.modules[row][col] = !qr.modules[row][col]
				}
			}
		}
	}
}

func (qr *QRCode) ToASCII() string {
	var result strings.Builder

	for y := 0; y < qr.size; y += 2 {
		for x := 0; x < qr.size; x++ {
			top := qr.modules[y][x]
			bottom := false
			if y+1 < qr.size {
				bottom = qr.modules[y+1][x]
			}

			if top && bottom {
				result.WriteString("██")
			} else if top && !bottom {
				result.WriteString("▀▀")
			} else if !top && bottom {
				result.WriteString("▄▄")
			} else {
				result.WriteString("  ")
			}
		}
		result.WriteString("\n")
	}

	return result.String()
}

func (qr *QRCode) String() string {
	return qr.ToASCII()
}

func calculateVersion(dataLen int) int {
	if dataLen <= 17 {
		return 1
	} else if dataLen <= 32 {
		return 2
	} else if dataLen <= 53 {
		return 3
	} else if dataLen <= 78 {
		return 4
	}
	return 5
}

func Generate(data string) (string, error) {
	code, err := qr.New(data, qr.Medium)
	if err != nil {
		return "", err
	}

	bitmap := code.Bitmap()
	var result strings.Builder

	for y := 0; y < len(bitmap); y += 2 {
		for x := 0; x < len(bitmap[y]); x++ {
			top := bitmap[y][x]
			bottom := false
			if y+1 < len(bitmap) {
				bottom = bitmap[y+1][x]
			}

			if top && bottom {
				result.WriteString("██")
			} else if top && !bottom {
				result.WriteString("▀▀")
			} else if !top && bottom {
				result.WriteString("▄▄")
			} else {
				result.WriteString("  ")
			}
		}
		result.WriteString("\n")
	}

	return result.String(), nil
}

func GenerateWithBorder(data string, border int) (string, error) {
	ascii, err := Generate(data)
	if err != nil {
		return "", err
	}

	lines := strings.Split(strings.TrimRight(ascii, "\n"), "\n")

	if border > 0 {
		var result strings.Builder
		lineWidth := len(lines[0])
		borderLine := strings.Repeat(" ", lineWidth+border*2) + "\n"

		for i := 0; i < border; i++ {
			result.WriteString(borderLine)
		}

		for _, line := range lines {
			result.WriteString(strings.Repeat(" ", border))
			result.WriteString(line)
			result.WriteString(strings.Repeat(" ", border))
			result.WriteString("\n")
		}

		for i := 0; i < border; i++ {
			result.WriteString(borderLine)
		}

		return result.String(), nil
	}

	return ascii, nil
}

func Print(data string) error {
	code, err := GenerateWithBorder(data, 2)
	if err != nil {
		return err
	}
	fmt.Print(code)
	return nil
}
