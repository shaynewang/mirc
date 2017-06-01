# Min Internet Relay Chat

A very lightweight single server IRC application

![pic](https://github.com/shaynewang/mirc/blob/master/example.png)

* build docker image 
    ``` docker build -t mirc . ```
* run server container
    ``` ./run.sh ```

To specify server ip address please do so in ```config.yaml```. The default port number for MIRC is 6667.

* run client
    ``` make ```
    ``` ./bin/client ```
