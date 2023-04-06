# srcds_watch

Prometheus exporter for srcds stats retrieved via status/stats rcon command.


## Example

    docker run -v $(pwd)/srcds_watch.yml:/app/srcds_watch.yml -it leighmacdonald/srcds_watch:latest ls -lah