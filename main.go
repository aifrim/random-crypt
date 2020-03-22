package main

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	"log"
	"math/big"
	"os"
)

const BlockSize = 16

var kSum int

/**
TODO: read more than one block per time - read like 64 blocks
*/

func main() {
	keyPath := flag.String("key", "", "the path to the keyPath")
	inputPath := flag.String("input", "", "the path to the file")
	action := flag.String("action", "encrypt", "encrypt / decrypt the file")

	flag.Parse()

	fmt.Println("Reading Key: ", *keyPath)
	key := readKey(*keyPath, BlockSize)
	kSum = int(key[0]+key[1]+key[2]+key[3]+
		key[4]+key[5]+key[6]+key[7]+
		key[8]+key[9]+key[10]+key[11]+
		key[12]+key[13]+key[14]+key[15])%32 + 1

	fmt.Println("Key", kSum, key)

	fmt.Println("Reading inputPath from: ", *inputPath)
	if *action == "encrypt" {
		encrypt(*inputPath, key)
	} else if *action == "decrypt" {
		decrypt(*inputPath, key)
	}
}

func check(err error) {
	if err != nil && err != io.EOF {
		panic(err)
	}
}

func readKey(keyPath string, len int) []byte {
	f, err := os.Open(keyPath)
	check(err)

	defer f.Close()

	r := bufio.NewReader(f)

	buf := make([]byte, len)
	l, err := r.Read(buf)

	if err != nil {
		log.Fatal(err)
	}

	if l != len {
		log.Fatal("Key file does not contain at least 256 bytes")
	}

	return buf
}

func encrypt(inputPath string, key []byte) {
	input, err := os.Open(inputPath)
	check(err)

	output, err := os.Create(inputPath + ".enc")
	check(err)

	defer input.Close()
	defer output.Close()

	iv := make([]byte, BlockSize)
	_, err = rand.Read(iv)
	check(err)

	_, err = output.Write(iv)
	check(err)

	enc, err := aes.NewCipher(key)
	check(err)
	ctr := cipher.NewCTR(enc, iv)

	outBuf := make([]byte, BlockSize)
	outBuf1 := make([]byte, BlockSize)
	outBuf2 := make([]byte, BlockSize)
	plainText := make([]byte, BlockSize)

	first := true
	do := false
	for {

		if first {
			do = doSplit(iv, key)
			first = false
		} else {
			do = doSplit(plainText, key)
		}

		plainText = make([]byte, BlockSize)
		l, err := input.Read(plainText)
		check(err)

		if l != 0 {
			if !do {
				ctr.XORKeyStream(outBuf, plainText[:BlockSize])
				_, err := output.Write(outBuf)
				check(err)
			} else {
				var plain1 []byte
				var plain2 []byte

				plain1, plain2 = getBlocks(plainText)

				ctr.XORKeyStream(outBuf1, plain1[:BlockSize])
				_, err := output.Write(outBuf1)
				check(err)

				ctr.XORKeyStream(outBuf2, plain2[:BlockSize])
				_, err = output.Write(outBuf2)
				check(err)
			}
		} else {
			return
		}
	}
}

func decrypt(inputPath string, key []byte) {
	input, err := os.Open(inputPath)
	check(err)

	size, err := getFileSize(inputPath)
	check(err)

	maxBlocks := size/16 - 1

	output, err := os.Create(inputPath + ".dec")
	check(err)

	defer input.Close()
	defer output.Close()

	outBuf := make([]byte, BlockSize)
	outBuf1 := make([]byte, BlockSize)
	outBuf2 := make([]byte, BlockSize)

	iv := make([]byte, BlockSize)
	_, err = input.Read(iv)
	check(err)

	enc, err := aes.NewCipher(key)
	check(err)
	ctr := cipher.NewCTR(enc, iv)

	var blocks int64 = 0
	first := true
	do := false
	cipher1 := make([]byte, BlockSize)
	cipher2 := make([]byte, BlockSize)
	for {
		l, err := input.Read(cipher1)
		check(err)

		if l != 0 {
			if first {
				do = doSplit(iv, key)
				first = false
			} else {
				do = doSplit(outBuf, key)
			}

			blocks++

			if !do {
				ctr.XORKeyStream(outBuf, cipher1[:BlockSize])

				if blocks == maxBlocks {
					outBuf = cleanBuff(outBuf)
				}

				_, err := output.Write(outBuf)
				check(err)
			} else {
				blocks++

				_, _ = input.Read(cipher2)
				ctr.XORKeyStream(outBuf1, cipher1[:BlockSize])
				ctr.XORKeyStream(outBuf2, cipher2[:BlockSize])

				outBuf = joinBlocks(outBuf1, outBuf2)

				if blocks == maxBlocks {
					outBuf = cleanBuff(outBuf)
				}

				_, err := output.Write(outBuf)
				check(err)
			}
		} else {
			return
		}
	}
}

func getBlocks(block []byte) ([]byte, []byte) {
	block1 := make([]byte, BlockSize)
	block2 := make([]byte, BlockSize)

	for i := 0; i < BlockSize; i++ {
		r1, r2 := getRandomPair(int(block[i]))

		block1[i] = byte(r1)
		block2[i] = byte(r2)
	}

	return block1, block2
}

func joinBlocks(block1 []byte, block2 []byte) []byte {
	block := make([]byte, BlockSize)

	for i := 0; i < BlockSize; i++ {
		block[i] = byte(int(block1[i]+block2[i]) % 256)
	}

	return block
}

func doSplit(block []byte, key []byte) bool {
	// TODO: logic that creates additional block needs to be IMPROVED

	do := int((block[0]^key[0])^(block[1]^key[1])^
		(block[2]^key[2])^(block[3]^key[3])^
		(block[4]^key[4])^(block[5]^key[5])^
		(block[6]^key[6])^(block[7]^key[7])^
		(block[8]^key[8])^(block[9]^key[9])^
		(block[10]^key[10])^(block[12]^key[12])^
		(block[1]^key[1])^(block[13]^key[1])^
		(block[14]^key[1])^(block[15]^key[1]))%kSum == 0

	return do
}

func getRandomPair(nr int) (int, int) {
	if nr == 0 {
		nr = 256
	}

	nrn, err := rand.Int(rand.Reader, big.NewInt(int64(nr)))
	check(err)

	n := int(nrn.Int64())

	return n, nr - n
}

func getFileSize(filepath string) (int64, error) {
	fi, err := os.Stat(filepath)

	if err != nil {
		return 0, err
	}
	return fi.Size(), nil
}

func cleanBuff(buf []byte) []byte {

	var retBuf []byte
	var size int
	add := false

	for i := BlockSize - 1; i >= 0; i-- {
		if buf[i] != 0 && !add {
			size = i + 1
			retBuf = make([]byte, size)
			add = true
		}

		if add {
			//noinspection ALL
			retBuf[i] = buf[i]
		}
	}

	return retBuf
}
