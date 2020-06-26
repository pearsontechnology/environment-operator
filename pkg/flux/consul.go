package flux

import (
    "github.com/pearsontechnology/environment-operator/pkg/bitesize"
    "encoding/json"
    "encoding/base64"
    "fmt"
    "strings"
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

// RenderHelmReleasesWithConsul creates a map of serviceIdentifier:HelmRelease yaml
func RenderHelmReleasesWithConsul(envs *bitesize.EnvironmentsBitesize, regPath string, cvalues ConsulValues) map[string]string {
	m := make(map[string]string)
	cv := make(map[string]string)

	for _, env := range envs.Environments {
		for _, svc := range env.Services {

			if svc.Type == "" { // EO uses nil type for web apps
				svc.Type = "webservice"
			}
			key := fmt.Sprintf("%s", svc.Name)

			   for _, key := range cvalues {
			   skey := strings.Split(key.Key,"/")
			   if len(skey) == 2 { // for only Consul key value is defined without any hierarchy
				cv["namespace"] = skey[0]
				cv["service"] = "Any"
				cv["key"] = skey[1]
				cv["value"] = key.Value
			   }
			   if len(skey) > 2 {
				cv["namespace"] = skey[0]
				cv["service"] = skey[len(skey)-2]
				cv["key"] = skey[len(skey)-1]
				cv["value"] = key.Value
			   }
			   if len(skey) > 3 && skey[1] == "namespace" { // for GLP specific mapping
				cv["namespace"] = skey[2]
				cv["service"] = skey[3]
				cv["key"] = skey[len(skey)-1]
				cv["value"] = key.Value
			   }
			   if env.Namespace == cv["namespace"]  {
				// for where service_name used in Consul is defined as an env variable and is different to BS service name
				for _, v := range svc.EnvVars {
					if v.Name == "service_name" && v.Value == cv["service"]{
						evalue:= bitesize.EnvVar{Name: cv["key"], Value: cv["value"]}
						svc.EnvVars = append(svc.EnvVars, evalue)
					}
				}
				// for case where Consul values are defined at the top level ad key,value without any hierachy
				if cv["service"] == "Any" {
					evalue:= bitesize.EnvVar{Name: cv["key"], Value: cv["value"]}
					svc.EnvVars = append(svc.EnvVars, evalue)
				}
			   }
			}
			svc = AddConfigMap(svc)
			svc = ModifySecretKeys(svc)
			val, err := RenderHelmRelease(env, svc, regPath)
			if err != nil {
				panic(err)
			}
			m[key] = val
		}
	}
	return m
}

// Add ConfigMaps for specific larger environment values
func AddConfigMap(svc bitesize.Service) (bitesize.Service){
    for _, v := range svc.EnvVars {
        if v.Name == "application.properties" || v.Name == "application.yml" {
           var items []bitesize.KeyToPath           
           items = append(items, bitesize.KeyToPath{
                   	Key: v.Name,
                   	Path: v.Name,
                 	})
	   evolume := bitesize.Volume{
		Name: v.Name,
 		Type: "configmap",
		Path: "/etc/properties/",
		Items: items,           
		}
	svc.Volumes = append(svc.Volumes, evolume)
        }
    }
    return svc
}

func ModifySecretKeys(svc bitesize.Service) (bitesize.Service){
    for  i := range svc.EnvVars {
        if svc.EnvVars[i].Secret !="" {
    	   svc.EnvVars[i].Name = svc.EnvVars[i].Value
           if strings.Contains(svc.EnvVars[i].Value,"/") {
              svalue := strings.Split(svc.EnvVars[i].Value,"/")
	      svc.EnvVars[i].Name = svalue[0]
	      svc.EnvVars[i].Value = svalue[1]
	   }
        }
    }
    return svc
}

