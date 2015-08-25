package imagestore

import (
	"crypto/rand"
	"log"
)

// Provides a continuous stream of random image "hashes" of a fixed length that is unique (does not exist in the store).
type HashGenerator struct {
	hashGetter chan string
	length     int
	store      ImageStore
}

func (this *HashGenerator) init() {
	go func() {
		storeObj := &StoreObject{
			"",
			"",
			"original",
			"",
		}

		for {
			str := ""

			for len(str) < this.length {
				c := 10
				bArr := make([]byte, c)
				_, err := rand.Read(bArr)
				if err != nil {
					log.Println("error:", err)
					break
				}

				for _, b := range bArr {
					if len(str) == this.length {
						break
					}

					/**
					 * Each byte will be in [0, 256), but we only care about:
					 *
					 * [48, 57]     0-9
					 * [65, 90]     A-Z
					 * [97, 122]    a-z
					 *
					 * Which means that the highest bit will always be zero, since the last byte with high bit
					 * zero is 01111111 = 127 which is higher than 122. Lower our odds of having to re-roll a byte by
					 * dividing by two (right bit shift of 1).
					 */

					b = b >> 1

					// The byte is any of        0-9                  A-Z                      a-z
					byteIsAllowable := (b >= 48 && b <= 57) || (b >= 65 && b <= 90) || (b >= 97 && b <= 122)

					if byteIsAllowable {
						str += string(b)
					}
				}

			}

			storeObj.Id = str

			exists, _ := this.store.Exists(storeObj)
			if !exists {
				this.hashGetter <- str
			}
		}
	}()
}

func (this *HashGenerator) Get() string {
	return <-this.hashGetter
}
