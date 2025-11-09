
// go mod init github.com/markburgess/MusicSST
// go mod tidy

//******************************************************************
//
// Demo of node by node addition, assuming that the arrows are predefined
//
//******************************************************************

package main

import (
	"fmt"
	"os"
        "github.com/dhowden/tag"
)

//******************************************************************

func main() {

	f, err := os.Open("/home/mark/TESTFLAC.flac")


	m, err := tag.ReadFrom(f)

	if err != nil {
		fmt.Print(err)
	} else {

		fmt.Println("Data format", m.Format())

		t := m.Raw()

		for i,j := range t {
			fmt.Printf("Key: %s = %s\n",i,j)
		}

	f.Close()
	}
}

