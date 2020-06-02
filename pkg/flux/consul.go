package flux

import (
    "encoding/json"
    "encoding/base64"
    "fmt"
    "io/ioutil"
)

// Struct definition for Consul export values
type ConsulValues []struct {
	Flags int    `json:"flags"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Read and parse the Consul input file
func ConsulRead(path string) (ConsulValues, error){

    // read our opened xmlFile as a byte array.
    var err error
    var contents []byte

    contents, err = ioutil.ReadFile(path)
    if err != nil {
       return nil, err
    }

    // we initialize our consul values array
    var cvalues ConsulValues

    // we unmarshal our byteArray 
    json.Unmarshal(contents, &cvalues)

    // decode the base64 encoded values
    for i := 0; i < len(cvalues); i++ {
        enc_string := cvalues[i].Value
	dec_string, err := base64.StdEncoding.DecodeString(enc_string)
	if err != nil {
		fmt.Println("error:", err)
	}
	cvalues[i].Value = string(dec_string)
    }
    return cvalues, nil

}


