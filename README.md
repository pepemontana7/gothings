#GoThings OAUTH Provider

The following is a barebones oauth provider, that can be used in a raspberry pi for integration with Smarthings ecosystem.

## Getting Started

These instructions will get you a copy of the project up and running on your raspberry pi v3 for development and testing purposes. More work needs to be done to get it to a proper production ready oauth provider. Currently, the steps for setup are manual but future setup script or container is planned via iofog.

### Prerequisites

What things you need to install the software and how to install them

```
Raspbery pi 3 with raspian
go
ngrok 
docker
```

The below are required if you plan to use a device that uses BLE
```
    iofog (bluetooth rest api)
    Arduino with Adafruit bluefruit LE shield and a sensor
```



### Installing

Below you will find guides to get the required setup (future will be in a container or setup script):

Folow steps to set up Raspbian [Raspian Setup](https://www.raspberrypi.org/documentation/installation/installing-images/).
Folow steps to set up go [GO Setup](https://golang.org/doc/install). 
Folow steps to set up ngrok [Ngrok Setup](https://ngrok.com/download). 
Folow steps to set up iofog [Iofog Setup](https://iotracks.com/products/iofog/installation/linux/raspbian).


After required packages are installed you will need to setup an ngrok tunnel:
Note: we are using port 14000 as the go server is using 14000 to start http server 
```
./ngrok http 14000
```

After tunnel is setup you can retrieve it from stdout if you did not run  ngrok in the background or using ui or curl call:

```
UI : http://127.0.0.1:4040
curl:   curl http://localhost:4040/api/tunnels
```

Your public url will look similar to  d7ttt3d.ngrok.io which you can use when starting the go server below:
Note: substitue YOUR_GO_PATH to the correct directory in your env
```
go get pepemontana7/gothings
cd YOUR_GO_PATH/src/github.com/pepemontana7/gothings/
./gothings --endpoint https://dd7ttt3d.ngrok.io --redirect https://graph.api.smartthings.com/oauth/callback &
```


The Endpoint should be availble now via https://dd7ttt3d.ngrok.io. It can be used in your smartapp oauth configuration.
The default client uses the creds below( Note: these can be overridden whiile starting gothings server):
```
client_id: 1234
client_secret: aabbccdd
user_login: test
pwd_login: test
```

### Resources
```
https://github.com/RangelReale/osin
https://github.com/StrykerSKS/Ecobee
http://docs.smartthings.com/en/latest/cloud-and-lan-connected-device-types-developers-guide/index.html?highlight=ecobee

```