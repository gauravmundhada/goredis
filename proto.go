package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/tidwall/resp"
)

type Command interface {
}

type SetCommand struct {
	key, val []byte
}

type GetCommand struct {
	key []byte
}

func parseCommand(raw string) (Command, error) {
	rd := resp.NewReader(bytes.NewBufferString(raw))

	for {
		v, _, err := rd.ReadValue()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(v)
		fmt.Printf("Read %s\n", v.Type())
		if v.Type() == resp.Array {
			for _, val := range v.Array() {
				//fmt.Println(val.String())
				switch strings.ToUpper(val.String()) {
				case "GET":
					fmt.Println(len(v.Array()))
					if len(v.Array()) != 2 {
						return nil, fmt.Errorf("Invalid number of arguments for 'GET' command")
					}
					cmd := GetCommand{
						key: v.Array()[1].Bytes(),
					}
					return cmd, nil
				case "SET":
					if len(v.Array()) != 3 {
						return nil, fmt.Errorf("Invalid number of arguments for 'SET' command")
					}
					cmd := SetCommand{
						key: v.Array()[1].Bytes(),
						val: v.Array()[2].Bytes(),
					}
					return cmd, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("Unknown command")
}
