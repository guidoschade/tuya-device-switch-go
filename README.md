# tuya-device-switch-go
Tuya cloud device switch and view command line application written in Golang

# switching on device with ID xxx (wifi power point)
go run tuya_switch.go -H https://openapi.tuyaeu.com -i ... -s ... -d ... -c switch_1 -v true

# switching off device with ID xxx (wifi power point)
go run tuya_switch.go -H https://openapi.tuyaeu.com -i ... -s ... -d ... -c switch_1 -v false

# viewing current device status (mode view)
go run tuya_switch.go -H https://openapi.tuyaeu.com -i ... -s ... -d ... -m view
