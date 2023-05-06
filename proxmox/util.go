package proxmox

import (
	"log"
	"regexp"
	"strconv"
	"strings"
)

var rxUserTokenExtract = regexp.MustCompile("[a-z0-9]+@[a-z0-9]+!([a-z0-9]+)")

func inArray(arr []string, str string) bool {
	for _, elem := range arr {
		if elem == str {
			return true
		}
	}

	return false
}

func Itob(i int) bool {
	return i == 1
}

func BoolInvert(b bool) bool {
	return !b
}

// Check the value of a key in a nested array of map[string]interface{}
func ItemInKeyOfArray(array []interface{}, key, value string) (existance bool) {
	//search for userid first
	for i := range array {
		item := array[i].(map[string]interface{})
		if string(item[key].(string)) == value {
			return true
		}
		if tok, keyok := item["tokens"]; keyok && tok != nil {
			if rxUserTokenExtract.MatchString(value) {
				matches := rxUserTokenExtract.FindStringSubmatch(value)
				for _, v := range tok.([]interface{}) {
					for _, v := range v.(map[string]interface{}) {
						if matches[1] == v {
							return true
						}
					}
				}
			}
		}
	}
	return false
}

// ParseSubConf - Parse standard sub-conf strings `key=value`.
func ParseSubConf(
	element string,
	separator string,
) (key string, value interface{}) {
	if strings.Contains(element, separator) {
		conf := strings.Split(element, separator)
		key, value := conf[0], conf[1]
		var interValue interface{}

		// Make sure to add value in right type,
		// because all subconfig are returned as strings from Proxmox API.
		if iValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			interValue = int(iValue)
		} else if bValue, err := strconv.ParseBool(value); err == nil {
			interValue = bValue
		} else {
			interValue = value
		}
		return key, interValue
	}
	return
}

// ParseConf - Parse standard device conf string `key1=val1,key2=val2`.
func ParseConf(
	kvString string,
	confSeparator string,
	subConfSeparator string,
	implicitFirstKey string,
) QemuDevice {
	var confMap = QemuDevice{}
	confList := strings.Split(kvString, confSeparator)

	if implicitFirstKey != "" {
		if !strings.Contains(confList[0], "=") {
			confMap[implicitFirstKey] = confList[0]
			confList = confList[1:]
		}
	}

	for _, item := range confList {
		key, value := ParseSubConf(item, subConfSeparator)
		confMap[key] = value
	}
	return confMap
}

func ParsePMConf(
	kvString string,
	implicitFirstKey string,
) QemuDevice {
	return ParseConf(kvString, ",", "=", implicitFirstKey)
}

// Convert a disk-size string to a GiB float
func DiskSizeGiB(dcSize interface{}) float64 {
	var diskSize float64
	switch dcSize := dcSize.(type) {
	case string:
		diskString := strings.ToUpper(dcSize)
		re := regexp.MustCompile("([0-9]+)([A-Z]*)")
		diskArray := re.FindStringSubmatch(diskString)

		diskSize, _ = strconv.ParseFloat(diskArray[1], 64)

		if len(diskArray) >= 3 {
			var diskSizeBytes float64
			switch diskArray[2] {
			// Convert IEC prefixed sizes to bytes
			case "T", "TiB":
				diskSizeBytes = diskSize * 1099511627776
			case "G", "GiB":
				diskSizeBytes = diskSize * 1073741824
			case "M", "MiB":
				diskSizeBytes = diskSize * 1048576
			case "K", "KiB":
				diskSizeBytes = diskSize * 1024
			// Convert SI prefixed sizes to bytes
			case "TB":
				diskSizeBytes = diskSize * 1000000000000
			case "GB":
				diskSizeBytes = diskSize * 1000000000
			case "MB":
				diskSizeBytes = diskSize * 1000000
			case "KB":
				diskSizeBytes = diskSize * 1000
			}
			// Convert bytes to IEC prefixed size (GiB)
			diskSize = diskSizeBytes / 1073741824
		}
	case float64:
		diskSize = dcSize
	}
	return diskSize
}

func AddToList(list, newItem string) string {
	if list != "" {
		return list + "," + newItem
	}
	return newItem
}

func CSVtoArray(csv string) []string {
	return strings.Split(csv, ",")
}

// Convert Array to a comma (,) delimited list
func ArrayToCSV(array interface{}) (csv string) {
	var arrayString []string
	switch array := array.(type) {
	case []interface{}:
		arrayString = ArrayToStringType(array)
	case []string:
		arrayString = array
	}
	csv = strings.Join(arrayString, `,`)
	return
}

// Convert Array of type []interface{} to array of type []string
func ArrayToStringType(inputarray []interface{}) (array []string) {
	array = make([]string, len(inputarray))
	for i, v := range inputarray {
		array[i] = v.(string)
	}
	return
}

// Creates a pointer to a string
func PointerString(text string) *string {
	return &text
}

// Creates a pointer to an int
func PointerInt(number int) *int {
	return &number
}

// Creates a pointer to a bool
func PointerBool(boolean bool) *bool {
	return &boolean
}

func failError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// Create list of http.Header out of string, separator is ","
func createHeaderList(header_string string, sess *Session) (*Session, error) {
	if header_string == "" {
		return sess, nil
	}
	header_string_split := strings.Split(header_string, ",")
	err := ValidateArrayEven(header_string_split, "Header key(s) and value(s) not even. Check your PM_HTTP_HEADERS env.")
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(header_string_split); i += 2 {
		sess.Headers[header_string_split[i]] = []string{header_string_split[i+1]}
	}
	return sess, nil
}
