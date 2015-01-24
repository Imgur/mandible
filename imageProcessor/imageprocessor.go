package imageprocessor

import (
	"log"
	"fmt"
	"os"
	"io/ioutil"
)

func init() {
    hashGetter = make(chan string)

    go func() {
        for {
            length := 7
            str    := ""

            for len(str) < length {
                c := 10
                bArr := make([]byte, c)
                _, err := rand.Read(bArr)
                if err != nil {
                    log.Println("error:", err)
                    break
                }
                
                for _, b := range bArr {
                    if len(str) == length {
                        break;
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


                    b = b >> 1;

                    // The byte is any of        0-9                  A-Z                      a-z
                    byteIsAllowable := (b >= 48 && b <= 57) || (b >= 65 && b <= 90) || (b >= 97 && b <= 122);
                    
                    if byteIsAllowable {
                        str += string(b)
                    }
                }

            }
            
            hashGetter <- str
        }
    }()
}

type multiProcessType []ProcessType

type ProcessType interface {
    Process(filename string) (error)
}

type ImageProcessor struct {
    processor ProcessType
}


func Factory(scale bool) ImageProcessor {
    processor := 
    return ImageProcessor{}
}

