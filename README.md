# Swarm tools

## Environment variables

```
PORT - default is 80
WHITELIST - eg. "dev_api,dev_web,helloworld"
AUTH_KEY - Bearer token, eg. "XYZ"
AUTH_KEY_FILE - eg. /run/secrets/swarm_tools_auth_key
```

## Swarm service example definition

```
services:
    swarm-tools:
        image: pajk/swarm-tools:0.2.0
        environment:
            AUTH_KEY: XYZ
            WHITELIST: helloworld,dev_api,dev_web
            PORT: 80
        ports:
            - 2380:80
        volumes:
            - /var/run/docker.sock:/var/run/docker.sock
        deploy:
            placement:
                constraints: [node.role == manager]
```

## Update service image

```
curl -X POST -H "Authorization: Bearer XYZ" "http://swarm:2380/services/update?name=helloworld&image=tutum/hello-world"
```

## List services

```
curl -H "Authorization: Bearer XYZ" "http://swarm:2380/services"

Example output:
id: "3gk2bvrjy0g8a8eduuh0wb3lw", name: "helloworld", image: "tutum/hello-world:latest", version: 9513
```