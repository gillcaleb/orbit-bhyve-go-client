package main

import (
    "fmt"

    "github.com/spf13/viper"
    "github.com/gillcaleb/orbit-bhyve-go-client/pkg/client"
)

func main() {

    viper.SetConfigName("config")
    viper.AddConfigPath(".")
    viper.AutomaticEnv()
    viper.SetConfigType("yaml")
    
    if err := viper.ReadInConfig(); err != nil {
        fmt.Printf("Error reading config file, %s", err)
    }

    config := client.Config{
        Endpoint: "https://api.orbitbhyve.com/v1",
        Email: viper.GetString("Email"),
        Password: viper.GetString("Password"),
        DeviceId: viper.GetString("DeviceId"),
    }

    c := client.NewClient(config)
    err := c.Init()
    if err != nil {
        fmt.Println("Error initializing client: %v", err)
    }
}
