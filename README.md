# wake-on-lan-proxy
Standard proxy server that provides some hacky server power management

## Rationale
Electricity is expensive, and I am lazy. This proxy will attempt to power on a server via Wake On LAN if it receives a request bound for the server, and wait for it to power-on before forwarding the request on and returning the response to the client. If the proxy goes a configurable length of time without receiving a request, it will then attempt to shut down the server to save on power usage.
