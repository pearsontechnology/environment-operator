package flux

import (
    "github.com/pearsontechnology/environment-operator/pkg/bitesize"
    "encoding/json"
    "encoding/base64"
    "fmt"
    "strings"
    "io/ioutil"
)

type ConsulValues []struct {
	Flags int    `json:"flags"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

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


func ConsulMerge(cvalues ConsulValues, envs *bitesize.EnvironmentsBitesize) (*bitesize.EnvironmentsBitesize, error){

    m := make(map[string]string)

    for _, env := range envs.Environments {
        for _, svc := range env.Services {

          if svc.Type == "" { // EO uses nil type for web apps
                                  svc.Type = "webservice"
             }


		for _, key := range cvalues {
		   //fmt.Println("Key", key.Key)
		   //fmt.Println("Value", key.Value)
		   skey := strings.Split(key.Key,"/")
		   if len(skey) == 2 {
			m["namespace"] = skey[0]
			m["service"] = "Any"
			m["key"] = skey[1]
			m["value"] = key.Value
			//fmt.Printf("%q\n",m)
		   }
		   if len(skey) > 2 {
			m["namespace"] = skey[0]
			m["service"] = skey[len(skey)-2]
			m["key"] = skey[len(skey)-1]
			m["value"] = key.Value
			//fmt.Printf("%q\n",m)
		   }

		    // fmt.Printf("%#v\n", env)
		   //if strings.TrimSpace(env.Namespace) == strings.TrimSpace(m["namespace"]) {
		        evalue:= bitesize.EnvVar{Name: m["key"], Value: m["value"]}
		        svc.EnvVars = append(svc.EnvVars, evalue)
		   if  m["namespace"] == string(env.Namespace) {
		     fmt.Printf("Namespace: %s %s \n", env.Namespace, m["namespace"])
		     if svc.Name == m["service"]{
		        evalue:= bitesize.EnvVar{Name: m["key"], Value: m["value"]}
		        svc.EnvVars = append(svc.EnvVars, evalue)
			//fmt.Printf("%q\n",svc.EnvVars)
			fmt.Printf("Inside check \n")
		     }
		  }


		   //fmt.Printf("%q\n",m)
		   //fmt.Println("Length : ", len(skey))
		   //m[key.Key] = "one"
		}

	          //fmt.Printf("%q\n",svc.EnvVars)
		  //key := fmt.Sprintf("%s-%s", env.Namespace, svc.Name)
		  //fmt.Printf("%#v \n", svc.Name)
		  //test2:= bitesize.EnvVar{Name: "test3"}
		  //svc.EnvVars = append(svc.EnvVars, test2)
		  //fmt.Printf("%#v \n", svc.EnvVars)
		  //val, err := RenderHelmRelease(env, svc, regPath)
		  //if err != nil {
		//	  panic(err)
        }
		  //m[key] = val
    }


        //fmt.Printf("%#v\n",envs)
        //fmt.Printf("%#v\n",svc.EnvVars)
	return envs,nil
}


