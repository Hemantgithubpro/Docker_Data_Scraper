# To check docker status
```bash
docker ps
```

# To stop the container, not delete it
```bash
docker compose stop
```

# To stop and delete the container including the data/volumes
```bash
docker compose down -v
```

# To check the logs of the container
```bash
docker compose logs -f
```

# To start the container (in detached mode)
```bash
docker compose up -d
```