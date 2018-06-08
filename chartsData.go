package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

var (
	data [][3]string
)

func parseTempHumLogLine(line string) {
	elements := strings.Split(line, " ")
	time := elements[0][1:6]
	temp := elements[2]
	hum := elements[4]
	data = append(data, [3]string{time, hum, temp})
}

func loadTempHumPointsFromLogFile(filename string) error {
	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		fmt.Println("Can't read sensors.log")
		return err
	}

	// получить размер файла
	stat, _ := file.Stat()

	data = make([][3]string, 0)

	reader := bufio.NewReader(file)
	line, _ := reader.ReadString('\n')
	slen := len(line) + 1
	parseTempHumLogLine(line)
	linesCount := stat.Size() / int64(slen)
	linesCount--
	each := linesCount / 30 // нужно смещать на (each-1)*slen
	if each > 0 {
		each--
	}
	for i := 0; i < 29; i++ {
		if each > 0 {
			for j := int64(0); j < each; j++ {
				_, err = reader.ReadString('\n')
				if err != nil {
					break
				}
			}
		}
		line, err = reader.ReadString('\n')
		if err != nil {
			break
		}
		if len(line) == slen-1 {
			parseTempHumLogLine(line)
		}
	}

	for err == nil {
		line, err = reader.ReadString('\n')
		if err != nil {
			break
		}
		if len(line) == slen-1 {
			parseTempHumLogLine(line)
		}
	}

	if err == io.EOF {
		return nil
	}

	return err
}

func getCharTempDataJSONStr() string {
	p := make(map[string][]string)
	for _, elem := range data {
		p["data"] = append(p["data"], elem[2])
		p["time"] = append(p["time"], elem[0])
	}
	jsonString, _ := json.Marshal(p)
	return string(jsonString)
}

func getCharHumDataJSONStr() string {
	p := make(map[string][]string)
	for _, elem := range data {
		p["data"] = append(p["data"], elem[1])
		p["time"] = append(p["time"], elem[0])
	}
	jsonString, _ := json.Marshal(p)
	return string(jsonString)
}
