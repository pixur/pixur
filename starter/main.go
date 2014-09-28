package main 

import (
  "log"
  "flag"
  "os"
  "encoding/json"

  "pixur.org/pixur"
)

var (
  config = flag.String("config", ".config.json", "The default configuration file")
  mysqlConfig = flag.String("mysql_config", "", "The default mysql config")
  spec = flag.String("spec", ":8888", "Default HTTP port")
)


func getConfig(path string) (*pixur.Config, error) {
  var config = new(pixur.Config)
  f, err := os.Open(path)

  if os.IsNotExist(err) {
    log.Println("Unable to open config file, using defaults")
    return config, nil
  } else if err != nil {
    return nil, err
  }
  defer f.Close()
  
  configDecoder := json.NewDecoder(f)
  if err := configDecoder.Decode(config); err != nil {
    return nil, err
  }
  
  return config, nil
}


func main() {
  flag.Parse()

  c, err := getConfig(*config)
  if err != nil {
    log.Fatal(err)
  }
  if *mysqlConfig != "" {
    c.MysqlConfig = *mysqlConfig
  }
  c.HttpSpec = *spec
  
  s := &pixur.Server{}
  
  log.Fatal(s.StartAndWait(c))
}