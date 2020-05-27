USSD WhoIs
===

A simple http server that provides Domain WhoIs check via a USSD application.
It uses [https://jsonwhois.io](https://jsonwhois.io) to perform the WhoIs checks and
is developed to run on [Africastalking's](https://build.at-labs.io/docs/ussd%2Foverview) USSD service.

It is developed using Go and requires Go to build.

## Building

In order to build this, you will need to have Go installed, atleast Go 1.12.

### Specifying the JSON Whois.io API Key

As previously mentioned, the application uses [jsonwhois.io](https://jsonwhois.io) for
performing the Who Is checks, so you must be signed up and have an API Key for that platform
before building the binary. The binary must be built with a valid, correct API Key.

```sh
$ go build -ldApiKey="__YOUR_API_KEY__"
```

## Running

Once you have built it, you can run it as follows:

```sh
$ ./ussd-whois -b "your-ip-address:8773"
```

Once you have it running on a server accessible to the internet, you can connect
it to the Africastalking platform by setting the `USSD Callback URL`

You can also use [ngrok](https://github.com/inconshreveable/ngrok) to bridge a 
running instance on your local dev machine to the internet and set the ngrok url as the `USSD Callback URL`

And finally, you should definitely use [dialoguss](https://github.com/nndi-oss/dialoguss)
to test this and other USSD applications. 

## CONTRIBUTING

Pull Requests are welcome. 

---

Copyright (c) 2020, NNDI