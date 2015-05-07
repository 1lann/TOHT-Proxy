# TOHT-Proxy
TCP over HTTP proxy. Tunnels TCP traffic over the HTTP protocol to bypass corporate/school firewalls. To be used with SSH or OpenVPN to form a tunnel for all your traffic. [WIP NOT FULLY TESTED]

Proudly written in Go (It's cross platform!)

## What makes this tunnel unique?
It disguises your TCP traffic as HTTP (in a very hacky manner). This is to circumvent firewalls which strictly only
allow valid HTTP that is negotiated by the firewall. This can also be used as a proxy. You have the 
client running on your computer behind the firewall, and the server running on the server hosting the VPN/SSH services
(setting up a loopback), or any other web server that can then proxy it to your server hosting the services.

Ok that was probably a bad explanation, here's an (irl) example for clarity:
- I can't access my personal server because the whitelisting firewall at my school is blocking it.
- I can however access Heroku servers.
- I can set up the server software on Heroku.
- This will form a tunnel from my computer, to heroku, then to my personal server.

## How do I use it?
1. Install go on your client machine and server
2. On your client, download client.go and modify these lines near the top of the file

  ```go
  const proxyDomain = "127.0.0.1:9002"
  const port = "8080"
  ```
  
  to the domain and port of the server that will be running server.go, and the port you would like to use
  as the tunnel on the client.

3. On your server, download server.go and modify these lines near the top of the file

  ```go
  const port = "9002"
  const target = "123.44.32.45:22"
  ```

  to the port the tunnel will be running on (must match proxyDomain port specified in the client), and the target
  server's IP and port that you would like to channel TCP traffic to. Can be a lookback like `127.0.0.1:22`. This
  will usually be your IP:SSH port or OpenVPN port. Note that this proxy only supports connections to a single port,
  which is why you should use SSH or OpenVPN to tunnel all of your traffic.
4. Run client.go, and server.go
5. When setting up OpenVPN, change the TCP IP to 127.0.0.1:PORT
  OR when setting up SSH, use `ssh username@127.0.0.1 -p PORT` (PORT being defined under client.go's `const port`)

## Limitations
- Can only tunnel to 1 port on a remote server.
- May be super buggy.
- Not guranteed to work (I will try my best to obfuscate it!), also untested against
my school's network.

## If you are staff member from my school
I intend no harm and I will not mention this to other students. I just wanted to work on a fun challenge and
learn more about how networking works.
