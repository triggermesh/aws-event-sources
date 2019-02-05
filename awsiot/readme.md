# Setup 

install python 3 & pip3

run ```pip3 install AWSIoTPythonSDK```

To get needed files follow the AWS official [tutorial](https://docs.aws.amazon.com/iot/latest/developerguide/iot-gs.html)

copy the following files inside your project's folder:
- Root CA
- Certificate for the Thing
- Private Key 

copy Your AWS IoT custom endpoint:
1. Open up your Thing
2. Look for ```interact``` item on the left pannel
3. Inside ```interact``` copy HTTPS endpoint for your Thing 

Update your ENV variables and run 
``` python3 main.py ```

# Docker
Run

```docker build . -t awsiot:latest```

```docker run -e THING_SHADOW_ENDPOINT='a1cdqjmv85rx59-ats.iot.us-west-2.amazonaws.com' -e ROOT_CA_PATH='rootca.crt' -e CERTIFICATE_PATH='certificate.pem.crt' -e PRIVATE_KEY_PATH='private.pem.key' -e TOPIC='triggertopic' -ti awsiot ```