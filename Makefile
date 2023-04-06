docker_run:
	docker build -t leighmacdonald/srcds_watch:latest .
	docker run -v $(pwd)/srcds_watch.yml:/app/srcds_watch.yml -it leighmacdonald/srcds_watch:latest