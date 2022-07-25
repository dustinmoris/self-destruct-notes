# self-destruct-notes

Go web service to create and read self destructing notes.

A public [Docker image is available on Docker Hub](https://hub.docker.com/repository/docker/dustinmoris/self-destruct-notes).

The quickest way to run this in your own environment is by running the following Docker compose:

```
version: "3.9"

services:
  db:
    image: redis:latest
    ports:
      - "6379:6379"
  web:
    image: dustinmoris/self-destruct-notes:1.0.0
    environment:
      - PORT=3000
      - REDIS_URL=redis://:@db:6379/1
    ports:
      - "3000:3000"
    depends_on:
      - db
```