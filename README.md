# autoscale-docker-container

## Need

- docker-socket-proxy

## Example to use it

```yaml
services:
  whoami:
    image: traefik/whoami
    container_name: whoami
    networks:
      - web
    labels:
      - traefik.enable=false
    restart: "no"

  whoami-starter:
    image: docker-starter:latest
    container_name: whoami-starter
    environment:
      TARGET_CONTAINER: whoami
      TARGET_HOST: whoami
      TARGET_PORT: 80
      DOCKER_API: http://docker-socket-proxy:2375
      IDLE_TIMEOUT_SECONDS: 60
      CHECK_INTERVAL_SECONDS: 30
    networks:
      - web
    restart: unless-stopped
    labels:
      - traefik.enable=true
      - traefik.http.routers.whoami.rule=Host(`custom.domain`)
      - traefik.http.services.whoami.loadbalancer.server.port=8080

networks:
  web:
    external: true
```
