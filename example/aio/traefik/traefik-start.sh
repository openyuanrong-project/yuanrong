docker run -d --name traefik \
  -p 4443:4443 \
  -p 8888:8888 \
  -v $(pwd)/traefik.yml:/etc/traefik/traefik.yml \
  -v $(pwd)/../cert/:/openyuanrong/cert/ \
  traefik:v3.6