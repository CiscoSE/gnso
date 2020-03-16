# NSO Go Controller

Simple app that provides NSO with a grpc interface. 

## Supported methods
 
* GetDevices
* EditConfig 
* GetConfig
* Query
* ExecOperation 

_Full proto definition can be found at proto/service.proto_

## NSO version and requirements

NSO 5.2 and greater are supported. RESTCONF API needs to be enabled

## Installation

You can run this in any linux with Go 1.14. The code can also be run in a container. The app will read the following env variables:

* NSO_USERNAME -> For example: admin
* NSO_PASSWORD -> For example: admin
* NSO_URL -> Needs to point to the RESTCONF URL. For example: http://localhost:8080/restconf

In adition to above, you need to generate cert.pem and key.pem files under tls directory. For example:

```bash

openssl req -newkey rsa:2048 -new -nodes -x509 -days 3650 -keyout key.pem -out cert.pem
```


## Contacts

* Santiago Flores Kante (sfloresk@cisco.com)

## License

Provided under Cisco Sample Code License, for details see [LICENSE](./LICENSE.md)

## Code of Conduct

Our code of conduct is available [here](./CODE_OF_CONDUCT.md)

## Contributing

See our contributing guidelines [here](./CONTRIBUTING.md)
