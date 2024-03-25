# Tempus

## Docker
### Build image
```console
docker build -t tempus -f Containerfile .
```

### Run container
```console
docker run \
    -e TEMPUS_TOKEN="TOKEN" \
    -e TEMPUS_GUILD_ID="GUILD_ID" \
    -v $(pwd)/data:/bot/data \
    tempus
```
